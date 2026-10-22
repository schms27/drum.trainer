<script lang="ts">
	// A thin bar while the browser reports no network. Recording still works
	// offline; signing in, the library and saving takes don't.
	import { WifiOff } from '@lucide/svelte';
	import { focusMode } from '$lib/focus.svelte';

	let online = $state(typeof navigator === 'undefined' ? true : navigator.onLine);
</script>

<svelte:window ononline={() => (online = true)} onoffline={() => (online = false)} />

{#if !online && !focusMode.active}
	<div class="bg-warn-soft text-fg text-sm border-b border-line" role="status">
		<div class="max-w-5xl mx-auto px-4 py-2 flex items-center gap-2">
			<WifiOff size={16} class="shrink-0 text-warn" aria-hidden="true" />
			<span>You're offline — signing in, the library and saving takes need a connection.</span>
		</div>
	</div>
{/if}
