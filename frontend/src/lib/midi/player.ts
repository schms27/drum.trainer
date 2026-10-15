import { drumGroup } from '$lib/drums';

export interface PlayableHit {
	note_number: number;
	time_ms: number;
	velocity: number;
}

/** 500 ms of white noise shared across all scheduled hits in one playback. */
function makeNoiseBuffer(ctx: AudioContext): AudioBuffer {
	const len = Math.ceil(ctx.sampleRate * 0.5);
	const buf = ctx.createBuffer(1, len, ctx.sampleRate);
	const data = buf.getChannelData(0);
	for (let i = 0; i < len; i++) data[i] = Math.random() * 2 - 1;
	return buf;
}

function scheduleHit(
	ctx: AudioContext,
	dest: AudioNode,
	noise: AudioBuffer,
	note: number,
	velocity: number,
	t: number
): void {
	const amp = (velocity / 127) * 0.8;

	// All tom pitches share one synth voice.
	const group = drumGroup(note);
	switch (group.startsWith('Tom') ? 'Tom' : group) {
		case 'Bass Drum': {
			const osc = ctx.createOscillator();
			const env = ctx.createGain();
			osc.type = 'sine';
			osc.frequency.setValueAtTime(80, t);
			osc.frequency.exponentialRampToValueAtTime(35, t + 0.1);
			env.gain.setValueAtTime(amp, t);
			env.gain.exponentialRampToValueAtTime(0.001, t + 0.3);
			osc.connect(env);
			env.connect(dest);
			osc.start(t);
			osc.stop(t + 0.35);
			break;
		}
		case 'Snare': {
			// Noise layer
			const ns = ctx.createBufferSource();
			ns.buffer = noise;
			const nf = ctx.createBiquadFilter();
			nf.type = 'bandpass';
			nf.frequency.value = 3000;
			nf.Q.value = 0.5;
			const ng = ctx.createGain();
			ng.gain.setValueAtTime(amp * 0.7, t);
			ng.gain.exponentialRampToValueAtTime(0.001, t + 0.15);
			ns.connect(nf);
			nf.connect(ng);
			ng.connect(dest);
			ns.start(t);
			ns.stop(t + 0.15);
			// Tone layer (crack)
			const osc = ctx.createOscillator();
			const og = ctx.createGain();
			osc.type = 'triangle';
			osc.frequency.value = 200;
			og.gain.setValueAtTime(amp * 0.5, t);
			og.gain.exponentialRampToValueAtTime(0.001, t + 0.08);
			osc.connect(og);
			og.connect(dest);
			osc.start(t);
			osc.stop(t + 0.1);
			break;
		}
		case 'Hi-Hat': {
			const ns = ctx.createBufferSource();
			ns.buffer = noise;
			const f = ctx.createBiquadFilter();
			f.type = 'highpass';
			f.frequency.value = 7000;
			const decay = note === 46 ? 0.25 : 0.06; // open vs closed/pedal
			const g = ctx.createGain();
			g.gain.setValueAtTime(amp * 0.5, t);
			g.gain.exponentialRampToValueAtTime(0.001, t + decay);
			ns.connect(f);
			f.connect(g);
			g.connect(dest);
			ns.start(t);
			ns.stop(t + decay + 0.01);
			break;
		}
		case 'Crash': {
			const ns = ctx.createBufferSource();
			ns.buffer = noise;
			const f = ctx.createBiquadFilter();
			f.type = 'highpass';
			f.frequency.value = 4000;
			const g = ctx.createGain();
			g.gain.setValueAtTime(amp * 0.6, t);
			g.gain.exponentialRampToValueAtTime(0.001, t + 0.5);
			ns.connect(f);
			f.connect(g);
			g.connect(dest);
			ns.start(t);
			ns.stop(t + 0.5);
			break;
		}
		case 'Ride': {
			const osc = ctx.createOscillator();
			const g = ctx.createGain();
			osc.type = 'sine';
			osc.frequency.value = 800;
			g.gain.setValueAtTime(amp * 0.35, t);
			g.gain.exponentialRampToValueAtTime(0.001, t + 0.3);
			osc.connect(g);
			g.connect(dest);
			osc.start(t);
			osc.stop(t + 0.35);
			break;
		}
		case 'Tom': {
			const osc = ctx.createOscillator();
			const g = ctx.createGain();
			osc.type = 'sine';
			// Lower frequency for floor toms, higher for rack toms
			const freq = note <= 43 ? 80 : note <= 47 ? 110 : 150;
			osc.frequency.setValueAtTime(freq * 1.6, t);
			osc.frequency.exponentialRampToValueAtTime(freq, t + 0.1);
			g.gain.setValueAtTime(amp, t);
			g.gain.exponentialRampToValueAtTime(0.001, t + 0.25);
			osc.connect(g);
			g.connect(dest);
			osc.start(t);
			osc.stop(t + 0.3);
			break;
		}
		default: {
			// 'Other' — short neutral blip
			const osc = ctx.createOscillator();
			const g = ctx.createGain();
			osc.type = 'sine';
			osc.frequency.value = 900;
			g.gain.setValueAtTime(amp * 0.3, t);
			g.gain.exponentialRampToValueAtTime(0.001, t + 0.05);
			osc.connect(g);
			g.connect(dest);
			osc.start(t);
			osc.stop(t + 0.06);
		}
	}
}

export interface Player {
	readonly playing: boolean;
	play(onEnded?: () => void): void;
	stop(): void;
}

export function createPlayer(hits: PlayableHit[], speed = 1): Player {
	let ctx: AudioContext | null = null;
	let endTimer: ReturnType<typeof setTimeout> | null = null;

	const sorted = [...hits].sort((a, b) => a.time_ms - b.time_ms);
	const duration = sorted.length > 0 ? sorted[sorted.length - 1].time_ms : 0;

	return {
		get playing() {
			return ctx !== null;
		},
		play(onEnded?: () => void) {
			if (ctx) return;
			ctx = new AudioContext();
			const master = ctx.createGain();
			master.gain.value = 0.85;
			master.connect(ctx.destination);

			const noiseBuffer = makeNoiseBuffer(ctx);
			const t0 = ctx.currentTime;

			for (const hit of sorted) {
				scheduleHit(ctx, master, noiseBuffer, hit.note_number, hit.velocity, t0 + hit.time_ms / speed / 1000);
			}

			endTimer = setTimeout(
				() => {
					this.stop();
					onEnded?.();
				},
				duration / speed + 600
			);
		},
		stop() {
			if (endTimer) {
				clearTimeout(endTimer);
				endTimer = null;
			}
			ctx?.close();
			ctx = null;
		}
	};
}
