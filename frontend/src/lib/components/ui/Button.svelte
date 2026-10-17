<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAnchorAttributes, HTMLButtonAttributes } from 'svelte/elements';

	type Variant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'success';
	type Size = 'sm' | 'md' | 'lg';

	type Props = {
		variant?: Variant;
		size?: Size;
		/** Renders a link styled as a button. */
		href?: string;
		/** Stretch to the container width. */
		block?: boolean;
		class?: string;
		children: Snippet;
	} & Omit<HTMLButtonAttributes, 'class' | 'children'> &
		Omit<HTMLAnchorAttributes, 'class' | 'children'>;

	let {
		variant = 'secondary',
		size = 'md',
		href,
		block = false,
		class: extra = '',
		children,
		...rest
	}: Props = $props();

	const variants: Record<Variant, string> = {
		primary: 'bg-accent text-accent-fg hover:bg-accent-hover',
		secondary: 'bg-control text-fg hover:bg-control-hover',
		ghost: 'text-fg-muted hover:text-fg hover:bg-control',
		danger: 'bg-bad-solid text-white hover:opacity-90',
		success: 'bg-good-solid text-white hover:opacity-90'
	};
	const sizes: Record<Size, string> = {
		sm: 'h-8 px-3 text-sm gap-1.5',
		md: 'h-10 px-4 text-sm gap-2',
		lg: 'h-12 px-6 text-base gap-2'
	};

	const classes = $derived(
		[
			'inline-flex items-center justify-center rounded-lg font-medium transition-colors cursor-pointer',
			'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent',
			'disabled:opacity-40 disabled:cursor-not-allowed',
			variants[variant],
			sizes[size],
			block ? 'w-full' : '',
			extra
		].join(' ')
	);
</script>

{#if href}
	<a {href} class={classes} {...rest as HTMLAnchorAttributes}>{@render children()}</a>
{:else}
	<button type="button" class={classes} {...rest as HTMLButtonAttributes}>{@render children()}</button>
{/if}
