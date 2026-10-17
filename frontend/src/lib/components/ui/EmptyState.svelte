<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		icon,
		action
	}: {
		title: string;
		description?: string;
		icon?: Snippet;
		action?: Snippet;
	} = $props();
</script>

<div
	class="flex flex-col items-center text-center gap-3 border border-dashed border-line-strong rounded-xl px-6 py-10"
>
	{#if icon}<div class="text-fg-subtle">{@render icon()}</div>{/if}
	<div>
		<p class="font-medium text-fg">{title}</p>
		{#if description}<p class="text-sm text-fg-muted mt-1 max-w-md">{description}</p>{/if}
	</div>
	{#if action}<div class="mt-1">{@render action()}</div>{/if}
</div>
