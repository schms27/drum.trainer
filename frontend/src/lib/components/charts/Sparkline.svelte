<script lang="ts">
	// A tiny trend line of hit rates (0–1), oldest first.
	let {
		values,
		width = 96,
		height = 28,
		label = 'Trend'
	}: { values: number[]; width?: number; height?: number; label?: string } = $props();

	const PAD = 3;

	// Scale to the values' own range (at least 10 points) so small changes show.
	const MIN_RANGE = 0.1;
	const range = $derived.by(() => {
		const lo = Math.min(...values);
		const hi = Math.max(...values);
		const pad = Math.max(0, MIN_RANGE - (hi - lo)) / 2;
		return { lo: lo - pad, hi: hi + pad };
	});

	const points = $derived(
		values.map((v, i) => ({
			x: values.length === 1 ? width / 2 : PAD + (i / (values.length - 1)) * (width - 2 * PAD),
			y: PAD + (1 - (v - range.lo) / (range.hi - range.lo)) * (height - 2 * PAD)
		}))
	);
	const path = $derived(points.map((p, i) => `${i ? 'L' : 'M'}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' '));
	const last = $derived(points[points.length - 1]);
</script>

{#if values.length > 0}
	<svg {width} {height} viewBox="0 0 {width} {height}" role="img" aria-label={label}>
		<path d={path} fill="none" style:stroke="var(--accent-text)" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" />
		<circle cx={last.x} cy={last.y} r="2.5" style:fill="var(--accent-text)" />
	</svg>
{/if}
