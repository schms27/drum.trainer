// Shared Web MIDI access: one MIDIAccess for the whole app, plus helpers to
// listen to a device and to react when devices are plugged in or removed.

let accessPromise: Promise<MIDIAccess> | null = null;

export function midiSupported(): boolean {
	return typeof navigator !== 'undefined' && typeof navigator.requestMIDIAccess === 'function';
}

export function getMidiAccess(): Promise<MIDIAccess> {
	if (!accessPromise) {
		accessPromise = navigator.requestMIDIAccess({ sysex: false }).catch((e) => {
			accessPromise = null; // allow a retry, e.g. after the user grants permission
			throw e;
		});
	}
	return accessPromise;
}

export interface NoteOn {
	note: number;
	velocity: number;
	/** performance.now() clock, ms. */
	timeStamp: number;
}

/** Returns the NoteOn in a MIDI message, or null for anything else. */
export function parseNoteOn(data: Uint8Array | null | undefined): Omit<NoteOn, 'timeStamp'> | null {
	if (!data || data.length < 3) return null;
	const isNoteOn = (data[0] & 0xf0) === 0x90 && data[2] > 0;
	return isNoteOn ? { note: data[1], velocity: data[2] } : null;
}

/**
 * Calls onNote for every NoteOn from the given input. Resolves to a function
 * that stops listening.
 */
export async function listenToInput(deviceId: string, onNote: (hit: NoteOn) => void): Promise<() => void> {
	const access = await getMidiAccess();
	const input = access.inputs.get(deviceId);
	if (!input) throw new Error(`MIDI input ${deviceId} not found`);
	const handler = (event: Event) => {
		const e = event as MIDIMessageEvent;
		const parsed = parseNoteOn(e.data);
		if (parsed) onNote({ ...parsed, timeStamp: e.timeStamp });
	};
	input.addEventListener('midimessage', handler);
	return () => input.removeEventListener('midimessage', handler);
}

/** Calls onChange whenever a MIDI device is connected or disconnected. */
export async function watchDevices(onChange: () => void): Promise<() => void> {
	const access = await getMidiAccess();
	const handler = () => onChange();
	access.addEventListener('statechange', handler);
	return () => access.removeEventListener('statechange', handler);
}

/** Forget the cached MIDIAccess (tests only). */
export function resetMidiAccessForTests(): void {
	accessPromise = null;
}
