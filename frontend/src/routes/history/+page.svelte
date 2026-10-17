<script lang="ts">
	import { ChevronRight, ListMusic } from '@lucide/svelte';
	import { listSessions } from '$lib/api/sessions';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import ProgressTabs from '$lib/components/layout/ProgressTabs.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { formatOffset, formatPercent, hitRateTone, timingTone, toneText } from '$lib/score-format';

	const { data } = $props();
	// svelte-ignore state_referenced_locally
	const { trackNames } = data;
	// svelte-ignore state_referenced_locally
	let sessions = $state(data.sessions);
	// svelte-ignore state_referenced_locally
	let hasMore = $state(data.hasMore);
	let loadingMore = $state(false);

	async function loadMore() {
		loadingMore = true;
		try {
			const oldest = sessions[sessions.length - 1];
			const page = await listSessions({ before: oldest.id });
			sessions = [...sessions, ...page.sessions];
			hasMore = page.hasMore;
		} finally {
			loadingMore = false;
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
	}
</script>

<svelte:head><title>Progress — Drum Trainer</title></svelte:head>

<PageHeader title="Progress" subtitle="Every take you have recorded, newest first." />
<ProgressTabs />

{#if sessions.length === 0}
	<EmptyState
		title="No takes recorded yet"
		description="Pick a track in your library, press Start and play along. Your results will show up here."
	>
		{#snippet icon()}<ListMusic size={32} aria-hidden="true" />{/snippet}
		{#snippet action()}<Button href="/" variant="primary">Go to library</Button>{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each sessions as session (session.id)}
			<li>
				<a
					href="/results/{session.id}"
					class="group flex items-center gap-4 bg-card border border-line rounded-xl px-4 py-3 hover:border-accent transition-colors"
				>
					<div class="min-w-0 flex-1">
						<p class="font-medium text-fg truncate">{trackNames.get(session.track_id) ?? 'Deleted track'}</p>
						<p class="text-xs text-fg-subtle mt-0.5">{formatDate(session.created_at)}</p>
					</div>
					<dl class="flex items-center gap-4 sm:gap-8 text-right shrink-0">
						<div>
							<dt class="text-xs text-fg-subtle">Notes hit</dt>
							<dd class="font-semibold tabular-nums {toneText[hitRateTone(session.hit_rate)]}">
								{formatPercent(session.hit_rate)}
								<span class="hidden sm:inline text-xs font-normal text-fg-subtle"
									>{session.matched_count}/{session.target_count}</span
								>
							</dd>
						</div>
						<div class="hidden sm:block">
							<dt class="text-xs text-fg-subtle">Average timing</dt>
							<dd class="font-semibold tabular-nums {toneText[timingTone(session.avg_offset_ms)]}">
								{formatOffset(session.avg_offset_ms)}
							</dd>
						</div>
					</dl>
					<ChevronRight size={18} class="text-fg-subtle group-hover:text-accent-text shrink-0" aria-hidden="true" />
				</a>
			</li>
		{/each}
	</ul>

	{#if hasMore}
		<div class="text-center">
			<Button onclick={loadMore} disabled={loadingMore}>{loadingMore ? 'Loading…' : 'Show older takes'}</Button>
		</div>
	{/if}
{/if}
