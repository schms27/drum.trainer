<script lang="ts">
	import { page } from '$app/state';

	// History and Stats are two views of the "Progress" section until they are
	// merged into a single page (see docs/ui-overhaul-plan.md, phase 3).
	const tabs = [
		{ href: '/history', label: 'Sessions' },
		{ href: '/stats', label: 'Practice time' }
	];
</script>

<nav class="flex gap-1 border-b border-line" aria-label="Progress">
	{#each tabs as tab (tab.href)}
		{@const active = page.url.pathname.startsWith(tab.href)}
		<a
			href={tab.href}
			aria-current={active ? 'page' : undefined}
			class="px-3 h-10 flex items-center text-sm font-medium border-b-2 -mb-px transition-colors
				{active ? 'border-accent text-fg' : 'border-transparent text-fg-muted hover:text-fg'}"
		>
			{tab.label}
		</a>
	{/each}
</nav>
