<script lang="ts">
	import { X } from '@lucide/svelte';
	import { chipStyle } from '$lib/tag-colors';
	import type { TagRef } from '$lib/types/track';

	let { tag, onRemove }: { tag: Pick<TagRef, 'name' | 'color'>; onRemove?: () => void } = $props();
</script>

<span
	class="inline-flex items-center gap-1 max-w-full rounded-full border px-2 py-0.5 text-xs font-medium"
	style={chipStyle(tag.color)}
>
	<span class="truncate">{tag.name}</span>
	{#if onRemove}
		<button
			type="button"
			onclick={onRemove}
			class="-mr-1 rounded-full p-0.5 hover:bg-black/10 cursor-pointer"
			aria-label="Remove tag {tag.name}"
		>
			<X size={12} aria-hidden="true" />
		</button>
	{/if}
</span>
