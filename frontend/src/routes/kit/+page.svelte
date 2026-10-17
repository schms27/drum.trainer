<script lang="ts">
	import { Drum, Pencil, Plus, Trash, CircleCheck } from '@lucide/svelte';
	import { deleteKit, setActiveKit, type KitProfile } from '$lib/api/kits';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import KitEditor from '$lib/components/kit/KitEditor.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// svelte-ignore state_referenced_locally
	let kits = $state<KitProfile[]>(data.kits);
	/** 'new', a kit id being edited, or null. */
	let editing = $state<'new' | number | null>(null);
	let toDelete = $state<KitProfile | null>(null);
	let confirmOpen = $state(false);
	let error = $state('');

	const activeKit = $derived(kits.find((k) => k.active) ?? null);

	function handleSaved(saved: KitProfile) {
		const exists = kits.some((k) => k.id === saved.id);
		kits = exists ? kits.map((k) => (k.id === saved.id ? saved : k)) : [...kits, saved];
		editing = null;
	}

	async function activate(id: number | null) {
		error = '';
		try {
			await setActiveKit(id);
			kits = kits.map((k) => ({ ...k, active: k.id === id }));
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not change the active kit';
		}
	}

	async function handleDelete() {
		if (!toDelete) return;
		await deleteKit(toDelete.id);
		const id = toDelete.id;
		kits = kits.filter((k) => k.id !== id);
	}
</script>

<svelte:head><title>My kit — Drum Trainer</title></svelte:head>

<PageHeader
	title="My kit"
	subtitle="Tell Drum Trainer which instrument each pad of your e-drum kit plays, so your hits are scored against the right notes."
>
	{#snippet actions()}
		{#if editing === null && kits.length > 0}
			<Button variant="primary" onclick={() => (editing = 'new')}><Plus size={16} aria-hidden="true" /> Add kit</Button>
		{/if}
	{/snippet}
</PageHeader>

<Notice>
	Most kits send standard General MIDI notes and work without any setup. Set up your kit if some pads show up as
	the wrong instrument or as "not mapped" in the pad check on the practice page.
</Notice>

{#if error}<Notice variant="error">{error}</Notice>{/if}

{#if editing === 'new'}
	<Card title="New kit">
		<KitEditor onSaved={handleSaved} onCancel={() => (editing = null)} />
	</Card>
{/if}

{#if kits.length === 0 && editing === null}
	<EmptyState title="No kit set up" description="Your hits are scored with the standard General MIDI drum notes.">
		{#snippet icon()}<Drum size={32} aria-hidden="true" />{/snippet}
		{#snippet action()}
			<Button variant="primary" onclick={() => (editing = 'new')}><Plus size={16} aria-hidden="true" /> Set up my kit</Button>
		{/snippet}
	</EmptyState>
{:else}
	{#each kits as kit (kit.id)}
		{#if editing === kit.id}
			<Card title="Edit {kit.name}">
				<KitEditor {kit} onSaved={handleSaved} onCancel={() => (editing = null)} />
			</Card>
		{:else}
			<section class="bg-card border rounded-xl p-5 flex flex-wrap items-center gap-4 {kit.active ? 'border-accent' : 'border-line'}">
				<div class="flex-1 min-w-[12rem]">
					<h2 class="font-semibold text-fg flex items-center gap-2">
						{kit.name}
						{#if kit.active}
							<span class="inline-flex items-center gap-1 text-xs font-medium text-accent-text bg-accent-soft rounded-full px-2 py-0.5">
								<CircleCheck size={12} aria-hidden="true" /> In use
							</span>
						{/if}
					</h2>
					<p class="text-sm text-fg-muted mt-0.5">
						{kit.mappings.length} {kit.mappings.length === 1 ? 'pad' : 'pads'} mapped
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					{#if !kit.active}
						<Button size="sm" onclick={() => activate(kit.id)}>Use this kit</Button>
					{/if}
					<Button size="sm" variant="ghost" onclick={() => (editing = kit.id)} disabled={editing !== null}>
						<Pencil size={14} aria-hidden="true" /> Edit
					</Button>
					<Button
						size="sm"
						variant="ghost"
						onclick={() => {
							toDelete = kit;
							confirmOpen = true;
						}}
						aria-label="Delete {kit.name}"
					>
						<Trash size={14} aria-hidden="true" />
					</Button>
				</div>
			</section>
		{/if}
	{/each}

	{#if activeKit}
		<p class="text-sm text-fg-muted">
			Scoring uses <span class="text-fg font-medium">{activeKit.name}</span>.
			<button type="button" class="text-accent-text hover:underline cursor-pointer" onclick={() => activate(null)}>
				Use standard General MIDI instead
			</button>
		</p>
	{:else if kits.length > 0}
		<p class="text-sm text-fg-muted">No kit in use — scoring uses the standard General MIDI drum notes.</p>
	{/if}
{/if}

<ConfirmDialog
	bind:open={confirmOpen}
	title="Delete “{toDelete?.name}”?"
	message="Your recorded takes are kept. New takes will be scored with {toDelete?.active ? 'the standard General MIDI notes' : 'the kit in use'}."
	confirmLabel="Delete kit"
	danger
	onConfirm={handleDelete}
/>
