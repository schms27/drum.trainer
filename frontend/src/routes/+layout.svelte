<script lang="ts">
	import '../app.css';
	import Header from '$lib/components/layout/Header.svelte';
	import { navigating } from '$app/state';

	let { children } = $props();
</script>

{#if navigating}
	<div class="nav-loading-track">
		<div class="nav-loading-bar"></div>
	</div>
{/if}

<div class="min-h-screen bg-zinc-950 text-white">
	<Header />
	<main class="max-w-5xl mx-auto px-4 py-8">
		{@render children()}
	</main>
</div>

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
