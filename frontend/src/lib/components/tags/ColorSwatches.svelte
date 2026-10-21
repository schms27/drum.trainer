<script lang="ts">
	import { Check } from '@lucide/svelte';
	import { TAG_COLOR_KEYS, tagHue } from '$lib/tag-colors';

	let { value = $bindable('teal'), label = 'Colour' }: { value?: string; label?: string } = $props();
</script>

<div class="flex flex-wrap gap-1.5" role="radiogroup" aria-label={label}>
	{#each TAG_COLOR_KEYS as c (c)}
		<button
			type="button"
			role="radio"
			aria-checked={value === c}
			aria-label={c}
			title={c}
			onclick={() => (value = c)}
			class="w-6 h-6 rounded-full flex items-center justify-center cursor-pointer ring-offset-2 ring-offset-[var(--card)] {value === c ? 'ring-2 ring-[var(--fg)]' : ''}"
			style="background: {tagHue(c)}"
		>
			{#if value === c}<Check size={14} class="text-white" aria-hidden="true" />{/if}
		</button>
	{/each}
</div>
