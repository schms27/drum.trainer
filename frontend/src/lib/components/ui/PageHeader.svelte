<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronLeft } from '@lucide/svelte';

	let {
		title,
		subtitle,
		back,
		actions
	}: {
		title: string;
		subtitle?: string;
		/** Optional back link shown above the title. */
		back?: { href: string; label: string };
		actions?: Snippet;
	} = $props();
</script>

<div class="flex flex-col gap-2">
	{#if back}
		<a
			href={back.href}
			class="inline-flex items-center gap-1 text-sm text-fg-muted hover:text-fg transition-colors w-fit"
		>
			<ChevronLeft size={16} aria-hidden="true" />{back.label}
		</a>
	{/if}
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="min-w-0">
			<h1 class="font-brand text-[1.65rem] leading-tight font-bold text-fg tracking-tight break-words">{title}</h1>
			{#if subtitle}<p class="text-sm text-fg-muted mt-1">{subtitle}</p>{/if}
		</div>
		{#if actions}<div class="flex flex-wrap items-center gap-2">{@render actions()}</div>{/if}
	</div>
</div>
