<script lang="ts">
	import { Clock, Gauge, Music, Trash, SlidersHorizontal, AudioLines, ChevronDown, Play } from '@lucide/svelte';
	import type { Track } from '$lib/types/track';
	import { deleteTrack } from '$lib/api/tracks';
	import { listPilots, uploadPilot, deletePilot, type PilotTrack } from '$lib/api/pilots';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';

	let { track, onDeleted }: { track: Track; onDeleted: (id: number) => void } = $props();

	let confirmDelete = $state(false);

	// Backing tracks ("pilot tracks" in the API): audio that plays along while recording.
	let showBacking = $state(false);
	let backing = $state<PilotTrack[]>([]);
	let backingLoaded = $state(false);
	let backingLoading = $state(false);

	let uploadName = $state('');
	let uploadFile = $state<File | null>(null);
	let uploading = $state(false);
	let uploadProgress = $state(0);
	let uploadError = $state('');
	let fileInput = $state<HTMLInputElement>();

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	async function handleDelete() {
		await deleteTrack(track.id);
		onDeleted(track.id);
	}

	async function toggleBacking() {
		showBacking = !showBacking;
		if (showBacking && !backingLoaded) {
			backingLoading = true;
			try {
				backing = await listPilots(track.id);
				backingLoaded = true;
			} finally {
				backingLoading = false;
			}
		}
	}

	function handleFileChange(e: Event) {
		uploadFile = (e.target as HTMLInputElement).files?.[0] ?? null;
		if (uploadFile && !uploadName) uploadName = uploadFile.name.replace(/\.[^.]+$/, '');
	}

	async function handleUpload() {
		if (!uploadFile || !uploadName.trim()) return;
		uploading = true;
		uploadProgress = 0;
		uploadError = '';
		try {
			const added = await uploadPilot(track.id, uploadName.trim(), uploadFile, (pct) => (uploadProgress = pct));
			backing = [...backing, added];
			uploadName = '';
			uploadFile = null;
			if (fileInput) fileInput.value = '';
		} catch (e: unknown) {
			uploadError = e instanceof Error ? e.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	async function handleDeleteBacking(id: number) {
		await deletePilot(track.id, id);
		backing = backing.filter((p) => p.id !== id);
	}
</script>

<article class="bg-card border border-line rounded-xl p-5 flex flex-col gap-4">
	<div class="flex items-start justify-between gap-2">
		<div class="min-w-0">
			<h3 class="font-semibold text-fg leading-tight break-words">{track.name}</h3>
			<p class="text-fg-subtle text-xs mt-1 truncate" title={track.filename}>{track.filename}</p>
		</div>
		<button
			type="button"
			onclick={() => (confirmDelete = true)}
			class="shrink-0 w-8 h-8 -mr-2 -mt-1 flex items-center justify-center rounded-lg text-fg-subtle hover:text-bad hover:bg-control transition-colors cursor-pointer"
			aria-label="Delete {track.name}"
			title="Delete track"
		>
			<Trash size={16} aria-hidden="true" />
		</button>
	</div>

	<dl class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-fg-muted">
		<div class="flex items-center gap-1" title="Length">
			<Clock size={14} aria-hidden="true" /><dt class="sr-only">Length</dt><dd>{formatDuration(track.duration_ms)}</dd>
		</div>
		<div class="flex items-center gap-1" title="Tempo">
			<Gauge size={14} aria-hidden="true" /><dt class="sr-only">Tempo</dt><dd>{Math.round(track.bpm)} BPM</dd>
		</div>
		<div class="flex items-center gap-1" title="Notes">
			<Music size={14} aria-hidden="true" /><dt class="sr-only">Notes</dt><dd>{track.note_count} notes</dd>
		</div>
	</dl>

	<Button href="/session/{track.id}" variant="primary" block>
		<Play size={16} aria-hidden="true" /> Practice
	</Button>

	<div class="flex flex-wrap gap-2 -mt-1">
		<Button href="/tracks/{track.id}/mapping" variant="ghost" size="sm">
			<SlidersHorizontal size={14} aria-hidden="true" /> Pad mapping
		</Button>
		<Button variant="ghost" size="sm" onclick={toggleBacking} aria-expanded={showBacking}>
			<AudioLines size={14} aria-hidden="true" />
			Backing tracks{backingLoaded ? ` (${backing.length})` : ''}
			<ChevronDown size={14} class="transition-transform {showBacking ? 'rotate-180' : ''}" aria-hidden="true" />
		</Button>
	</div>

	{#if showBacking}
		<div class="flex flex-col gap-3 border-t border-line pt-4">
			<p class="text-xs text-fg-muted">
				Audio (MP3, WAV, …) that plays along while you record — for example the song without drums.
			</p>
			{#if backingLoading}
				<p class="text-fg-subtle text-xs">Loading…</p>
			{:else}
				{#each backing as item (item.id)}
					<div class="flex items-center gap-2 text-sm">
						<span class="flex-1 text-fg truncate" title={item.name}>{item.name}</span>
						<span class="text-fg-subtle text-xs shrink-0">{formatSize(item.file_size)}</span>
						<button
							type="button"
							onclick={() => handleDeleteBacking(item.id)}
							class="w-7 h-7 flex items-center justify-center rounded-md text-fg-subtle hover:text-bad hover:bg-control cursor-pointer"
							aria-label="Remove backing track {item.name}"
						>
							<Trash size={14} aria-hidden="true" />
						</button>
					</div>
				{/each}

				<div class="flex flex-col gap-2">
					<input
						bind:this={fileInput}
						type="file"
						accept="audio/*"
						onchange={handleFileChange}
						aria-label="Backing track audio file"
						class="text-xs text-fg-muted file:mr-2 file:h-8 file:px-3 file:rounded-lg file:border-0 file:text-xs file:bg-control file:text-fg hover:file:bg-control-hover file:cursor-pointer"
					/>
					{#if uploadFile}
						<input
							type="text"
							placeholder="Name, e.g. Full mix"
							bind:value={uploadName}
							aria-label="Backing track name"
							class="bg-control text-fg text-sm rounded-lg px-3 h-9 border border-line placeholder:text-fg-subtle focus:outline-none focus:ring-2 focus:ring-accent"
						/>
						{#if uploading}
							<div class="h-1.5 bg-control rounded-full overflow-hidden" role="progressbar" aria-valuenow={uploadProgress} aria-valuemin={0} aria-valuemax={100}>
								<div class="h-full bg-accent rounded-full transition-[width]" style="width: {uploadProgress}%"></div>
							</div>
						{:else}
							<Button size="sm" onclick={handleUpload} disabled={!uploadName.trim()}>Add backing track</Button>
						{/if}
					{/if}
					{#if uploadError}<p class="text-bad text-xs">{uploadError}</p>{/if}
				</div>
			{/if}
		</div>
	{/if}
</article>

<ConfirmDialog
	bind:open={confirmDelete}
	title="Delete “{track.name}”?"
	message="This removes the track, its backing tracks and all practice sessions recorded for it. This cannot be undone."
	confirmLabel="Delete track"
	danger
	onConfirm={handleDelete}
/>
