<script lang="ts">
	import { noteLabel } from '$lib/drums';
	import { snap, type EditorNote } from '$lib/editor/notes';

	// Drum grid: one row per note number, time left to right. Click an empty
	// cell to add a hit, click a hit to remove it, shift-click to select it,
	// drag a hit (or the selection) to move it. Positions snap to `stepMs`.
	let {
		notes,
		rows,
		bpm,
		beatsPerBar,
		bars,
		stepMs,
		pxPerBeat = 64,
		selected,
		onAdd,
		onRemove,
		onMove,
		onToggleSelect
	}: {
		notes: EditorNote[];
		/** Note numbers, top to bottom. */
		rows: number[];
		bpm: number;
		beatsPerBar: number;
		bars: number;
		stepMs: number;
		pxPerBeat?: number;
		selected: Set<number>;
		onAdd: (noteNumber: number, timeMs: number) => void;
		onRemove: (id: number) => void;
		onMove: (ids: Set<number>, deltaMs: number, deltaRows: number) => void;
		onToggleSelect: (id: number) => void;
	} = $props();

	const ROW_H = 28;
	const HEADER_H = 20;
	const PAD_R = 24;

	const beatMs = $derived(60000 / bpm);
	const pxPerMs = $derived(pxPerBeat / beatMs);
	const width = $derived(bars * beatsPerBar * pxPerBeat + PAD_R);
	const height = $derived(HEADER_H + rows.length * ROW_H);
	const rowOf = $derived(new Map(rows.map((n, i) => [n, i])));

	// Grid lines: every step, every beat, every bar.
	const lines = $derived.by(() => {
		const out: { x: number; kind: 'step' | 'beat' | 'bar' }[] = [];
		const totalMs = bars * beatsPerBar * beatMs;
		const steps = Math.round(totalMs / stepMs);
		const showSteps = stepMs * pxPerMs >= 6;
		for (let i = 0; i <= steps; i++) {
			const ms = i * stepMs;
			const beats = ms / beatMs;
			const onBeat = Math.abs(beats - Math.round(beats)) < 1e-6;
			const kind = onBeat ? (Math.round(beats) % beatsPerBar === 0 ? 'bar' : 'beat') : 'step';
			if (kind === 'step' && !showSteps) continue;
			out.push({ x: ms * pxPerMs, kind });
		}
		return out;
	});

	let svg: SVGSVGElement;
	let drag = $state<{ id: number; startX: number; startY: number; dMs: number; dRows: number; moved: boolean } | null>(
		null
	);
	const dragging = $derived(drag?.moved ? (selected.has(drag.id) ? selected : new Set([drag.id])) : null);

	function point(e: PointerEvent) {
		const r = svg.getBoundingClientRect();
		return { x: e.clientX - r.left, y: e.clientY - r.top };
	}

	function pointerDown(e: PointerEvent) {
		if (e.button !== 0) return;
		// Hits carry their id; anything else is the background.
		const hitId = (e.target as Element).closest('[data-id]')?.getAttribute('data-id');
		const hitNote = hitId ? notes.find((n) => n.id === Number(hitId)) : undefined;
		if (hitNote) {
			noteDown(e, hitNote);
			return;
		}
		const { x, y } = point(e);
		const row = Math.floor((y - HEADER_H) / ROW_H);
		if (row < 0 || row >= rows.length) return;
		const ms = snap(x / pxPerMs, stepMs);
		if (ms > bars * beatsPerBar * beatMs) return;
		onAdd(rows[row], ms);
	}

	function noteDown(e: PointerEvent, n: EditorNote) {
		const { x, y } = point(e);
		drag = { id: n.id, startX: x, startY: y, dMs: 0, dRows: 0, moved: false };
		svg.setPointerCapture(e.pointerId);
	}

	function pointerMove(e: PointerEvent) {
		if (!drag) return;
		const { x, y } = point(e);
		const dx = x - drag.startX;
		const dy = y - drag.startY;
		if (!drag.moved && Math.hypot(dx, dy) < 4) return;
		const n = notes.find((m) => m.id === drag!.id);
		if (!n) return;
		// Snap where the grabbed hit lands; the rest of the selection follows.
		const target = snap(n.time_ms + dx / pxPerMs, stepMs);
		drag = { ...drag, moved: true, dMs: target - n.time_ms, dRows: Math.round(dy / ROW_H) };
	}

	function pointerUp(e: PointerEvent) {
		if (!drag) return;
		const d = drag;
		drag = null;
		if (svg.hasPointerCapture(e.pointerId)) svg.releasePointerCapture(e.pointerId);
		if (d.moved) {
			if (d.dMs !== 0 || d.dRows !== 0) onMove(selected.has(d.id) ? selected : new Set([d.id]), d.dMs, d.dRows);
		} else if (e.shiftKey || e.ctrlKey || e.metaKey) {
			onToggleSelect(d.id);
		} else {
			onRemove(d.id);
		}
	}

	function cx(n: EditorNote) {
		return (n.time_ms + (dragging?.has(n.id) ? drag!.dMs : 0)) * pxPerMs;
	}
	function cy(n: EditorNote) {
		let row = rowOf.get(n.note_number) ?? 0;
		if (dragging?.has(n.id)) row = Math.max(0, Math.min(rows.length - 1, row + drag!.dRows));
		return HEADER_H + row * ROW_H + ROW_H / 2;
	}
