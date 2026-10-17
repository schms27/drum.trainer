<script lang="ts">
	import { Check, Copy, Link2, RefreshCw } from '@lucide/svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import { RIGHTS_LABELS, rotateShareLink, shareTrack, shareUrl, unshareTrack } from '$lib/api/sharing';
	import type { RightsBasis, Track } from '$lib/types/track';

	let {
		open = $bindable(false),
		track = $bindable()
	}: {
		open?: boolean;
		track: Track;
	} = $props();

	let dialog: HTMLDialogElement;
	let rights = $state<RightsBasis | null>(null);
	let attested = $state(false);
	let busy = $state(false);
	let error = $state('');
	let copied = $state(false);

	const link = $derived(track.share_token ? shareUrl(track.share_token) : '');
	const rightsOptions = Object.entries(RIGHTS_LABELS) as [RightsBasis, string][];

	$effect(() => {
		if (open && !dialog.open) {
			error = '';
			copied = false;
			rights = track.rights_basis;
			attested = false;
			dialog.showModal();
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});

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

	const createLink = () =>
		run(async () => {
			if (!rights) return;
			const res = await shareTrack(track.id, rights);
			track = { ...track, visibility: res.visibility, share_token: res.share_token, rights_basis: rights };
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
			attested = false;
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
				Anyone with the link can see the track and add a copy to their own library. The link isn’t listed or
				searchable anywhere.
			</p>
		</div>

		{#if track.share_token}
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
			<p class="text-xs text-fg-subtle">
				Reset the link to stop the current one from working. People who already added a copy keep it.
			</p>
			<div class="flex flex-wrap justify-between gap-2">
				<Button variant="ghost" onclick={stopSharing} disabled={busy} class="text-bad">Stop sharing</Button>
				<div class="flex gap-2">
					<Button variant="secondary" onclick={resetLink} disabled={busy}>
						<RefreshCw size={16} aria-hidden="true" /> Reset link
					</Button>
					<Button variant="ghost" onclick={() => (open = false)}>Done</Button>
				</div>
			</div>
		{:else}
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
					you don’t have permission, keep the track private.
				</Notice>
			{/if}

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
				<Button variant="ghost" onclick={() => (open = false)} disabled={busy}>Cancel</Button>
				<Button variant="primary" onclick={createLink} disabled={busy || !rights || !attested}>
					<Link2 size={16} aria-hidden="true" /> Create link
				</Button>
			</div>
		{/if}

		{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}
	</div>
</dialog>
