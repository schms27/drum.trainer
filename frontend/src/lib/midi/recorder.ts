import type { RecordedHit } from '$lib/types/session';

export interface MidiDevice {
	id: string;
	name: string;
}

/** List all available MIDI input devices. */
export async function listMidiInputs(): Promise<MidiDevice[]> {
	if (!navigator.requestMIDIAccess) return [];
	try {
		const access = await navigator.requestMIDIAccess({ sysex: false });
		const inputs: MidiDevice[] = [];
		access.inputs.forEach((input) => {
			inputs.push({ id: input.id, name: input.name ?? input.id });
		});
		return inputs;
	} catch {
		return [];
	}
}

export type RecorderState = 'idle' | 'waiting' | 'recording' | 'stopped';

export interface Recorder {
	state: RecorderState;
	hits: RecordedHit[];
	/** Arms the recorder: listens for MIDI but only starts the clock on the first hit. */
	arm: () => void;
	stop: () => RecordedHit[];
	reset: () => void;
	destroy: () => void;
}

export interface RecorderOptions {
	/** Called the moment the first hit is received and the clock starts. */
	onStarted?: () => void;
}

/**
 * Creates a recorder that listens to a specific MIDI input.
 * The caller is responsible for calling destroy() when done.
 */
export async function createRecorder(
	deviceId: string,
	options?: RecorderOptions
): Promise<Recorder> {
	const access = await navigator.requestMIDIAccess({ sysex: false });
	const input = access.inputs.get(deviceId);

	if (!input) throw new Error(`MIDI input ${deviceId} not found`);

	let state: RecorderState = 'idle';
	let hits: RecordedHit[] = [];
	let startTime = 0;

	const onMessage = (event: MIDIMessageEvent) => {
		if (state !== 'waiting' && state !== 'recording') return;

		const data = event.data;
		if (!data || data.length < 3) return;

		const status = data[0];
		const note = data[1];
		const velocity = data[2];

		// NoteOn: status high nibble = 0x9, velocity > 0
		const isNoteOn = (status & 0xf0) === 0x90 && velocity > 0;
		if (!isNoteOn) return;

		// First hit while waiting — start the clock now.
		if (state === 'waiting') {
			startTime = event.timeStamp;
			state = 'recording';
			options?.onStarted?.();
		}

		const timeMs = event.timeStamp - startTime;
		hits.push({ note_number: note, time_ms: timeMs, velocity });
	};

	input.addEventListener('midimessage', onMessage as EventListener);

	return {
		get state() {
			return state;
		},
		get hits() {
			return [...hits];
		},
		arm() {
			hits = [];
			state = 'waiting';
		},
		stop() {
			state = 'stopped';
			return [...hits];
		},
		reset() {
			hits = [];
			state = 'idle';
		},
		destroy() {
			input.removeEventListener('midimessage', onMessage as EventListener);
		}
	};
}
