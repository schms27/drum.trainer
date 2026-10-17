<script lang="ts">
	import '../app.css';
	import '@fontsource/outfit/700.css';
	import type { Snippet } from 'svelte';
	import Header from '$lib/components/layout/Header.svelte';
	import { navigating } from '$app/state';
	import type { LayoutData } from './$types';
	import '$lib/theme.svelte';
	import { focusMode } from '$lib/focus.svelte';

	const appVersion = import.meta.env.VITE_APP_VERSION ?? 'dev';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

{#if data.user}
	{#if navigating.to}
		<div class="nav-loading-track" aria-hidden="true">
			<div class="nav-loading-bar"></div>
		</div>
	{/if}

	<div class="min-h-screen bg-canvas text-fg flex flex-col">
		<!-- Focus mode (during a take) hides the page chrome. -->
		{#if !focusMode.active}<Header user={data.user} />{/if}
		<main class="max-w-5xl w-full mx-auto px-4 py-8 flex-1 flex flex-col gap-8">
			{@render children()}
		</main>
		{#if !focusMode.active}
			<footer class="text-center text-fg-subtle text-xs py-4">Drum Trainer · version {appVersion}</footer>
		{/if}
	</div>
{:else}
	{@render children()}
{/if}

<style>
	.nav-loading-track {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: var(--panel);
		z-index: 9999;
		overflow: hidden;
	}

	.nav-loading-bar {
		position: absolute;
		height: 100%;
		width: 45%;
		background: var(--accent);
		border-radius: 0 2px 2px 0;
		animation: nav-slide 1.2s ease-in-out infinite;
	}

	@keyframes nav-slide {
		0% {
			left: -45%;
		}
		100% {
			left: 110%;
		}
	}
</style>
