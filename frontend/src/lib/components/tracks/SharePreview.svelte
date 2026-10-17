<script lang="ts">
	import { LANES } from '$lib/drums';
	import type { NoteEvent } from '$lib/types/track';

	let {
		notes,
		durationMs,
		bpm,
		timeSigNumerator
	}: {
		notes: NoteEvent[];
		durationMs: number;
		bpm: number;
		timeSigNumerator: number;
	} = $props();

	const WIDTH = 680;
	const ROW_H = 24;
	const LABEL_W = 90;

	// Only lanes that have notes, in kit order; anything unmapped goes last.
	const lanes = $derived.by(() => {
		const used = new Set(notes.map((n) => n.instrument));
		const ordered: string[] = LANES.filter((l) => used.has(l));
		if ([...used].some((u) => !ordered.includes(u))) ordered.push('Other');
		return ordered;
	});
	const height = $derived(Math.max(lanes.length, 1) * ROW_H);
	const x = (ms: number) => LABEL_W + (ms / Math.max(durationMs, 1)) * (WIDTH - LABEL_W - 8);
	const laneOf = (instr: string) => {
		const i = lanes.indexOf(instr);
		return i >= 0 ? i : lanes.length - 1;
	};

	const barLines = $derived.by(() => {
		const barMs = (60000 / (bpm || 120)) * (timeSigNumerator || 4);
		const lines: number[] = [];
		for (let t = 0; t <= durationMs + 1; t += barMs) lines.push(t);
		return lines;
	});
</script>

<svg viewBox="0 0 {WIDTH} {height}" class="w-full h-auto" role="img" aria-label="Drum notes in the opening bars">
	{#each lanes as lane, i (lane)}
		<text x="0" y={i * ROW_H + ROW_H / 2} dominant-baseline="middle" class="fill-fg-muted text-[11px]">{lane}</text>
		<line x1={LABEL_W} x2={WIDTH - 8} y1={i * ROW_H + ROW_H / 2} y2={i * ROW_H + ROW_H / 2} class="stroke-line-subtle" />
	{/each}
	{#each barLines as t (t)}
		<line x1={x(t)} x2={x(t)} y1="0" y2={height} class="stroke-line" />
	{/each}
	{#each notes as n, i (i)}
		<circle
			cx={x(n.time_ms)}
			cy={laneOf(n.instrument) * ROW_H + ROW_H / 2}
			r={3 + (n.velocity / 127) * 3}
			class="fill-accent"
		/>
	{/each}
</svg>
