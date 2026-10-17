<script lang="ts">
	import { goto } from '$app/navigation';
	import { ChevronRight, Flame, ListMusic } from '@lucide/svelte';
	import { listSessions } from '$lib/api/sessions';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import AccuracyChart from '$lib/components/charts/AccuracyChart.svelte';
	import WeeklyBars from '$lib/components/stats/WeeklyBars.svelte';
	import { formatOffset, formatPercent, hitRateTone, timingTone, toneText } from '$lib/score-format';
	import type { SessionSummary } from '$lib/types/session';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const stats = $derived(data.stats);
	const thisWeek = $derived(stats.weekly[stats.weekly.length - 1]);
	const trackNames = $derived(new Map(data.tracks.map((t) => [t.id, t.name])));
	// Tracks with at least one take, most practised first.
	const practised = $derived(
		data.tracks
			.filter((t) => (t.stats?.sessions ?? 0) > 0)
			.sort((a, b) => (b.stats?.sessions ?? 0) - (a.stats?.sessions ?? 0))
	);

	// ── Accuracy chart for one track ────────────────────────────────────────
	// svelte-ignore state_referenced_locally
	let chartTrackId = $state(data.trackFilter ?? practised[0]?.id ?? 0);
	let chartTakes = $state<SessionSummary[]>([]);
	let chartLoading = $state(false);

	$effect(() => {
		const id = chartTrackId;
		if (!id) return;
		let cancelled = false;
		chartLoading = true;
		listSessions({ trackId: id, limit: 100 })
			.then((page) => {
				if (!cancelled) chartTakes = [...page.sessions].reverse();
			})
			.finally(() => {
				if (!cancelled) chartLoading = false;
			});
		return () => {
			cancelled = true;
		};
	});

	// ── Takes list (filter lives in the URL so it can be linked to) ─────────
	// svelte-ignore state_referenced_locally
	let takes = $state(data.takes);
	// svelte-ignore state_referenced_locally
	let hasMore = $state(data.hasMore);
	let loadingMore = $state(false);

	$effect(() => {
		takes = data.takes;
		hasMore = data.hasMore;
	});

	function setFilter(value: string) {
		const id = Number(value);
		goto(id ? `/progress?track=${id}` : '/progress', { keepFocus: true, noScroll: true });
		if (id) chartTrackId = id;
	}

	async function loadMore() {
		loadingMore = true;
		try {
			const page = await listSessions({ trackId: data.trackFilter ?? undefined, before: takes[takes.length - 1].id });
			takes = [...takes, ...page.sessions];
			hasMore = page.hasMore;
		} finally {
			loadingMore = false;
		}
	}

	function fmtMinutes(m: number) {
		if (m < 60) return `${Math.round(m)} min`;
		const h = Math.floor(m / 60);
		const rem = Math.round(m % 60);
		return rem > 0 ? `${h} h ${rem} min` : `${h} h`;
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
	}

	const selectClass =
		'bg-control text-fg text-sm rounded-lg px-2 h-9 border border-line max-w-full focus:outline-none focus:ring-2 focus:ring-accent';
</script>

<svelte:head><title>Progress — Drum Trainer</title></svelte:head>

<PageHeader title="Progress" subtitle="How your practice is going." />

