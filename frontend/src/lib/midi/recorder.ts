import type { RecordedHit } from '$lib/types/session';
import { getMidiAccess, midiSupported, parseNoteOn } from './input';

export interface MidiDevice {
	id: string;
	name: string;
}

/** List all available MIDI input devices. */
export async function listMidiInputs(): Promise<MidiDevice[]> {
	if (!midiSupported()) return [];
	try {
		const access = await getMidiAccess();
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

/**
 * With a count-in, hits up to this long before beat 1 are kept (a slightly
 * early first note); anything earlier is noise from the count-in bar.
 */
export const EARLY_HIT_LIMIT_MS = 300;

export interface Recorder {
	state: RecorderState;
	hits: RecordedHit[];
	/** Arms the recorder: listens for MIDI but only starts the clock on the first hit. */
	arm: () => void;
	/**
	 * Starts recording with a fixed time zero (performance.now() clock), e.g.
	 * the downbeat after a count-in. `latencyMs` is subtracted from every hit to
	 * correct for the kit's measured input delay.
	 */
	armAt: (startTimeStamp: number, latencyMs?: number) => void;
	/**
	 * Moves time zero (armAt mode) to a more accurate value, shifting hits that
	 * were already recorded. Used once the audio output clock is known.
	 */
	setStartTime: (startTimeStamp: number) => void;
	stop: () => RecordedHit[];
	reset: () => void;
	destroy: () => void;
}

export interface RecorderOptions {
	/** Called the moment the first hit is received and the clock starts (arm() only). */
	onStarted?: () => void;
	/**
	 * Timestamp (ms) assigned to the first hit with arm(). Set this to the time
	 * of the first target note so recorded hits share the target track's time
	 * base even when the MIDI file has a lead-in before the first note.
	 */
	firstHitTimeMs?: number;
}

/**
 * Creates a recorder that listens to a specific MIDI input.
 * The caller is responsible for calling destroy() when done.
 */
export async function createRecorder(deviceId: string, options?: RecorderOptions): Promise<Recorder> {
	const access = await getMidiAccess();
	const input = access.inputs.get(deviceId);

	if (!input) throw new Error(`MIDI input ${deviceId} not found`);

	let state: RecorderState = 'idle';
	let hits: RecordedHit[] = [];
	let startTime = 0;
	let latency = 0;
	let fixedStart = false;

	const onMessage = (event: MIDIMessageEvent) => {
		if (state !== 'waiting' && state !== 'recording') return;

		const noteOn = parseNoteOn(event.data);
		if (!noteOn) return;

		// First hit while waiting — start the clock now.
		if (state === 'waiting') {
			startTime = event.timeStamp - (options?.firstHitTimeMs ?? 0);
			state = 'recording';
			options?.onStarted?.();
		}

		const timeMs = event.timeStamp - latency - startTime;
		if (fixedStart && timeMs < -EARLY_HIT_LIMIT_MS) return;
		hits.push({ note_number: noteOn.note, time_ms: timeMs, velocity: noteOn.velocity });
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
			latency = 0;
			fixedStart = false;
			state = 'waiting';
		},
		armAt(startTimeStamp: number, latencyMs = 0) {
			hits = [];
			startTime = startTimeStamp;
			latency = latencyMs;
			fixedStart = true;
			state = 'recording';
		},
		setStartTime(startTimeStamp: number) {
			const delta = startTime - startTimeStamp;
			startTime = startTimeStamp;
			hits = hits.map((h) => ({ ...h, time_ms: h.time_ms + delta }));
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
