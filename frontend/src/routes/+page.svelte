<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { PencilRuler, Plus, Search, Tags, X } from '@lucide/svelte';
	import TagChip from '$lib/components/tags/TagChip.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import {
		SORTS,
		filterFromParams,
		filterToParams,
		filterTracks,
		parseQuery,
		type LibraryFilter,
		type LibrarySort
	} from '$lib/library-filter';
	import MidiUpload from '$lib/components/upload/MidiUpload.svelte';
	import TrackList from '$lib/components/tracks/TrackList.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import SetupChecklist from '$lib/components/onboarding/SetupChecklist.svelte';
	import type { Track, TrackWithNotes } from '$lib/types/track';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let tracks = $state<Track[]>(data.tracks);
	let showUpload = $state(false);

	// ── Search, tag filter and sort (kept in the URL) ───────────────────────
	const filter = $derived(filterFromParams(page.url.searchParams));
	// The text box is local while typing; the URL follows it.
	// svelte-ignore state_referenced_locally
	let query = $state(filter.q);
	// The last q this page wrote to the URL. Only a different one (Back,
	// Forward, a link) replaces the text box, so a URL update that lands
	// while the user keeps typing never swallows keystrokes.
	// svelte-ignore state_referenced_locally
	let writtenQ = filter.q;
	$effect(() => {
		if (filter.q !== writtenQ) {
			writtenQ = filter.q;
			query = filter.q;
		}
	});
	const allTags = $derived(
		[...new Map(tracks.flatMap((t) => t.tags ?? []).map((t) => [t.id, t])).values()].sort((a, b) =>
			a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
		)
	);
	const shown = $derived(filterTracks(tracks, { ...filter, q: query }));
	const words = $derived(parseQuery(query).words);
	const filtering = $derived(query.trim() !== '' || filter.tagIds.length > 0);

	function update(next: Partial<LibraryFilter>) {
		const params = filterToParams({ ...filter, q: query, ...next });
		writtenQ = params.get('q') ?? '';
		goto(params.size ? `/?${params}` : '/', { replaceState: true, keepFocus: true, noScroll: true });
	}
	function toggleTag(id: number) {
		update({ tagIds: filter.tagIds.includes(id) ? filter.tagIds.filter((t) => t !== id) : [...filter.tagIds, id] });
	}
	function clearFilters() {
		query = '';
		update({ q: '', tagIds: [] });
	}

	let searchInput = $state<HTMLInputElement>();
	function onWindowKey(e: KeyboardEvent) {
		if (e.key !== '/' || (e.target as HTMLElement).closest('input, textarea, select')) return;
		e.preventDefault();
		searchInput?.focus();
	}

	function handleUploaded(track: TrackWithNotes) {
		const { notes: _notes, ...rest } = track;
		tracks = [rest, ...tracks];
		showUpload = false;
	}
</script>

<svelte:head><title>Library — Drum Trainer</title></svelte:head>
<svelte:window onkeydown={onWindowKey} />

{#if tracks.length === 0}
	<PageHeader title="Welcome to Drum Trainer" subtitle="Play along with a MIDI drum track on your e-drum kit and see how accurately you hit every note." />

	<ol class="grid gap-4 sm:grid-cols-3 text-sm">
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">1. Add a track</p>
			<p class="text-fg-muted mt-1">Upload a MIDI file (.mid) with the drum part you want to practise.</p>
		</li>
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">2. Play along</p>
			<p class="text-fg-muted mt-1">Connect your e-drum kit by USB and play the track in Chrome or Edge.</p>
		</li>
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">3. Review</p>
			<p class="text-fg-muted mt-1">See which notes you hit, missed or played early or late.</p>
		</li>
	</ol>

	<div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
		<Card title="Check your setup" description="Recording needs a few things from your browser and kit.">
			<SetupChecklist />
		</Card>

		<Card title="Add your first track">
			<MidiUpload onUploaded={handleUploaded} />
			<p class="text-sm text-fg-muted mt-4">
				No MIDI file? <a href="/tracks/new" class="text-accent-text hover:underline">Write a track yourself</a> in the editor.
			</p>
		</Card>
	</div>
{:else}
	<PageHeader
		title="Library"
		subtitle={filtering ? `${shown.length} of ${tracks.length} tracks` : `${tracks.length} ${tracks.length === 1 ? 'track' : 'tracks'}`}
	>
		{#snippet actions()}
			<Button href="/tags" variant="ghost">
				<Tags size={16} aria-hidden="true" /> Tags
			</Button>
			<Button href="/tracks/new" variant="secondary">
				<PencilRuler size={16} aria-hidden="true" /> New track
			</Button>
			<Button variant={showUpload ? 'secondary' : 'primary'} onclick={() => (showUpload = !showUpload)} aria-expanded={showUpload}>
				{#if showUpload}
					Cancel
				{:else}
					<Plus size={16} aria-hidden="true" /> Add track
				{/if}
			</Button>
		{/snippet}
	</PageHeader>

	{#if showUpload}
		<Card title="Add a track" description="Upload a MIDI file (.mid) with the drum part you want to practise.">
			<MidiUpload onUploaded={handleUploaded} />
		</Card>
	{/if}

	<div class="flex flex-col gap-3">
		<div class="flex flex-wrap items-center gap-2">
			<div class="relative flex-1 min-w-[14rem]">
				<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none" aria-hidden="true" />
				<label class="sr-only" for="library-search">Search your tracks</label>
				<input
					bind:this={searchInput}
					id="library-search"
					type="search"
					value={query}
					oninput={(e) => {
						query = e.currentTarget.value;
						update({ q: query });
					}}
					onkeydown={(e) => e.key === 'Escape' && clearFilters()}
					placeholder="Search by name, file or tag…  ( / )"
					autocomplete="off"
					class="w-full bg-control text-fg rounded-lg pl-9 pr-3 h-10 text-sm border border-line focus:outline-none focus:ring-2 focus:ring-accent"
				/>
			</div>
			<label class="sr-only" for="library-sort">Sort</label>
			<select
				id="library-sort"
				value={filter.sort}
				onchange={(e) => update({ sort: e.currentTarget.value as LibrarySort })}
				class="bg-control text-fg text-sm rounded-lg px-2 h-10 border border-line focus:outline-none focus:ring-2 focus:ring-accent"
			>
				{#each SORTS as s (s.value)}<option value={s.value}>{s.label}</option>{/each}
			</select>
		</div>
		{#if allTags.length}
			<div class="flex flex-wrap items-center gap-1.5" role="group" aria-label="Filter by tag">
				{#each allTags as tag (tag.id)}
					{@const on = filter.tagIds.includes(tag.id)}
					<button
						type="button"
						aria-pressed={on}
						onclick={() => toggleTag(tag.id)}
						class="rounded-full cursor-pointer transition-opacity {on || !filter.tagIds.length ? '' : 'opacity-50 hover:opacity-100'} {on ? 'ring-2 ring-[var(--fg-muted)] ring-offset-1 ring-offset-[var(--canvas)]' : ''}"
					>
						<TagChip {tag} />
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#if shown.length}
		<TrackList tracks={shown} highlight={words} />
	{:else}
		<EmptyState title="No tracks match" description="Try other words, or remove a tag filter.">
			{#snippet icon()}<Search size={32} aria-hidden="true" />{/snippet}
			{#snippet action()}<Button onclick={clearFilters}><X size={16} aria-hidden="true" /> Clear filters</Button>{/snippet}
		</EmptyState>
	{/if}
{/if}
