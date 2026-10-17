<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		actions,
		class: extra = '',
		children
	}: {
		title?: string;
		description?: string;
		/** Rendered at the right of the title row. */
		actions?: Snippet;
		class?: string;
		children: Snippet;
	} = $props();
</script>

<section class="bg-card border border-line rounded-xl p-5 {extra}">
	{#if title || actions}
		<div class="flex flex-wrap items-start justify-between gap-x-4 gap-y-2 mb-4">
			<div class="min-w-0">
				{#if title}<h2 class="text-sm font-semibold text-fg">{title}</h2>{/if}
				{#if description}<p class="text-xs text-fg-muted mt-1">{description}</p>{/if}
			</div>
			{#if actions}<div class="flex items-center gap-2 shrink-0">{@render actions()}</div>{/if}
		</div>
	{/if}
	{@render children()}
</section>
