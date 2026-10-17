<script lang="ts">
	import { Monitor, Moon, Sun } from '@lucide/svelte';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	const options: { value: ThemePreference; label: string }[] = [
		{ value: 'system', label: 'Use system theme' },
		{ value: 'light', label: 'Light theme' },
		{ value: 'dark', label: 'Dark theme' }
	];
</script>

<div class="inline-flex rounded-lg bg-control p-0.5" role="radiogroup" aria-label="Color theme">
	{#each options as opt (opt.value)}
		{@const active = theme.preference === opt.value}
		<button
			type="button"
			role="radio"
			aria-checked={active}
			aria-label={opt.label}
			title={opt.label}
			onclick={() => theme.set(opt.value)}
			class="w-8 h-7 flex items-center justify-center rounded-md transition-colors cursor-pointer
				{active ? 'bg-card text-fg shadow-sm' : 'text-fg-subtle hover:text-fg'}"
		>
			{#if opt.value === 'system'}
				<Monitor size={15} aria-hidden="true" />
			{:else if opt.value === 'light'}
				<Sun size={15} aria-hidden="true" />
			{:else}
				<Moon size={15} aria-hidden="true" />
			{/if}
		</button>
	{/each}
</div>
