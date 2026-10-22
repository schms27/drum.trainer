<script lang="ts">
	// Shown when a new version has been deployed while the app is open
	// (SvelteKit polls /_app/version.json, see kit.version in svelte.config.js).
	// Hidden during a take so nothing pops up while drumming.
	import { updated } from '$app/state';
	import { RefreshCw, X } from '@lucide/svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { focusMode } from '$lib/focus.svelte';

	let dismissed = $state(false);
</script>

{#if updated.current && !dismissed && !focusMode.active}
	<div
		class="fixed bottom-4 inset-x-4 sm:left-auto sm:right-4 sm:max-w-sm z-50 flex items-center gap-3 rounded-lg border border-line bg-panel px-4 py-3 text-sm text-fg shadow-lg"
		role="status"
	>
		<span class="flex-1">A new version of Drum Trainer is available.</span>
		<Button variant="primary" size="sm" onclick={() => location.reload()}>
			<RefreshCw size={14} aria-hidden="true" />
			Reload
		</Button>
		<button
			type="button"
			class="shrink-0 rounded p-1 text-fg-muted hover:text-fg hover:bg-control"
			aria-label="Dismiss"
			onclick={() => (dismissed = true)}
		>
			<X size={16} aria-hidden="true" />
		</button>
	</div>
{/if}
