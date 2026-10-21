<script lang="ts">
	import { goto } from '$app/navigation';
	import { Play, SlidersHorizontal, Trash, ChevronRight, Share2, PencilRuler } from '@lucide/svelte';
	import { deleteTrack } from '$lib/api/tracks';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import BackingTracks from '$lib/components/tracks/BackingTracks.svelte';
	import ShareDialog from '$lib/components/tracks/ShareDialog.svelte';
	import AccuracyChart from '$lib/components/charts/AccuracyChart.svelte';
	import TempoBadge from '$lib/components/results/TempoBadge.svelte';
	import { TEMPO_FILTERS, matchesTempo, type TempoFilter } from '$lib/tempo';
	import {
		formatOffset,
		formatPercent,
		formatRelativeDate,
		hitRateTone,
		timingTone,
		toneText
	} from '$lib/score-format';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	const { takes, hasMoreTakes } = data;
	// svelte-ignore state_referenced_locally
	let track = $state(data.track);
	// svelte-ignore state_referenced_locally
	let backing = $state(data.pilots);
	let confirmDelete = $state(false);
	let shareOpen = $state(false);

	// svelte-ignore state_referenced_locally
	const stats = data.track.stats;

	// Tempo filter for the tiles, chart and recent takes. "Any tempo" uses the
	// server's stats (all takes); otherwise they come from the loaded takes.
	let tempo = $state<TempoFilter>('all');
	const filtered = $derived(takes.filter((t) => matchesTempo(t.tempo_percent, tempo)));
	const shown = $derived(
		tempo === 'all'
			? { best: stats?.best_hit_rate ?? 0, last: stats?.last_hit_rate ?? 0, count: stats?.sessions ?? 0 }
			: {
					best: filtered.reduce((m, t) => Math.max(m, t.hit_rate), 0),
					last: filtered[0]?.hit_rate ?? 0,
					count: filtered.length
				}
	);
	// The API lists newest first; the chart wants oldest first.
	const chronological = $derived([...filtered].reverse());
	const tempoLabel = $derived(TEMPO_FILTERS.find((f) => f.value === tempo)?.label.toLowerCase() ?? '');

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
	}

	async function handleDelete() {
		await deleteTrack(track.id);
		goto('/');
	}
</script>

<svelte:head><title>{track.name} — Drum Trainer</title></svelte:head>

<PageHeader
	title={track.name}
	subtitle="{formatDuration(track.duration_ms)} · {Math.round(track.bpm)} BPM · {track.note_count} notes · {track.filename}"
	back={{ href: '/', label: 'Library' }}
