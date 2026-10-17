<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { ExternalLink } from '@lucide/svelte';
	import { disableTrack, rejectReport, restoreTrack, type AdminReport } from '$lib/api/reports';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const tabs = [
		{ status: 'open', label: 'Open' },
		{ status: 'actioned', label: 'Removed' },
		{ status: 'rejected', label: 'Rejected' }
	] as const;

	const kindLabel = { copyright: 'Copyright', abuse: 'Abuse', other: 'Other' } as const;

	// The one report whose action form is open, and that form's fields.
	let active = $state<{ id: number; action: 'remove' | 'reject' | 'restore' } | null>(null);
	let text = $state('');
	let flag = $state(true);
	let busy = $state(false);
	let error = $state('');
	let message = $state('');

	function open(r: AdminReport, action: 'remove' | 'reject' | 'restore') {
		active = { id: r.id, action };
		error = '';
		if (action === 'remove') {
			text = r.kind === 'copyright' && r.work_identified
				? `Removed after a copyright notice regarding “${r.work_identified}”.`
				: 'Removed for breaking the terms of use.';
			flag = r.kind === 'copyright';
		} else {
			text = '';
			flag = action === 'restore';
		}
	}

	async function confirm(r: AdminReport) {
		if (!active) return;
		busy = true;
		error = '';
		try {
			if (active.action === 'remove') {
				const res = await disableTrack(r.track_id, text, flag);
				message = `Removed “${r.track_name}” and its copies (${res.tracks_affected} ${res.tracks_affected === 1 ? 'track' : 'tracks'}). Owner has ${res.owner_strikes} ${res.owner_strikes === 1 ? 'strike' : 'strikes'}${res.owner_banned ? ' and can no longer share' : ''}.`;
			} else if (active.action === 'reject') {
				await rejectReport(r.id, text);
				message = `Rejected the report about “${r.track_name}”.`;
			} else {
				const res = await restoreTrack(r.track_id, text, flag);
				message = `Restored “${r.track_name}” (${res.tracks_affected} ${res.tracks_affected === 1 ? 'track' : 'tracks'}).`;
			}
			active = null;
			await invalidateAll();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
	}
</script>

<svelte:head><title>Reports — Drum Trainer</title></svelte:head>

<PageHeader title="Reports" subtitle="Copyright and abuse reports about shared tracks" />

<nav class="flex gap-1 border-b border-line -mt-2" aria-label="Report status">
	{#each tabs as tab (tab.status)}
		<a
			href="?status={tab.status}"
			aria-current={data.status === tab.status ? 'page' : undefined}
			class="px-3 py-2 text-sm -mb-px border-b-2 {data.status === tab.status
				? 'border-accent text-fg font-medium'
				: 'border-transparent text-fg-muted hover:text-fg'}">{tab.label}</a
		>
	{/each}
</nav>

{#if message}<Notice>{message}</Notice>{/if}

{#if data.reports.length === 0}
	<EmptyState title={data.status === 'open' ? 'No open reports' : 'Nothing here yet'} />
{:else}
	<div class="flex flex-col gap-4">
		{#each data.reports as r (r.id)}
			<Card>
				<div class="flex flex-col gap-4">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="text-xs text-fg-muted">
								#{r.id} · {kindLabel[r.kind]} · {formatDate(r.created_at)}
							</p>
							<h2 class="font-semibold text-fg mt-0.5 break-words">{r.track_name}</h2>
						</div>
						<div class="flex flex-wrap gap-1.5 text-xs">
							{#if !r.track_exists}<span class="rounded bg-control px-2 py-0.5">Deleted by owner</span>{/if}
							{#if r.track_disabled}<span class="rounded bg-bad-soft px-2 py-0.5">Removed</span>{/if}
							{#if r.copies > 0}<span class="rounded bg-control px-2 py-0.5">{r.copies} {r.copies === 1 ? 'copy' : 'copies'}</span>{/if}
						</div>
					</div>

					<dl class="grid gap-x-6 gap-y-2 sm:grid-cols-2 text-sm">
						<div>
							<dt class="text-xs text-fg-muted">Reported by</dt>
							<dd>{r.reporter_name} · <a href="mailto:{r.reporter_email}" class="text-accent-text hover:underline">{r.reporter_email}</a></dd>
						</div>
						<div>
							<dt class="text-xs text-fg-muted">Shared by</dt>
							<dd>
								{r.owner_name ?? 'Deleted account'}{#if r.owner_email}&nbsp;· {r.owner_email}{/if}
								<span class="text-fg-muted">· {r.owner_strikes} {r.owner_strikes === 1 ? 'strike' : 'strikes'}{r.owner_banned ? ', banned' : ''}</span>
							</dd>
						</div>
						{#if r.work_identified}
							<div>
								<dt class="text-xs text-fg-muted">Work named</dt>
								<dd>{r.work_identified}</dd>
							</div>
						{/if}
						<div>
							<dt class="text-xs text-fg-muted">Owner declared</dt>
							<dd>{r.rights_basis?.replace('_', ' ') ?? '—'}</dd>
						</div>
						<div class="sm:col-span-2">
							<dt class="text-xs text-fg-muted">Details</dt>
							<dd class="whitespace-pre-wrap break-words">{r.details}</dd>
						</div>
						{#if r.kind === 'copyright'}
							<div class="sm:col-span-2 text-xs text-fg-muted">
								Good-faith statement {r.good_faith ? 'confirmed' : 'not confirmed'}
							</div>
						{/if}
						{#if r.resolution_note}
							<div class="sm:col-span-2">
								<dt class="text-xs text-fg-muted">Resolution{r.resolved_at ? ` · ${formatDate(r.resolved_at)}` : ''}</dt>
								<dd>{r.resolution_note}</dd>
							</div>
						{/if}
					</dl>

					{#if active?.id === r.id}
						<div class="flex flex-col gap-3 border-t border-line pt-4">
							<label class="flex flex-col gap-1.5 text-sm">
								<span class="font-medium">
									{active.action === 'remove' ? 'Reason (shown to the owner and everyone with a copy)' : 'Note (internal)'}
								</span>
								<textarea bind:value={text} rows="2" maxlength="500" class="w-full rounded-lg border border-line bg-canvas px-3 py-2 text-sm"></textarea>
							</label>
							{#if active.action !== 'reject'}
								<label class="flex items-center gap-2 text-sm">
									<input type="checkbox" bind:checked={flag} class="accent-(--accent)" />
									{active.action === 'remove' ? 'Give the owner a strike' : 'Take back the strike from the removal'}
								</label>
							{/if}
							{#if error}<p class="text-sm text-bad" role="alert">{error}</p>{/if}
							<div class="flex justify-end gap-2">
								<Button variant="ghost" onclick={() => (active = null)} disabled={busy}>Cancel</Button>
								<Button
									variant={active.action === 'remove' ? 'danger' : 'primary'}
									onclick={() => confirm(r)}
									disabled={busy || (active.action === 'remove' && !text.trim())}
								>
									{active.action === 'remove' ? 'Remove track and copies' : active.action === 'reject' ? 'Reject report' : 'Restore track'}
								</Button>
							</div>
						</div>
					{:else}
						<div class="flex flex-wrap items-center justify-between gap-2 border-t border-line pt-4">
							{#if r.status === 'open' && !r.track_disabled}
								<a href="/s/{r.share_token}" target="_blank" class="inline-flex items-center gap-1 text-sm text-accent-text hover:underline">
									View shared page <ExternalLink size={14} aria-hidden="true" />
								</a>
							{:else}
								<span></span>
							{/if}
							<div class="flex gap-2">
								{#if r.status === 'open'}
									<Button variant="ghost" onclick={() => open(r, 'reject')}>Reject</Button>
									<Button variant="danger" onclick={() => open(r, 'remove')}>Remove track</Button>
								{:else if r.status === 'actioned' && r.track_disabled}
									<Button variant="secondary" onclick={() => open(r, 'restore')}>Restore track</Button>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			</Card>
		{/each}
	</div>
{/if}
