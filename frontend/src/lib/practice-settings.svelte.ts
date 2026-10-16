// Practice preferences, remembered in this browser.

export type StartMode = 'count-in' | 'first-hit';

interface Stored {
	startMode: StartMode;
	clickDuringTake: boolean;
	autoStop: boolean;
	/** Measured input latency (ms) per MIDI device id. */
	latency: Record<string, number>;
}

const STORAGE_KEY = 'drum-trainer:practice-settings';

const defaults: Stored = {
	startMode: 'count-in',
	clickDuringTake: true,
	autoStop: true,
	latency: {}
};

function load(): Stored {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) {
			const parsed = JSON.parse(raw) as Partial<Stored>;
			return {
				startMode: parsed.startMode === 'first-hit' ? 'first-hit' : 'count-in',
				clickDuringTake: parsed.clickDuringTake ?? defaults.clickDuringTake,
				autoStop: parsed.autoStop ?? defaults.autoStop,
				latency: typeof parsed.latency === 'object' && parsed.latency ? parsed.latency : {}
			};
		}
	} catch {
		// Unavailable or corrupt storage — fall back to defaults.
	}
	return { ...defaults, latency: {} };
}

const state = $state<Stored>(typeof window !== 'undefined' ? load() : { ...defaults, latency: {} });

function save() {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	} catch {
		// Not persisted, but still applied for this visit.
	}
}

export const practiceSettings = {
	get startMode() {
		return state.startMode;
	},
	set startMode(v: StartMode) {
		state.startMode = v;
		save();
	},
	get clickDuringTake() {
		return state.clickDuringTake;
	},
	set clickDuringTake(v: boolean) {
		state.clickDuringTake = v;
		save();
	},
	get autoStop() {
		return state.autoStop;
	},
	set autoStop(v: boolean) {
		state.autoStop = v;
		save();
	},
	/** Measured latency for a device, or null if it has not been calibrated. */
	latencyFor(deviceId: string): number | null {
		return deviceId in state.latency ? state.latency[deviceId] : null;
	},
	setLatency(deviceId: string, ms: number | null) {
		if (ms === null) delete state.latency[deviceId];
		else state.latency[deviceId] = ms;
		save();
	}
};
