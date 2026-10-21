<script lang="ts">
	import { Clock, Gauge, Globe, Link2, Play } from '@lucide/svelte';
	import type { Track } from '$lib/types/track';
	import Button from '$lib/components/ui/Button.svelte';
	import Sparkline from '$lib/components/charts/Sparkline.svelte';
	import TagChip from '$lib/components/tags/TagChip.svelte';
	import { formatPercent, formatRelativeDate, hitRateTone, toneText } from '$lib/score-format';

	import { highlight as highlightParts } from '$lib/library-filter';

	let { track, highlight = [] }: { track: Track; highlight?: string[] } = $props();
	const nameParts = $derived(highlightParts(track.name, highlight));

	const stats = $derived(track.stats);

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}
</script>

<article class="group relative bg-card border border-line rounded-xl p-5 flex flex-col gap-4 hover:border-accent transition-colors">
	<div class="min-w-0">
		<h3 class="font-semibold text-fg leading-tight break-words">
			<!-- The title link covers the whole card (see ::after), the Practice button sits above it. -->
			<a href="/tracks/{track.id}" class="after:absolute after:inset-0 after:rounded-xl focus-visible:outline-none">
				{#each nameParts as part, i (i)}{#if part.match}<mark class="bg-accent-soft text-fg rounded-sm">{part.text}</mark>{:else}{part.text}{/if}{/each}
			</a>
		</h3>
		<p class="flex flex-wrap gap-x-3 text-xs text-fg-muted mt-1.5">
			<span class="inline-flex items-center gap-1"><Clock size={13} aria-hidden="true" />{formatDuration(track.duration_ms)}</span>
			<span class="inline-flex items-center gap-1"><Gauge size={13} aria-hidden="true" />{Math.round(track.bpm)} BPM</span>
			<span>{track.note_count} notes</span>
			{#if track.disabled}
				<span class="text-bad font-medium">Removed</span>
			{:else if track.visibility === 'public'}
				<span class="inline-flex items-center gap-1 text-accent-text"><Globe size={13} aria-hidden="true" />Public</span>
			{:else if track.share_token}
				<span class="inline-flex items-center gap-1 text-accent-text"><Link2 size={13} aria-hidden="true" />Shared</span>
			{:else if track.copied}
				<span class="text-fg-subtle">From a share link</span>
			{/if}
		</p>
		{#if track.tags?.length}
			<ul class="flex flex-wrap gap-1 mt-2" aria-label="Tags">
				{#each track.tags.slice(0, 3) as tag (tag.id)}
					<li class="min-w-0"><TagChip {tag} /></li>
				{/each}
				{#if track.tags.length > 3}
					<li class="text-xs text-fg-subtle self-center" title={track.tags.slice(3).map((t) => t.name).join(', ')}>
						+{track.tags.length - 3}
					</li>
				{/if}
			</ul>
		{/if}
	</div>

	{#if stats && stats.sessions > 0 && stats.best_hit_rate !== null}
		<div class="flex items-end justify-between gap-3">
			<div>
				<p class="text-xs text-fg-muted">Best</p>
				<p class="text-xl font-bold tabular-nums {toneText[hitRateTone(stats.best_hit_rate)]}">{formatPercent(stats.best_hit_rate)}</p>
				<p class="text-xs text-fg-subtle">
					{[`${stats.sessions} ${stats.sessions === 1 ? 'take' : 'takes'}`, stats.last_practiced_at && formatRelativeDate(stats.last_practiced_at)]
						.filter(Boolean)
						.join(' · ')}
				</p>
			</div>
			<Sparkline values={stats.recent_hit_rates} label="Notes hit in the last {stats.recent_hit_rates.length} takes" />
		</div>
	{:else}
		<p class="text-sm text-fg-subtle">Not practised yet</p>
	{/if}

	{#if track.disabled}
		<Button href="/tracks/{track.id}" variant="secondary" block class="relative z-10">Why was it removed?</Button>
	{:else}
		<Button href="/session/{track.id}" variant="primary" block class="relative z-10">
			<Play size={16} aria-hidden="true" /> Practice
		</Button>
	{/if}
</article>
