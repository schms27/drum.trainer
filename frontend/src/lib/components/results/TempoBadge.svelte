<script lang="ts">
	import { describeTempo, formatTempo, tempoKind } from '$lib/tempo';

	// Tempo of a take as a small badge; highlighted when it isn't the original tempo.
	let { percent, practiceBpm = null }: { percent: number | null | undefined; practiceBpm?: number | null } = $props();

	const kind = $derived(tempoKind(percent));
	const label = $derived(describeTempo(percent, practiceBpm));
</script>

<span
	class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium tabular-nums whitespace-nowrap
		{kind === null ? 'text-fg-subtle' : kind === 'original' ? 'bg-control text-fg-muted' : 'bg-accent-soft text-accent-text'}"
	title={label}
	aria-label={label}
>
	{formatTempo(percent)}
</span>