{#if stats.totals.sessions === 0 && takes.length === 0}
	<EmptyState
		title="No takes recorded yet"
		description="Pick a track in your library, press Start and play along. Your progress will show up here."
	>
		{#snippet icon()}<ListMusic size={32} aria-hidden="true" />{/snippet}
		{#snippet action()}<Button href="/" variant="primary">Go to library</Button>{/snippet}
	</EmptyState>
{:else}
	<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Streak</p>
			<p class="text-3xl font-bold mt-1 tabular-nums text-fg flex items-center gap-1.5">
				{#if stats.streak_days > 0}<Flame size={24} class="text-warn" aria-hidden="true" />{/if}
				{stats.streak_days}
				<span class="text-base font-normal text-fg-muted">{stats.streak_days === 1 ? 'day' : 'days'}</span>
			</p>
			<p class="text-xs text-fg-subtle mt-1">
				{#if stats.practiced_today}
					Practised today
				{:else if stats.streak_days > 0}
					Play today to keep it going
				{:else}
					Practise on consecutive days to build one
				{/if}
			</p>
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">This week</p>
			<p class="text-3xl font-bold mt-1 tabular-nums text-fg">{fmtMinutes(thisWeek?.minutes_played ?? 0)}</p>
			<p class="text-xs text-fg-subtle mt-1">{thisWeek?.session_count ?? 0} {thisWeek?.session_count === 1 ? 'take' : 'takes'}</p>
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Past 12 months</p>
			<p class="text-3xl font-bold mt-1 tabular-nums text-fg">{fmtMinutes(stats.totals.minutes_played)}</p>
			<p class="text-xs text-fg-subtle mt-1">{stats.totals.sessions} takes</p>
		</div>
		<div class="bg-card border border-line rounded-xl p-5">
			<p class="text-xs font-medium text-fg-muted">Tracks practised</p>
			<p class="text-3xl font-bold mt-1 tabular-nums text-fg">{stats.totals.tracks_practiced}</p>
			<p class="text-xs text-fg-subtle mt-1">in the past 12 months</p>
		</div>
	</div>

	{#if practised.length > 0}
		<Card title="Notes hit per take" description="Are you getting more accurate? Click a point to open that take.">
			{#snippet actions()}
				<label class="sr-only" for="chart-track">Track</label>
				<select id="chart-track" class={selectClass} bind:value={chartTrackId}>
					{#each practised as t (t.id)}
						<option value={t.id}>{t.name}</option>
					{/each}
				</select>
			{/snippet}
			{#if chartTakes.length > 0}
				<div class:opacity-50={chartLoading} class="transition-opacity">
					<AccuracyChart takes={chartTakes} />
				</div>
			{:else}
				<p class="text-sm text-fg-subtle">Loading…</p>
			{/if}
		</Card>
	{/if}

	<WeeklyBars data={stats.weekly} valueKey="minutes_played" label="Minutes played per week" unit="min" />

	<Card title="All takes">
		{#snippet actions()}
			<label class="sr-only" for="takes-filter">Show takes for</label>
			<select
				id="takes-filter"
				class={selectClass}
				value={String(data.trackFilter ?? '')}
				onchange={(e) => setFilter(e.currentTarget.value)}
			>
				<option value="">All tracks</option>
				{#each practised as t (t.id)}
					<option value={String(t.id)}>{t.name}</option>
				{/each}
			</select>
		{/snippet}

		{#if takes.length === 0}
			<p class="text-sm text-fg-subtle">No takes for this track yet.</p>
		{:else}
			<ul class="flex flex-col divide-y divide-line-subtle -my-2">
				{#each takes as take (take.id)}
					<li>
						<a href="/results/{take.id}" class="group flex items-center gap-4 py-3">
							<div class="min-w-0 flex-1">
								<p class="font-medium text-fg truncate">{trackNames.get(take.track_id) ?? 'Deleted track'}</p>
								<p class="text-xs text-fg-subtle mt-0.5">{formatDate(take.created_at)}</p>
							</div>
							<dl class="flex items-center gap-4 sm:gap-8 text-right shrink-0">
								<div>
									<dt class="text-xs text-fg-subtle">Notes hit</dt>
									<dd class="font-semibold tabular-nums {toneText[hitRateTone(take.hit_rate)]}">
										{formatPercent(take.hit_rate)}
										<span class="hidden sm:inline text-xs font-normal text-fg-subtle">{take.matched_count}/{take.target_count}</span>
									</dd>
								</div>
								<div class="hidden sm:block">
									<dt class="text-xs text-fg-subtle">Average timing</dt>
									<dd class="font-semibold tabular-nums {toneText[timingTone(take.avg_offset_ms)]}">
										{formatOffset(take.avg_offset_ms)}
									</dd>
								</div>
							</dl>
							<ChevronRight size={18} class="text-fg-subtle group-hover:text-accent-text shrink-0" aria-hidden="true" />
						</a>
					</li>
				{/each}
			</ul>
			{#if hasMore}
				<div class="text-center mt-4">
					<Button onclick={loadMore} disabled={loadingMore}>{loadingMore ? 'Loading…' : 'Show older takes'}</Button>
				</div>
			{/if}
		{/if}
	</Card>
{/if}
