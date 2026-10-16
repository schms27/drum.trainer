<script lang="ts">
	import Button from './Button.svelte';

	let {
		open = $bindable(false),
		title,
		message,
		confirmLabel = 'Confirm',
		danger = false,
		onConfirm
	}: {
		open?: boolean;
		title: string;
		message: string;
		confirmLabel?: string;
		danger?: boolean;
		onConfirm: () => void | Promise<void>;
	} = $props();

	let dialog: HTMLDialogElement;
	let busy = $state(false);
	let error = $state('');

	$effect(() => {
		if (open && !dialog.open) {
			error = '';
			dialog.showModal();
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});

	async function confirm() {
		busy = true;
		error = '';
		try {
			await onConfirm();
			open = false;
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Something went wrong.';
		} finally {
			busy = false;
		}
	}
</script>

<dialog
	bind:this={dialog}
	onclose={() => (open = false)}
	class="m-auto w-[calc(100%-2rem)] max-w-md rounded-xl border border-line bg-card text-fg p-0 shadow-xl backdrop:bg-black/50"
>
	<div class="p-6 flex flex-col gap-4">
		<h2 class="text-lg font-semibold">{title}</h2>
		<p class="text-sm text-fg-muted">{message}</p>
		{#if error}<p class="text-sm text-bad">{error}</p>{/if}
		<div class="flex justify-end gap-2 mt-2">
			<Button variant="ghost" onclick={() => (open = false)} disabled={busy}>Cancel</Button>
			<Button variant={danger ? 'danger' : 'primary'} onclick={confirm} disabled={busy}>
				{confirmLabel}
			</Button>
		</div>
	</div>
</dialog>
