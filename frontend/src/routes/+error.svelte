<script lang="ts">
	import { page } from '$app/state';
	import { invalidateAll } from '$app/navigation';
	import { CircleAlert, RotateCcw, WifiOff } from '@lucide/svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';

	const offline = $derived(page.error?.offline === true);
	let retrying = $state(false);

	async function retry() {
		retrying = true;
		try {
			// Rerun every load function; a working connection replaces this page.
			await invalidateAll();
		} finally {
			retrying = false;
		}
	}
</script>

<svelte:head><title>{offline ? 'Offline' : 'Error'} — Drum Trainer</title></svelte:head>

{#if offline}
	<EmptyState title="You're offline" description={page.error?.message}>
		{#snippet icon()}<WifiOff size={32} aria-hidden="true" />{/snippet}
		{#snippet action()}
			<Button onclick={retry} disabled={retrying}>
				<RotateCcw size={16} aria-hidden="true" />
				{retrying ? 'Trying…' : 'Try again'}
			</Button>
		{/snippet}
	</EmptyState>
{:else}
	<EmptyState
		title={page.status === 404 ? 'Page not found' : 'Something went wrong'}
		description={page.status === 404 ? 'There is nothing at this address.' : page.error?.message}
	>
		{#snippet icon()}<CircleAlert size={32} aria-hidden="true" />{/snippet}
		{#snippet action()}<Button href="/">Back to the library</Button>{/snippet}
	</EmptyState>
{/if}
