<script lang="ts">
	import { Trash, Upload } from '@lucide/svelte';
	import { uploadPilot, deletePilot, type PilotTrack } from '$lib/api/pilots';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	// Backing tracks ("pilot tracks" in the API): audio that plays along while recording.
	let { trackId, items = $bindable([]) }: { trackId: number; items?: PilotTrack[] } = $props();

	let uploadName = $state('');
	let uploadFile = $state<File | null>(null);
	let uploading = $state(false);
	let uploadProgress = $state(0);
	let error = $state('');
	let fileInput = $state<HTMLInputElement>();

	function formatSize(bytes: number): string {
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	function handleFileChange(e: Event) {
		uploadFile = (e.target as HTMLInputElement).files?.[0] ?? null;
		if (uploadFile && !uploadName) uploadName = uploadFile.name.replace(/\.[^.]+$/, '');
	}

	async function handleUpload() {
		if (!uploadFile || !uploadName.trim()) return;
		uploading = true;
		uploadProgress = 0;
		error = '';
		try {
			const added = await uploadPilot(trackId, uploadName.trim(), uploadFile, (pct) => (uploadProgress = pct));
			items = [...items, added];
			uploadName = '';
			uploadFile = null;
			if (fileInput) fileInput.value = '';
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	async function handleDelete(id: number) {
		error = '';
		try {
			await deletePilot(trackId, id);
			items = items.filter((p) => p.id !== id);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not remove the backing track';
		}
	}
</script>

<div class="flex flex-col gap-4">
	{#if items.length > 0}
		<ul class="flex flex-col divide-y divide-line-subtle">
			{#each items as item (item.id)}
				<li class="flex items-center gap-3 py-2 text-sm">
					<span class="flex-1 text-fg truncate" title={item.name}>{item.name}</span>
					<span class="text-fg-subtle text-xs shrink-0">{formatSize(item.file_size)}</span>
					<button
						type="button"
						onclick={() => handleDelete(item.id)}
						class="w-8 h-8 flex items-center justify-center rounded-lg text-fg-subtle hover:text-bad hover:bg-control cursor-pointer"
						aria-label="Remove backing track {item.name}"
						title="Remove"
					>
						<Trash size={14} aria-hidden="true" />
					</button>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="text-sm text-fg-muted">No backing tracks yet.</p>
	{/if}

	<div class="flex flex-col gap-2 border-t border-line pt-4">
		<label class="text-sm font-medium text-fg" for="backing-file-{trackId}">Add audio (MP3, WAV, OGG, FLAC, AAC, M4A)</label>
		<input
			bind:this={fileInput}
			id="backing-file-{trackId}"
			type="file"
			accept="audio/*"
			onchange={handleFileChange}
			class="text-sm text-fg-muted file:mr-3 file:h-9 file:px-3 file:rounded-lg file:border-0 file:text-sm file:bg-control file:text-fg hover:file:bg-control-hover file:cursor-pointer"
		/>
		{#if uploadFile}
			<div class="flex flex-wrap gap-2">
				<input
					type="text"
					placeholder="Name, e.g. Full mix"
					bind:value={uploadName}
					aria-label="Backing track name"
					class="flex-1 min-w-[12rem] bg-control text-fg text-sm rounded-lg px-3 h-10 border border-line placeholder:text-fg-subtle focus:outline-none focus:ring-2 focus:ring-accent"
				/>
				<Button variant="primary" onclick={handleUpload} disabled={uploading || !uploadName.trim()}>
					<Upload size={16} aria-hidden="true" />
					{uploading ? `Uploading… ${uploadProgress}%` : 'Upload'}
				</Button>
			</div>
		{/if}
		{#if error}<Notice variant="error">{error}</Notice>{/if}
	</div>
</div>
