<script lang="ts">
	import '../app.css';
	import Header from '$lib/components/layout/Header.svelte';

	let { children } = $props();
</script>

<div class="min-h-screen bg-zinc-950 text-white">
	<Header />
	<main class="max-w-5xl mx-auto px-4 py-8">
		{@render children()}
	</main>
</div>
