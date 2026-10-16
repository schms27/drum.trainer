<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Info, TriangleAlert, CircleAlert } from '@lucide/svelte';

	let {
		variant = 'info',
		children
	}: {
		variant?: 'info' | 'warning' | 'error';
		children: Snippet;
	} = $props();

	const styles = {
		info: 'bg-accent-soft',
		warning: 'bg-warn-soft',
		error: 'bg-bad-soft'
	};
	const iconColor = { info: 'text-accent-text', warning: 'text-warn', error: 'text-bad' };
</script>

<div
	class="flex gap-3 rounded-lg px-4 py-3 text-sm text-fg {styles[variant]}"
	role={variant === 'error' ? 'alert' : 'status'}
>
	<span class="shrink-0 mt-0.5 {iconColor[variant]}">
		{#if variant === 'info'}
			<Info size={16} aria-hidden="true" />
		{:else if variant === 'warning'}
			<TriangleAlert size={16} aria-hidden="true" />
		{:else}
			<CircleAlert size={16} aria-hidden="true" />
		{/if}
	</span>
	<div class="min-w-0">{@render children()}</div>
</div>
