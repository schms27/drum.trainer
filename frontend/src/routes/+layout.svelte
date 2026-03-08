<script lang="ts">
	import '../app.css';
	import Header from '$lib/components/layout/Header.svelte';
	import { navigating } from '$app/state';
	import { PUBLIC_APP_VERSION } from '$env/dynamic/public';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: any } = $props();
</script>

{#if data.user}
	{#if navigating.to}
		<div class="nav-loading-track">
			<div class="nav-loading-bar"></div>
		</div>
	{/if}

	<div class="min-h-screen bg-zinc-950 text-white flex flex-col">
		<Header user={data.user} />
		<main class="max-w-5xl w-full mx-auto px-4 py-8 flex-1">
			{@render children()}
		</main>
		<footer class="text-center text-zinc-600 text-xs py-4">
			v{PUBLIC_APP_VERSION}
		</footer>
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
		background: #18181b; /* zinc-900 */
		z-index: 9999;
		overflow: hidden;
	}

	.nav-loading-bar {
		position: absolute;
		height: 100%;
		width: 45%;
		background: #6366f1; /* indigo-500 */
		border-radius: 0 2px 2px 0;
		animation: nav-slide 1.2s ease-in-out infinite;
	}

	@keyframes nav-slide {
		0%   { left: -45%; }
		100% { left: 110%; }
	}
</style>
