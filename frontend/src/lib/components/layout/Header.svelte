<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Library, ChartLine, LogOut, Menu, X, SlidersHorizontal, Flag, Compass } from '@lucide/svelte';
	import { logout, type AuthUser } from '$lib/api/auth';
	import ThemeToggle from '$lib/components/ui/ThemeToggle.svelte';

	let { user }: { user: AuthUser } = $props();

	let menuOpen = $state(false);
	let menuEl = $state<HTMLDivElement>();

	const navItems = $derived([
		{ href: '/', label: 'Library', icon: Library, match: (p: string) => p === '/' || p.startsWith('/tracks') || p.startsWith('/session') },
		{ href: '/progress', label: 'Progress', icon: ChartLine, match: (p: string) => p.startsWith('/progress') || p.startsWith('/results') },
		{ href: '/explore', label: 'Explore', icon: Compass, match: (p: string) => p.startsWith('/explore') || p.startsWith('/s/') },
		...(user.is_admin
			? [{ href: '/admin/reports', label: 'Reports', icon: Flag, match: (p: string) => p.startsWith('/admin') }]
			: [])
	]);

	const path = $derived(page.url.pathname);

	// Close the menu whenever the route changes.
	$effect(() => {
		path;
		menuOpen = false;
	});

	function onWindowClick(e: MouseEvent) {
		// composedPath() is fixed at dispatch time, so it still includes the menu
		// even when the clicked icon has already been swapped out (☰ ↔ ✕).
		if (menuOpen && menuEl && !e.composedPath().includes(menuEl)) menuOpen = false;
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') menuOpen = false;
	}

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

<svelte:window onclick={onWindowClick} onkeydown={onKeydown} />

{#snippet avatar()}
	{#if user.avatar_url}
		<img src={user.avatar_url} alt="" class="w-8 h-8 rounded-full object-cover" />
	{:else}
		<span
			class="w-8 h-8 rounded-full bg-accent text-accent-fg flex items-center justify-center text-xs font-semibold"
			>{initials(user.name)}</span
		>
	{/if}
{/snippet}

<header class="bg-panel border-b border-line sticky top-0 z-40">
	<div class="max-w-5xl mx-auto px-4 h-14 flex items-center gap-6">
		<a href="/" class="flex items-center gap-2 font-bold text-lg text-fg hover:text-accent-text transition-colors">
			<img src="/logo.svg" alt="" width="32" height="32" class="w-8 h-8" />
			<span class="font-brand text-[1.2rem] tracking-tight">Drum Trainer</span>
		</a>

		<!-- Desktop navigation -->
		<nav class="hidden sm:flex items-center gap-1" aria-label="Main">
			{#each navItems as item (item.href)}
				{@const active = item.match(path)}
				<a
					href={item.href}
					aria-current={active ? 'page' : undefined}
					class="flex items-center gap-2 h-9 px-3 rounded-lg text-sm font-medium transition-colors
						{active ? 'bg-accent-soft text-accent-text' : 'text-fg-muted hover:text-fg hover:bg-control'}"
				>
					<item.icon size={16} aria-hidden="true" />
					{item.label}
				</a>
			{/each}
		</nav>

		<!-- User menu (desktop) / full menu (mobile) -->
		<div class="ml-auto relative" bind:this={menuEl}>
			<button
				type="button"
				onclick={() => (menuOpen = !menuOpen)}
				aria-expanded={menuOpen}
				aria-haspopup="true"
				aria-label="Open menu"
				class="flex items-center gap-2 rounded-full sm:pr-1 cursor-pointer hover:bg-control transition-colors"
			>
				<span class="sm:hidden w-9 h-9 flex items-center justify-center text-fg">
					{#if menuOpen}<X size={22} aria-hidden="true" />{:else}<Menu size={22} aria-hidden="true" />{/if}
				</span>
				<span class="hidden sm:block">{@render avatar()}</span>
			</button>

			{#if menuOpen}
				<div
					class="absolute right-0 mt-2 w-64 rounded-xl border border-line bg-card shadow-xl p-2 flex flex-col gap-1"
				>
					<div class="flex items-center gap-3 px-2 py-2">
						{@render avatar()}
						<div class="min-w-0">
							<p class="text-sm font-medium text-fg truncate">{user.name}</p>
							{#if user.email}<p class="text-xs text-fg-subtle truncate">{user.email}</p>{/if}
						</div>
					</div>

					<nav class="sm:hidden flex flex-col gap-1 border-t border-line pt-1" aria-label="Main">
						{#each navItems as item (item.href)}
							{@const active = item.match(path)}
							<a
								href={item.href}
								aria-current={active ? 'page' : undefined}
								class="flex items-center gap-3 px-2 h-10 rounded-lg text-sm
									{active ? 'bg-accent-soft text-accent-text font-medium' : 'text-fg hover:bg-control'}"
							>
								<item.icon size={16} aria-hidden="true" />
								{item.label}
							</a>
						{/each}
					</nav>

					<a
						href="/kit"
						aria-current={path.startsWith('/kit') ? 'page' : undefined}
						class="flex items-center gap-3 px-2 h-10 rounded-lg text-sm border-t border-line
							{path.startsWith('/kit') ? 'bg-accent-soft text-accent-text font-medium' : 'text-fg hover:bg-control'}"
					>
						<SlidersHorizontal size={16} aria-hidden="true" />
						My kit
					</a>

					<div class="flex items-center justify-between px-2 h-10 border-t border-line">
						<span class="text-sm text-fg-muted">Theme</span>
						<ThemeToggle />
					</div>

					<button
						type="button"
						onclick={handleLogout}
						class="flex items-center gap-3 px-2 h-10 rounded-lg text-sm text-fg hover:bg-control cursor-pointer border-t border-line"
					>
						<LogOut size={16} aria-hidden="true" />
						Sign out
					</button>
				</div>
			{/if}
		</div>
	</div>
</header>
