<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import { createTag, listTags, setTrackTags, type Tag } from '$lib/api/tags';
	import type { TagRef } from '$lib/types/track';
	import TagChip from './TagChip.svelte';
	import ColorSwatches from './ColorSwatches.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { TAG_COLOR_KEYS } from '$lib/tag-colors';

	// The tags on one track: remove with the chip's ×, add an existing tag or
	// create a new one from the typed text.
	let { trackId, tags = $bindable() }: { trackId: number; tags: TagRef[] } = $props();

	let all = $state<Tag[] | null>(null);
	let open = $state(false);
	let query = $state('');
	let newColor = $state<string>(TAG_COLOR_KEYS[5]);
	let error = $state('');
	let busy = $state(false);
	let input = $state<HTMLInputElement>();

	const applied = $derived(new Set(tags.map((t) => t.id)));
	const q = $derived(query.trim().toLowerCase());
	const matches = $derived((all ?? []).filter((t) => !applied.has(t.id) && t.name.toLowerCase().includes(q)));
	const exact = $derived((all ?? []).some((t) => t.name.toLowerCase() === q));

	async function startAdding() {
		open = true;
		error = '';
		if (!all) all = await listTags().catch(() => []);
		queueMicrotask(() => input?.focus());
	}

	async function save(ids: number[]) {
		busy = true;
		error = '';
		try {
			tags = await setTrackTags(trackId, ids);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not save the tags';
		} finally {
			busy = false;
		}
	}

	async function add(tag: TagRef) {
		await save([...tags.map((t) => t.id), tag.id]);
		query = '';
	}

	async function createAndAdd() {
		const name = query.trim();
		if (!name) return;
		busy = true;
		try {
			const tag = await createTag(name, newColor);
			all = [...(all ?? []), tag];
			await add(tag);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not create the tag';
			busy = false;
		}
	}

	function onKey(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			query = '';
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (matches.length && (exact || matches.length === 1)) add(matches.find((t) => t.name.toLowerCase() === q) ?? matches[0]);
			else if (q && !exact) createAndAdd();
		}
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex flex-wrap items-center gap-1.5">
		{#each tags as tag (tag.id)}
			<TagChip {tag} onRemove={busy ? undefined : () => save(tags.filter((t) => t.id !== tag.id).map((t) => t.id))} />
		{/each}
		{#if !open}
			<Button size="sm" variant="ghost" onclick={startAdding}><Plus size={14} aria-hidden="true" /> Add tag</Button>
		{/if}
	</div>

	{#if open}
		<div class="flex flex-col gap-2 border border-line rounded-lg p-3">
			<label class="sr-only" for="tag-search-{trackId}">Find or create a tag</label>
			<input
				bind:this={input}
				id="tag-search-{trackId}"
				bind:value={query}
				onkeydown={onKey}
				maxlength="30"
				placeholder="Find or create a tag…"
				autocomplete="off"
				class="bg-control text-fg rounded-lg px-3 h-9 text-sm border border-line focus:outline-none focus:ring-2 focus:ring-accent"
			/>
			{#if matches.length}
				<ul class="flex flex-wrap gap-1.5" aria-label="Matching tags">
					{#each matches as tag (tag.id)}
						<li>
							<button type="button" class="cursor-pointer rounded-full" onclick={() => add(tag)} disabled={busy} aria-label="Add tag {tag.name}">
								<TagChip {tag} />
							</button>
						</li>
					{/each}
				</ul>
			{/if}
			{#if q && !exact}
				<div class="flex flex-col gap-2">
					<p class="text-sm text-fg-muted">Create a new tag <TagChip tag={{ name: query.trim(), color: newColor }} /></p>
					<ColorSwatches bind:value={newColor} label="Colour of the new tag" />
					<div><Button size="sm" variant="secondary" onclick={createAndAdd} disabled={busy}>Create and add</Button></div>
				</div>
			{:else if all && all.length === 0 && !q}
				<p class="text-sm text-fg-subtle">You have no tags yet. Type a name to create one.</p>
			{/if}
			<div><Button size="sm" variant="ghost" onclick={() => ((open = false), (query = ''))}>Done</Button></div>
		</div>
	{/if}
	{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}
</div>
