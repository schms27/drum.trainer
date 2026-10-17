<script lang="ts">
	import { goto } from '$app/navigation';
	import { Clock, Gauge, LibraryBig, Search } from '@lucide/svelte';
	import { getExplore } from '$lib/api/sharing';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import type { PublicTrack } from '$lib/types/track';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// Reset whenever the load function reruns (new search or sort); "Load more" appends.
	let tracks = $state<PublicTrack[]>([]);
	let page = $state(1);
	let hasMore = $state(false);
	let search = $state('');
	$effect.pre(() => {
		tracks = data.first.tracks;
		page = 1;
		hasMore = data.first.has_more;
		search = data.query;
	});

	let loadingMore = $state(false);
	let error = $state('');

	const sorts = [
		{ value: 'newest', label: 'Newest' },
		{ value: 'popular', label: 'Most added' }
	] as const;

	const sourceLabel: Record<string, string> = {
		original: 'Original',
		licensed: 'Licensed',
		public_domain: 'Public domain'
	};

	function href(q: string, sort: string) {
		const params = new URLSearchParams();
		if (q) params.set('q', q);
		if (sort !== 'newest') params.set('sort', sort);
		const qs = params.toString();
		return qs ? `/explore?${qs}` : '/explore';
	}

	function submitSearch(e: SubmitEvent) {
		e.preventDefault();
		goto(href(search.trim(), data.sort), { keepFocus: true, noScroll: true });
	}

	async function loadMore() {
		loadingMore = true;
		error = '';
		try {
			const next = await getExplore(data.query, data.sort, page + 1);
			tracks = [...tracks, ...next.tracks];
			page = next.page;
			hasMore = next.has_more;
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not load more tracks.';
		} finally {
			loadingMore = false;
		}
	}

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}
</script>

<svelte:head><title>Explore — Drum Trainer</title></svelte:head>

<PageHeader title="Explore" subtitle="Drum tracks other drummers have shared publicly. Add one to your library to practise it." />

<div class="flex flex-col sm:flex-row sm:items-center gap-3 justify-between">
	<form onsubmit={submitSearch} role="search" class="relative flex-1 max-w-md">
		<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none" aria-hidden="true" />
		<input
			type="search"
			bind:value={search}
			maxlength="100"
			placeholder="Search by name"
			aria-label="Search public tracks by name"
			class="w-full h-10 rounded-lg border border-line bg-card pl-9 pr-3 text-sm text-fg"
		/>
	</form>
	<nav class="flex gap-1 rounded-lg bg-control p-1 w-fit" aria-label="Sort">
		{#each sorts as s (s.value)}
			<a
				href={href(data.query, s.value)}
				aria-current={data.sort === s.value ? 'page' : undefined}
				class="px-3 py-1.5 rounded-md text-sm {data.sort === s.value ? 'bg-card text-fg font-medium shadow-sm' : 'text-fg-muted hover:text-fg'}"
				>{s.label}</a
			>
		{/each}
	</nav>
</div>

{#if tracks.length === 0}
	<EmptyState
		title={data.query ? `No public tracks match “${data.query}”` : 'No public tracks yet'}
		description={data.query ? 'Try another name.' : 'Tracks appear here when their owners share them with everyone.'}
	/>
{:else}
	<ul class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each tracks as t (t.share_token)}
			<li>
				<a
					href="/s/{t.share_token}"
					class="h-full bg-card border border-line rounded-xl p-5 flex flex-col gap-3 hover:border-accent transition-colors"
				>
					<h2 class="font-semibold text-fg leading-tight break-words">{t.name}</h2>
					<p class="flex flex-wrap gap-x-3 gap-y-1 text-xs text-fg-muted">
						<span class="inline-flex items-center gap-1"><Clock size={13} aria-hidden="true" />{formatDuration(t.duration_ms)}</span>
						<span class="inline-flex items-center gap-1"><Gauge size={13} aria-hidden="true" />{Math.round(t.bpm)} BPM</span>
						<span>{t.note_count} notes</span>
					</p>
					<p class="mt-auto flex items-center justify-between gap-2 text-xs">
						<span class="rounded bg-control px-2 py-0.5 text-fg-muted">{sourceLabel[t.rights_basis] ?? t.rights_basis}</span>
						{#if t.copies > 0}
							<span class="inline-flex items-center gap-1 text-fg-subtle">
								<LibraryBig size={13} aria-hidden="true" />Added {t.copies}×
							</span>
						{/if}
					</p>
				</a>
			</li>
		{/each}
	</ul>
	{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}
	{#if hasMore}
		<div class="flex justify-center">
			<Button variant="secondary" onclick={loadMore} disabled={loadingMore}>Load more</Button>
		</div>
	{/if}
{/if}

<p class="text-xs text-fg-subtle">
	Tracks are shared by their owners, who are responsible for them.
	<a href="/copyright" class="hover:text-fg hover:underline">Report a copyright problem</a>.
</p>