</script>

<div class="flex border border-line rounded-lg overflow-hidden bg-card">
	<div class="shrink-0 border-r border-line bg-card" aria-hidden="true">
		<div style:height="{HEADER_H}px"></div>
		{#each rows as n, i (n)}
			<div
				class="w-36 px-2 flex items-center gap-1 text-xs truncate {i % 2 ? 'bg-[var(--chart-row-alt)]' : 'bg-[var(--chart-row)]'}"
				style:height="{ROW_H}px"
				title="{noteLabel(n)} #{n}"
			>
				<span class="text-fg truncate">{noteLabel(n)}</span>
				<span class="text-fg-subtle font-mono">#{n}</span>
			</div>
		{/each}
	</div>
	<div class="overflow-x-auto flex-1 min-w-0">
		<svg
			bind:this={svg}
			{width}
			{height}
			class="block select-none touch-none cursor-crosshair"
			role="application"
			aria-label="Note grid: {notes.length} hits on {rows.length} rows, {bars} bars. Click an empty spot to add a hit, click a hit to remove it, shift-click to select, drag to move."
			onpointerdown={pointerDown}
			onpointermove={pointerMove}
			onpointerup={pointerUp}
			onpointercancel={() => (drag = null)}
		>
			{#each rows as n, i (n)}
				<rect
					x="0"
					y={HEADER_H + i * ROW_H}
					{width}
					height={ROW_H}
					style:fill={i % 2 ? 'var(--chart-row-alt)' : 'var(--chart-row)'}
				/>
			{/each}
			{#each lines as l (l.x)}
				<line
					x1={l.x}
					x2={l.x}
					y1={l.kind === 'bar' ? 0 : HEADER_H}
					y2={height}
					style:stroke={l.kind === 'step' ? 'var(--chart-grid)' : 'var(--chart-grid-strong)'}
					stroke-width={l.kind === 'bar' ? 1.5 : 1}
					stroke-dasharray={l.kind === 'step' ? '2 3' : undefined}
				/>
			{/each}
			{#each { length: bars } as _, b (b)}
				<text x={b * beatsPerBar * pxPerBeat + 4} y="14" font-size="11" style:fill="var(--fg-subtle)">{b + 1}</text>
			{/each}
			{#each notes as n (n.id)}
				{#if rowOf.has(n.note_number)}
					<circle
						cx={cx(n)}
						cy={cy(n)}
						r="7"
						class="cursor-pointer"
						style:fill="var(--accent-text)"
						style:fill-opacity={0.35 + (0.65 * n.velocity) / 127}
						style:stroke={selected.has(n.id) ? 'var(--fg)' : 'none'}
						stroke-width="2"
						data-id={n.id}
					>
						<title>{noteLabel(n.note_number)} at {(n.time_ms / 1000).toFixed(3)} s, velocity {n.velocity}</title>
					</circle>
				{/if}
			{/each}
		</svg>
	</div>
</div>
