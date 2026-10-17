import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createRecorder } from '$lib/midi/recorder';
import { resetMidiAccessForTests } from '$lib/midi/input';

// --- Mock Web MIDI API ---

interface MockInput {
	addEventListener: ReturnType<typeof vi.fn>;
	removeEventListener: ReturnType<typeof vi.fn>;
	/** Simulate an incoming MIDI message. */
	dispatch: (data: Uint8Array, timeStamp: number) => void;
}

function createMockInput(): MockInput {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let captured: ((event: any) => void) | undefined;
	return {
		addEventListener: vi.fn((_type: string, listener: (e: unknown) => void) => {
			captured = listener;
		}),
		removeEventListener: vi.fn(),
		dispatch(data: Uint8Array, timeStamp: number) {
			captured?.({ data, timeStamp });
		},
	};
}

function stubMidiAccess(input: MockInput, deviceId = 'dev-1') {
	vi.stubGlobal('navigator', {
		requestMIDIAccess: vi.fn().mockResolvedValue({
			inputs: { get: (id: string) => (id === deviceId ? input : undefined) },
		}),
	});
}

/** Build a NoteOn MIDI status byte for channel `ch` (0-indexed). */
function noteOnByte(ch: number) {
	return 0x90 | ch;
}

function noteOnMsg(note: number, velocity: number, ch = 9): Uint8Array {
	return new Uint8Array([noteOnByte(ch), note, velocity]);
}

// ---

describe('createRecorder', () => {
	let input: MockInput;

	beforeEach(() => {
		vi.resetAllMocks();
		resetMidiAccessForTests();
		input = createMockInput();
		stubMidiAccess(input);
	});

	it('starts in idle state', async () => {
		const rec = await createRecorder('dev-1');
		expect(rec.state).toBe('idle');
		expect(rec.hits).toHaveLength(0);
	});

	it('transitions to waiting after arm()', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		expect(rec.state).toBe('waiting');
	});

	it('transitions to recording on first MIDI hit', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		expect(rec.state).toBe('recording');
	});

	it('calls onStarted exactly once when recording begins', async () => {
		const onStarted = vi.fn();
		const rec = await createRecorder('dev-1', { onStarted });
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		input.dispatch(noteOnMsg(38, 90), 1200);
		expect(onStarted).toHaveBeenCalledTimes(1);
	});

	it('records the first hit with time_ms = 0', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 5000);
		expect(rec.hits[0]).toMatchObject({ note_number: 36, time_ms: 0, velocity: 100 });
	});

	it('offsets all hits by firstHitTimeMs', async () => {
		const rec = await createRecorder('dev-1', { firstHitTimeMs: 2000 });
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		input.dispatch(noteOnMsg(38, 90), 1250);
		expect(rec.hits.map((h) => h.time_ms)).toEqual([2000, 2250]);
	});

	it('records subsequent hits with relative timestamps', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000); // clock starts at 1000
		input.dispatch(noteOnMsg(38, 90), 1250);  // +250ms
		input.dispatch(noteOnMsg(42, 80), 1500);  // +500ms

		const hits = rec.hits;
		expect(hits).toHaveLength(3);
		expect(hits[0].time_ms).toBe(0);
		expect(hits[1].time_ms).toBe(250);
		expect(hits[2].time_ms).toBe(500);
	});

	it('ignores messages while idle (before arm)', async () => {
		const rec = await createRecorder('dev-1');
		input.dispatch(noteOnMsg(36, 100), 1000);
		expect(rec.hits).toHaveLength(0);
		expect(rec.state).toBe('idle');
	});

	it('ignores NoteOn with velocity 0 (NoteOff convention)', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000); // real hit — starts clock
		input.dispatch(new Uint8Array([0x90, 36, 0]), 1100); // vel=0 → NoteOff
		expect(rec.hits).toHaveLength(1);
	});

	it('ignores non-NoteOn status bytes', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000); // start clock
		input.dispatch(new Uint8Array([0x80, 36, 64]), 1200); // NoteOff (0x8x)
		input.dispatch(new Uint8Array([0xb0, 7, 100]), 1300); // CC
		expect(rec.hits).toHaveLength(1);
	});

	it('stop() sets state to stopped and returns captured hits', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		const hits = rec.stop();
		expect(rec.state).toBe('stopped');
		expect(hits).toHaveLength(1);
	});

	it('ignores hits after stop()', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		rec.stop();
		input.dispatch(noteOnMsg(38, 90), 1200);
		expect(rec.hits).toHaveLength(1);
	});

	it('reset() clears hits and returns to idle', async () => {
		const rec = await createRecorder('dev-1');
		rec.arm();
		input.dispatch(noteOnMsg(36, 100), 1000);
		rec.reset();
		expect(rec.state).toBe('idle');
		expect(rec.hits).toHaveLength(0);
	});

	it('destroy() removes the MIDI event listener', async () => {
		const rec = await createRecorder('dev-1');
		rec.destroy();
		expect(input.removeEventListener).toHaveBeenCalledWith(
			'midimessage',
			expect.any(Function),
		);
	});

	it('armAt() records immediately, relative to the given start time', async () => {
		const rec = await createRecorder('dev-1');
		rec.armAt(5000);
		expect(rec.state).toBe('recording');
		input.dispatch(noteOnMsg(36, 100), 5010);
		input.dispatch(noteOnMsg(38, 90), 5510);
		expect(rec.hits.map((h) => h.time_ms)).toEqual([10, 510]);
	});

	it('armAt() subtracts the latency correction', async () => {
		const rec = await createRecorder('dev-1');
		rec.armAt(5000, 25);
		input.dispatch(noteOnMsg(36, 100), 5025);
		expect(rec.hits[0].time_ms).toBe(0);
	});

	it('armAt() keeps slightly early hits but drops count-in noise', async () => {
		const rec = await createRecorder('dev-1');
		rec.armAt(5000);
		input.dispatch(noteOnMsg(42, 80), 3000); // during the count-in
		input.dispatch(noteOnMsg(36, 100), 4980); // 20 ms early for beat 1
		expect(rec.hits.map((h) => h.time_ms)).toEqual([-20]);
	});

	it('setStartTime() moves time zero and shifts recorded hits', async () => {
		const rec = await createRecorder('dev-1');
		rec.armAt(5000);
		input.dispatch(noteOnMsg(36, 100), 5100);
		rec.setStartTime(5060);
		input.dispatch(noteOnMsg(38, 90), 5560);
		expect(rec.hits.map((h) => h.time_ms)).toEqual([40, 500]);
	});

	it('throws when the device id is not found', async () => {
		await expect(createRecorder('no-such-device')).rejects.toThrow('not found');
	});
});