>
	{#snippet actions()}
		{#if !track.disabled}
			<Button href="/session/{track.id}" variant="primary" size="lg">
				<Play size={18} aria-hidden="true" /> Practice
			</Button>
		{/if}
	{/snippet}
</PageHeader>

{#if track.disabled}
	<Notice variant="error">
		<p class="font-medium">
			{track.copied ? 'The track this was copied from was removed' : 'This track was removed'}, so it can’t be practised or
			shared any more.
		</p>
		{#if track.disabled_reason}<p class="mt-1">Reason: {track.disabled_reason}</p>{/if}
		<p class="mt-1 text-fg-muted">
			Your takes are kept. See our <a href="/copyright" class="text-accent-text hover:underline">copyright policy</a> if you
			think this is a mistake.
		</p>
	</Notice>
{/if}

{#if stats && stats.sessions > 0}
	<div class="flex items-center justify-end gap-2 -mb-2">
		<label class="text-sm text-fg-muted" for="tempo-filter">Show</label>
		<select
			id="tempo-filter"
			class="bg-control text-fg text-sm rounded-lg px-2 h-9 border border-line focus:outline-none focus:ring-2 focus:ring-accent"
			bind:value={tempo}
		>
			{#each TEMPO_FILTERS as f (f.value)}
				<option value={f.value}>{f.label}</option>
			{/each}
		</select>
	</div>
	<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Best</p>
			<p class="text-3xl font-bold mt-1 tabular-nums {shown.count ? toneText[hitRateTone(shown.best)] : 'text-fg-subtle'}">
				{shown.count ? formatPercent(shown.best) : '—'}
			</p>
			<p class="text-xs text-fg-subtle mt-1">notes hit{tempo === 'all' ? '' : ` · ${tempoLabel}`}</p>
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Last take</p>
			<p class="text-3xl font-bold mt-1 tabular-nums {shown.count ? toneText[hitRateTone(shown.last)] : 'text-fg-subtle'}">
				{shown.count ? formatPercent(shown.last) : '—'}
			</p>
			<p class="text-xs text-fg-subtle mt-1">notes hit{tempo === 'all' ? '' : ` · ${tempoLabel}`}</p>
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Takes</p>
			<p class="text-3xl font-bold mt-1 tabular-nums text-fg">{shown.count}</p>
			{#if tempo !== 'all'}<p class="text-xs text-fg-subtle mt-1">{tempoLabel}</p>{/if}
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Last practised</p>
			<p class="text-xl font-bold mt-2 text-fg first-letter:uppercase">
				{stats.last_practiced_at ? formatRelativeDate(stats.last_practiced_at) : '—'}
			</p>
		</div>
	</div>

	<Card title="Notes hit per take" description={hasMoreTakes ? 'Your latest 100 takes.' : undefined}>
		{#if chronological.length > 0}
			<AccuracyChart takes={chronological} />
		{:else}
			<p class="text-sm text-fg-subtle">No takes at this tempo yet.</p>
		{/if}
	</Card>

	<Card title="Recent takes">
		<ul class="flex flex-col divide-y divide-line-subtle -my-2">
			{#each filtered.slice(0, 5) as take (take.id)}
				<li>
					<a href="/results/{take.id}" class="group flex items-center gap-4 py-2.5">
						<span class="flex-1 text-sm text-fg">{formatDate(take.created_at)}</span>
						<TempoBadge percent={take.tempo_percent} practiceBpm={take.practice_bpm} />
						<span class="text-sm font-semibold tabular-nums {toneText[hitRateTone(take.hit_rate)]}">{formatPercent(take.hit_rate)}</span>
						<span class="hidden sm:inline text-sm tabular-nums w-16 text-right {toneText[timingTone(take.avg_offset_ms)]}">
							{formatOffset(take.avg_offset_ms)}
						</span>
						<ChevronRight size={16} class="text-fg-subtle group-hover:text-accent-text" aria-hidden="true" />
					</a>
				</li>
			{/each}
		</ul>
		{#if filtered.length === 0}
			<p class="text-sm text-fg-subtle">No takes at this tempo yet.</p>
		{/if}
		{#if filtered.length > 5}
			<a
				href="/progress?track={track.id}{tempo === 'all' ? '' : `&tempo=${tempo}`}"
				class="inline-block mt-4 text-sm text-accent-text hover:underline"
			>
				All {tempo === 'all' ? stats.sessions : filtered.length} {tempo === 'all' ? '' : tempoLabel + ' '}takes in Progress
			</a>
		{/if}
	</Card>
{:else}
	<EmptyState title="No takes yet" description="Record your first take — your progress on this track will show up here.">
		{#snippet action()}
			{#if !track.disabled}
				<Button href="/session/{track.id}" variant="primary"><Play size={16} aria-hidden="true" /> Practice now</Button>
			{/if}
		{/snippet}
	</EmptyState>
{/if}

<div class="grid gap-4 md:grid-cols-2">
	<Card title="Backing tracks" description="Audio that plays along while you record — for example the song without drums.">
		<BackingTracks trackId={track.id} bind:items={backing} />
	</Card>

	<div class="flex flex-col gap-4">
		<Card title="Notes" description="Change the drum part: simplify it for a smaller kit, or add, move and remove hits.">
			<div class="flex flex-wrap gap-2">
				{#if !track.disabled}
					<Button href="/tracks/{track.id}/edit" variant="secondary">
						<PencilRuler size={16} aria-hidden="true" /> Edit notes
					</Button>
				{/if}
				<Button href="/tracks/{track.id}/mapping" variant="secondary">
					<SlidersHorizontal size={16} aria-hidden="true" /> Edit pad mapping
				</Button>
			</div>
		</Card>

		{#if track.copied}
			<Card title="Sharing" description="You added this track from someone’s share link, so it can’t be shared further.">
				<p class="text-sm text-fg-muted">It’s your own copy: changes you make don’t affect the original.</p>
			</Card>
		{:else if !track.disabled}
			<Card
				title="Sharing"
				description={track.visibility === 'public'
					? 'Public. Listed in Explore, so any drummer can find it and add a copy.'
					: track.share_token
						? 'Shared by link. Anyone with the link can add a copy to their library.'
						: 'Private. Share it by link or list it in Explore so other drummers can practise it.'}
			>
				<Button variant="secondary" onclick={() => (shareOpen = true)}>
					<Share2 size={16} aria-hidden="true" />
					{track.share_token ? 'Manage sharing' : 'Share'}
				</Button>
			</Card>
		{/if}

		<Card title="Delete track" description="Removes the track, its backing tracks and all takes recorded for it.">
			<Button variant="secondary" onclick={() => (confirmDelete = true)} class="text-bad">
				<Trash size={16} aria-hidden="true" /> Delete track
			</Button>
		</Card>
	</div>
</div>

<ShareDialog bind:open={shareOpen} bind:track />

<ConfirmDialog
	bind:open={confirmDelete}
	title="Delete “{track.name}”?"
	message={track.share_token
		? 'This removes the track, its backing tracks and all practice sessions recorded for it, and the share link stops working. Copies others already added are kept. This cannot be undone.'
		: 'This removes the track, its backing tracks and all practice sessions recorded for it. This cannot be undone.'}
	confirmLabel="Delete track"
	danger
	onConfirm={handleDelete}
/>
