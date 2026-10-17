<script lang="ts">
	import { page } from '$app/state';
	import { Check, Copy, Globe, Link2, RefreshCw, Settings2 } from '@lucide/svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import {
		LISTABLE_RIGHTS,
		RIGHTS_LABELS,
		rotateShareLink,
		shareTrack,
		shareUrl,
		unshareTrack
	} from '$lib/api/sharing';
	import type { RightsBasis, SharedVisibility, Track } from '$lib/types/track';

	let {
		open = $bindable(false),
		track = $bindable()
	}: {
		open?: boolean;
		track: Track;
	} = $props();

	let dialog: HTMLDialogElement;
	// The settings form shows for a private track, or when changing a shared one.
	let editing = $state(false);
	let rights = $state<RightsBasis | null>(null);
	let visibility = $state<SharedVisibility>('unlisted');
	let attested = $state(false);
	let busy = $state(false);
	let error = $state('');
	let copied = $state(false);

	const shared = $derived(!!track.share_token);
	const link = $derived(track.share_token ? shareUrl(track.share_token) : '');
	const rightsOptions = Object.entries(RIGHTS_LABELS) as [RightsBasis, string][];
	const canList = $derived(rights !== null && LISTABLE_RIGHTS.includes(rights));

	// A transcription can't be public: fall back to link-only when it's picked.
	$effect(() => {
		if (!canList && visibility === 'public') visibility = 'unlisted';
	});

	$effect(() => {
		if (open && !dialog.open) {
			error = '';
			copied = false;
			resetForm();
			editing = !shared;
			dialog.showModal();
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});

	function resetForm() {
		rights = track.rights_basis;
		visibility = track.visibility === 'public' ? 'public' : 'unlisted';
		attested = false;
	}

	async function run(action: () => Promise<void>) {
		busy = true;
		error = '';
		try {
			await action();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	const save = () =>
		run(async () => {
			if (!rights) return;
			const res = await shareTrack(track.id, rights, visibility);
			track = { ...track, visibility: res.visibility, share_token: res.share_token, rights_basis: rights };
			editing = false;
		});

	const resetLink = () =>
		run(async () => {
			const res = await rotateShareLink(track.id);
			track = { ...track, share_token: res.share_token };
			copied = false;
		});

	const stopSharing = () =>
		run(async () => {
			await unshareTrack(track.id);
			track = { ...track, visibility: 'private', share_token: null };
			resetForm();
			editing = true;
		});

	async function copyLink() {
		try {
			await navigator.clipboard.writeText(link);
			copied = true;
		} catch {
			error = 'Could not copy — select the link and copy it manually.';
		}
	}
</script>

<dialog
	bind:this={dialog}
	onclose={() => (open = false)}
	aria-labelledby="share-title"
	class="m-auto w-[calc(100%-2rem)] max-w-lg rounded-xl border border-line bg-card text-fg p-0 shadow-xl backdrop:bg-black/50"
>
	<div class="p-6 flex flex-col gap-5">
		<div>
			<h2 id="share-title" class="text-lg font-semibold">Share “{track.name}”</h2>
			<p class="text-sm text-fg-muted mt-1">
				People who open the link can see the track and add a copy to their own library.
			</p>
		</div>

		{#if !shared && page.data.user?.sharing_banned}
			<Notice variant="error">
				Sharing is turned off for your account because several tracks you shared were removed after copyright reports.
				See our <a href="/copyright" class="text-accent-text hover:underline">copyright policy</a>.
			</Notice>
			<div class="flex justify-end">
				<Button variant="ghost" onclick={() => (open = false)}>Close</Button>
			</div>
		{:else if editing}
			<fieldset class="flex flex-col gap-2">
				<legend class="text-sm font-medium mb-2">Where does this drum part come from?</legend>
				{#each rightsOptions as [value, label] (value)}
					<label class="flex items-center gap-3 rounded-lg border border-line px-3 py-2.5 text-sm cursor-pointer has-checked:border-accent has-checked:bg-accent-soft">
						<input type="radio" name="rights" {value} bind:group={rights} class="accent-(--accent)" />
						{label}
					</label>
				{/each}
			</fieldset>

			{#if rights === 'transcription'}
				<Notice variant="warning">
					Sharing a transcription of a copyrighted song needs permission from the rights holder, even with friends. If
					you don’t have permission, keep the track private. Transcriptions can’t be listed in Explore.
				</Notice>
			{/if}

			<fieldset class="flex flex-col gap-2">
				<legend class="text-sm font-medium mb-2">Who can find it?</legend>
				<label class="flex items-start gap-3 rounded-lg border border-line px-3 py-2.5 text-sm cursor-pointer has-checked:border-accent has-checked:bg-accent-soft">
					<input type="radio" name="visibility" value="unlisted" bind:group={visibility} class="mt-0.5 accent-(--accent)" />
					<span><span class="font-medium">Only people with the link</span><br /><span class="text-fg-muted">Not listed or searchable anywhere.</span></span>
				</label>
				<label
					class="flex items-start gap-3 rounded-lg border border-line px-3 py-2.5 text-sm has-checked:border-accent has-checked:bg-accent-soft {canList
						? 'cursor-pointer'
						: 'opacity-50 cursor-not-allowed'}"
				>
					<input type="radio" name="visibility" value="public" bind:group={visibility} disabled={!canList} class="mt-0.5 accent-(--accent)" />
					<span><span class="font-medium">Everyone</span><br /><span class="text-fg-muted">Listed in Explore, where any drummer can find it.</span></span>
				</label>
			</fieldset>

			<p class="text-xs text-fg-muted">
				Only the drum notes are shared. Your backing tracks, the original MIDI file and your practice history stay
				private.
			</p>

			<label class="flex items-start gap-3 text-sm cursor-pointer">
				<input type="checkbox" bind:checked={attested} class="mt-0.5 accent-(--accent)" />
				<span>
					I have the right to share this track, and I’m responsible for what I share under the
					<a href="/terms" target="_blank" class="text-accent-text hover:underline">terms of use</a>.
				</span>
			</label>

			<div class="flex justify-end gap-2">
				<Button
					variant="ghost"
					onclick={() => (shared ? ((editing = false), resetForm()) : (open = false))}
					disabled={busy}>Cancel</Button
				>
				<Button variant="primary" onclick={save} disabled={busy || !rights || !attested}>
					{#if shared}
						Save
					{:else if visibility === 'public'}
						<Globe size={16} aria-hidden="true" /> Publish
					{:else}
						<Link2 size={16} aria-hidden="true" /> Create link
					{/if}
				</Button>
			</div>
		{:else}
			<div class="flex flex-col gap-2">
				<label for="share-link" class="text-xs font-medium text-fg-muted">Share link</label>
				<div class="flex gap-2">
					<input
						id="share-link"
						readonly
						value={link}
						onfocus={(e) => e.currentTarget.select()}
						class="flex-1 min-w-0 h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
					/>
					<Button variant="primary" onclick={copyLink} aria-label="Copy link">
						{#if copied}<Check size={16} aria-hidden="true" /> Copied{:else}<Copy size={16} aria-hidden="true" /> Copy{/if}
					</Button>
				</div>
			</div>
			<p class="flex items-center gap-2 text-sm">
				{#if track.visibility === 'public'}
					<Globe size={16} class="text-accent-text" aria-hidden="true" /> Listed in
					<a href="/explore" class="text-accent-text hover:underline">Explore</a>
				{:else}
					<Link2 size={16} class="text-fg-muted" aria-hidden="true" /> Only people with the link
				{/if}
			</p>
			<p class="text-xs text-fg-subtle">
				Reset the link to stop the current one from working. People who already added a copy keep it.
			</p>
			<div class="flex flex-wrap justify-between gap-2">
				<Button variant="ghost" onclick={stopSharing} disabled={busy} class="text-bad">Stop sharing</Button>
				<div class="flex flex-wrap gap-2">
					<Button variant="secondary" onclick={() => (editing = true)} disabled={busy}>
						<Settings2 size={16} aria-hidden="true" /> Change
					</Button>
					<Button variant="secondary" onclick={resetLink} disabled={busy}>
						<RefreshCw size={16} aria-hidden="true" /> Reset link
					</Button>
					<Button variant="ghost" onclick={() => (open = false)}>Done</Button>
				</div>
			</div>
		{/if}

		{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}
	</div>
</dialog>
