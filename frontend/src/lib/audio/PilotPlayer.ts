import { pilotAudioUrl } from '$lib/api/pilots';

interface PilotState {
	id: number;
	trackId: number;
	name: string;
	buffer: AudioBuffer;
	gain: GainNode;
	source: AudioBufferSourceNode | null;
	muted: boolean;
}

/**
 * Manages playback of one or more pilot audio tracks via the Web Audio API.
 * All pilots are started simultaneously and kept in sync with the MIDI recorder
 * by anchoring both to the same AudioContext time.
 */
export class PilotPlayer {
	private ctx: AudioContext;
	private pilots: PilotState[] = [];
	private startContextTime = 0;
	private playbackRate = 1;
	private playing = false;

	constructor(ctx: AudioContext) {
		this.ctx = ctx;
	}

	/** Load audio for a set of pilots. Call before startAll(). */
	async load(pilots: { id: number; trackId: number; name: string }[]): Promise<void> {
		const loaded = await Promise.all(
			pilots.map(async (p) => {
				const url = pilotAudioUrl(p.trackId, p.id);
				const resp = await fetch(url);
				if (!resp.ok) throw new Error(`Failed to fetch pilot audio: ${resp.status}`);
				const arrayBuf = await resp.arrayBuffer();
				const buffer = await this.ctx.decodeAudioData(arrayBuf);
				const gain = this.ctx.createGain();
				gain.gain.value = 1;
				gain.connect(this.ctx.destination);
				return { id: p.id, trackId: p.trackId, name: p.name, buffer, gain, source: null, muted: false };
			})
		);
		this.pilots = loaded;
	}

	/**
	 * Start all pilots at the given AudioContext time (defaults to now).
	 * Pass `contextTime` = the same value you anchor your MIDI clock to.
	 */
	startAll(contextTime?: number, playbackRate = 1): void {
		if (this.playing) this.stopAll();
		this.playbackRate = playbackRate;
		this.startContextTime = contextTime ?? this.ctx.currentTime;

		for (const p of this.pilots) {
			const src = this.ctx.createBufferSource();
			src.buffer = p.buffer;
			src.playbackRate.value = this.playbackRate;
			src.connect(p.gain);
			src.start(this.startContextTime);
			p.source = src;
		}
		this.playing = true;
	}

	stopAll(): void {
		for (const p of this.pilots) {
			try { p.source?.stop(); } catch { /* already stopped */ }
			p.source = null;
		}
		this.playing = false;
	}

	/** Update playback rate on all sources (e.g. when BPM changes). */
	setPlaybackRate(rate: number): void {
		this.playbackRate = rate;
		for (const p of this.pilots) {
			if (p.source) p.source.playbackRate.value = rate;
		}
	}

	setMuted(pilotId: number, muted: boolean): void {
		const p = this.pilots.find((x) => x.id === pilotId);
		if (!p) return;
		p.muted = muted;
		const target = muted ? 0 : 1;
		// 10ms ramp to avoid click artifacts.
		p.gain.gain.linearRampToValueAtTime(target, this.ctx.currentTime + 0.01);
	}

	isMuted(pilotId: number): boolean {
		return this.pilots.find((x) => x.id === pilotId)?.muted ?? false;
	}

	isPlaying(): boolean {
		return this.playing;
	}

	destroy(): void {
		this.stopAll();
		for (const p of this.pilots) {
			p.gain.disconnect();
		}
		this.pilots = [];
	}
}
