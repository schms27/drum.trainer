<script lang="ts">
	import { goto } from '$app/navigation';
	import { logout, type AuthUser } from '$lib/api/auth';

	let { user }: { user: AuthUser | null } = $props();

	async function handleLogout() {
		await logout();
		goto('/login');
	}

	function initials(name: string): string {
		return name
			.split(' ')
			.map((p) => p[0])
			.join('')
			.slice(0, 2)
			.toUpperCase();
	}
</script>

<header class="bg-zinc-900 border-b border-zinc-700 px-6 py-4 flex items-center gap-4">
	<a href="/" class="text-white font-bold text-xl tracking-tight hover:text-indigo-400 transition">
		🥁 Drum Trainer
	</a>

	{#if user}
		<nav class="flex gap-4 ml-auto text-sm items-center">
			<a href="/" class="text-zinc-400 hover:text-white transition">Tracks</a>
			<a href="/history" class="text-zinc-400 hover:text-white transition">History</a>
			<a href="/stats" class="text-zinc-400 hover:text-white transition">Stats</a>

			<div class="flex items-center gap-3 ml-2 pl-4 border-l border-zinc-700">
				{#if user.avatar_url}
					<img
						src={user.avatar_url}
						alt={user.name}
						class="w-7 h-7 rounded-full object-cover"
					/>
				{:else}
					<div
						class="w-7 h-7 rounded-full bg-indigo-600 flex items-center justify-center text-xs font-semibold text-white"
					>
						{initials(user.name)}
					</div>
				{/if}
				<span class="text-zinc-300 text-sm">{user.name}</span>
				<button
					onclick={handleLogout}
					class="text-zinc-500 hover:text-white text-sm transition cursor-pointer"
				>
					Sign out
				</button>
			</div>
		</nav>
	{:else}
		<div class="ml-auto">
			<a href="/login" class="text-zinc-400 hover:text-white text-sm transition">Sign in</a>
		</div>
	{/if}
</header>
