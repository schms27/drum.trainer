<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Flag, LibraryBig, Play } from '@lucide/svelte';
	import { copySharedTrack } from '$lib/api/sharing';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import SharePreview from '$lib/components/tracks/SharePreview.svelte';
	import { INSTRUMENTS } from '$lib/drums';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	const { token, shared } = data;

	let busy = $state(false);
	let error = $state('');

	const signedIn = $derived(!!page.data.user);
	const loginHref = `/login?next=${encodeURIComponent(`/s/${token}`)}`;
	const bars = Math.round((shared.preview_ms / 60000) * (shared.bpm / (shared.time_sig_numerator || 4)));
	const instruments = Object.entries(shared.instruments).sort(
		([a], [b]) => INSTRUMENTS.indexOf(a as never) - INSTRUMENTS.indexOf(b as never)
	);

	const sourceLabel: Record<string, string> = {
		original: 'Original groove by the person who shared it',
		licensed: 'Shared with permission from the rights holder',
		public_domain: 'Public-domain music',
		transcription: 'A transcription of an existing song'
	};

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}

	async function addToLibrary() {
		busy = true;
		error = '';
		try {
			const id = await copySharedTrack(token);
			goto(`/tracks/${id}`);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not add the track.';
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>{shared.name} — Drum Trainer</title>
	{#if shared.visibility !== 'public'}<meta name="robots" content="noindex, nofollow" />{/if}
</svelte:head>

<PageHeader
	title={shared.name}
	subtitle="{shared.visibility === 'public' ? 'Public' : 'Shared'} drum track · {formatDuration(shared.duration_ms)} · {Math.round(shared.bpm)} BPM · {shared.note_count} notes"
	back={shared.visibility === 'public' ? { href: '/explore', label: 'Explore' } : undefined}
>
	{#snippet actions()}
		{#if shared.is_owner}
			<Button href="/" variant="secondary" size="lg">This is your track</Button>
		{:else if shared.copied_track_id !== null}
			<Button href="/session/{shared.copied_track_id}" variant="primary" size="lg">
				<Play size={18} aria-hidden="true" /> Practice your copy
			</Button>
		{:else if signedIn}
			<Button variant="primary" size="lg" onclick={addToLibrary} disabled={busy}>
				<LibraryBig size={18} aria-hidden="true" /> Add to my library
			</Button>
		{:else}
			<Button href={loginHref} variant="primary" size="lg">
				<LibraryBig size={18} aria-hidden="true" /> Sign in to practise
			</Button>
		{/if}
	{/snippet}
</PageHeader>

{#if error}<p class="text-sm text-bad -mt-4" role="alert">{error}</p>{/if}

<Card title="Opening {bars} {bars === 1 ? 'bar' : 'bars'}" description="Add the track to your library to practise all of it.">
	<SharePreview
		notes={shared.preview_notes}
		durationMs={shared.preview_ms}
		bpm={shared.bpm}
		timeSigNumerator={shared.time_sig_numerator}
	/>
</Card>

<div class="grid gap-4 md:grid-cols-2">
	<Card title="Instruments">
		<ul class="flex flex-col gap-1.5 text-sm">
			{#each instruments as [name, count] (name)}
				<li class="flex justify-between gap-4">
					<span class="text-fg">{name}</span><span class="tabular-nums text-fg-muted">{count}</span>
				</li>
			{/each}
		</ul>
	</Card>

	<Card title="About this track">
		<p class="text-sm text-fg">{sourceLabel[shared.rights_basis] ?? 'Shared by a Drum Trainer user'}.</p>
		<p class="text-xs text-fg-muted mt-2">
			Drum Trainer users share tracks under their own responsibility. Adding it gives you your own copy of the drum
			notes; backing audio isn’t shared.
		</p>
		<a href="/report?token={encodeURIComponent(token)}" class="inline-flex items-center gap-1.5 text-xs text-fg-subtle hover:text-fg mt-4">
			<Flag size={13} aria-hidden="true" /> Report a copyright problem
		</a>
	</Card>
</div>
