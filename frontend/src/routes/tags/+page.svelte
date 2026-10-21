<script lang="ts">
	import { Pencil, Trash } from '@lucide/svelte';
	import { createTag, deleteTag, updateTag, type Tag } from '$lib/api/tags';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import TagChip from '$lib/components/tags/TagChip.svelte';
	import ColorSwatches from '$lib/components/tags/ColorSwatches.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	let tags = $state<Tag[]>(data.tags);

	let editing = $state<number | null>(null);
	let editName = $state('');
	let editColor = $state('teal');
	let error = $state('');
	let toDelete = $state<Tag | null>(null);
	let confirmOpen = $state(false);

	let newName = $state('');
	let newColor = $state('teal');

	const sorted = $derived([...tags].sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })));

	function startEdit(tag: Tag) {
		editing = tag.id;
		editName = tag.name;
		editColor = tag.color;
		error = '';
	}

	async function saveEdit(e: SubmitEvent, tag: Tag) {
		e.preventDefault();
		error = '';
		const name = editName.trim();
		try {
			await updateTag(tag.id, { name, color: editColor });
			tags = tags.map((t) => (t.id === tag.id ? { ...t, name, color: editColor } : t));
			editing = null;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Could not save the tag';
		}
	}

	async function create(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			tags = [...tags, await createTag(newName.trim(), newColor)];
			newName = '';
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Could not create the tag';
		}
	}

	async function handleDelete() {
		if (!toDelete) return;
		await deleteTag(toDelete.id);
		tags = tags.filter((t) => t.id !== toDelete!.id);
	}

	const inputClass =
		'bg-control text-fg rounded-lg px-3 h-9 text-sm border border-line focus:outline-none focus:ring-2 focus:ring-accent';
</script>

<svelte:head><title>Tags — Drum Trainer</title></svelte:head>

<PageHeader title="Tags" subtitle="Your own labels for organising the library. Only you can see them." back={{ href: '/', label: 'Library' }} />

{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}

<Card title="New tag">
	<form class="flex flex-col gap-3" onsubmit={create}>
		<div class="flex flex-wrap items-center gap-2">
			<label class="sr-only" for="new-tag">Tag name</label>
			<input id="new-tag" class="{inputClass} w-56" bind:value={newName} maxlength="30" placeholder="e.g. Warm-up" />
			{#if newName.trim()}<TagChip tag={{ name: newName.trim(), color: newColor }} />{/if}
		</div>
		<ColorSwatches bind:value={newColor} label="Colour of the new tag" />
		<div><Button type="submit" variant="secondary" disabled={!newName.trim()}>Add tag</Button></div>
	</form>
</Card>

{#if sorted.length === 0}
	<EmptyState title="No tags yet" description="Create one here, or add tags from a track’s page." />
{:else}
	<Card title="Your tags">
		<ul class="flex flex-col divide-y divide-line-subtle -my-2">
			{#each sorted as tag (tag.id)}
				<li class="py-3">
					{#if editing === tag.id}
						<form class="flex flex-col gap-3" onsubmit={(e) => saveEdit(e, tag)}>
							<div class="flex flex-wrap items-center gap-2">
								<label class="sr-only" for="edit-{tag.id}">Name of {tag.name}</label>
								<input
									id="edit-{tag.id}"
									class="{inputClass} w-56"
									bind:value={editName}
									maxlength="30"
									onkeydown={(e) => e.key === 'Escape' && (editing = null)}
								/>
								<TagChip tag={{ name: editName.trim() || tag.name, color: editColor }} />
							</div>
							<ColorSwatches bind:value={editColor} label="Colour of {tag.name}" />
							<div class="flex gap-2">
								<Button type="submit" size="sm" variant="primary" disabled={!editName.trim()}>Save</Button>
								<Button size="sm" variant="ghost" onclick={() => (editing = null)}>Cancel</Button>
							</div>
						</form>
					{:else}
						<div class="flex items-center gap-3">
							<div class="flex-1 min-w-0"><TagChip {tag} /></div>
							<span class="text-xs text-fg-subtle tabular-nums">
								{tag.track_count}
								{tag.track_count === 1 ? 'track' : 'tracks'}
							</span>
							<Button size="sm" variant="ghost" onclick={() => startEdit(tag)} aria-label="Edit {tag.name}">
								<Pencil size={14} aria-hidden="true" />
							</Button>
							<Button
								size="sm"
								variant="ghost"
								onclick={() => {
									toDelete = tag;
									confirmOpen = true;
								}}
								aria-label="Delete {tag.name}"
							>
								<Trash size={14} aria-hidden="true" />
							</Button>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	</Card>
{/if}

<ConfirmDialog
	bind:open={confirmOpen}
	title="Delete the tag “{toDelete?.name}”?"
	message={toDelete?.track_count
		? `It’s removed from ${toDelete.track_count} ${toDelete.track_count === 1 ? 'track' : 'tracks'}. The tracks themselves are kept.`
		: 'No tracks use it.'}
	confirmLabel="Delete tag"
	danger
	onConfirm={handleDelete}
/>
