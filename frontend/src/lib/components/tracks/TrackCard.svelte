<script lang="ts">
	import type { Track } from '$lib/types/track';
	import { deleteTrack } from '$lib/api/tracks';
	import { listPilots, uploadPilot, deletePilot, type PilotTrack } from '$lib/api/pilots';

	let {
		track,
		onDeleted
	}: {
		track: Track;
		onDeleted: (id: number) => void;
	} = $props();

	let deleting = $state(false);
	let showPilots = $state(false);
	let pilots = $state<PilotTrack[]>([]);
	let pilotsLoaded = $state(false);
	let pilotsLoading = $state(false);

	// Upload state
	let uploadName = $state('');
	let uploadFile = $state<File | null>(null);
	let uploading = $state(false);
	let uploadProgress = $state(0);
	let uploadError = $state('');

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	async function handleDelete() {
		if (!confirm(`Delete "${track.name}"?`)) return;
		deleting = true;
		try {
			await deleteTrack(track.id);
			onDeleted(track.id);
		} finally {
			deleting = false;
		}
	}

	async function togglePilots() {
		showPilots = !showPilots;
		if (showPilots && !pilotsLoaded) {
			pilotsLoading = true;
			try {
				pilots = await listPilots(track.id);
				pilotsLoaded = true;
			} finally {
				pilotsLoading = false;
			}
		}
	}

	function handleFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		uploadFile = input.files?.[0] ?? null;
		if (uploadFile && !uploadName) {
			uploadName = uploadFile.name.replace(/\.[^.]+$/, '');
		}
	}

	async function handleUpload() {
		if (!uploadFile || !uploadName.trim()) return;
		uploading = true;
		uploadProgress = 0;
		uploadError = '';
		try {
			const pilot = await uploadPilot(track.id, uploadName.trim(), uploadFile, (pct) => {
				uploadProgress = pct;
			});
			pilots = [...pilots, pilot];
			uploadName = '';
			uploadFile = null;
			uploadProgress = 0;
			// Reset file input
			const input = document.getElementById(`pilot-file-${track.id}`) as HTMLInputElement | null;
			if (input) input.value = '';
		} catch (e: unknown) {
			uploadError = e instanceof Error ? e.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	async function handleDeletePilot(pilotId: number) {
		await deletePilot(track.id, pilotId);
		pilots = pilots.filter((p) => p.id !== pilotId);
	}
</script>

<div class="bg-zinc-800 rounded-xl p-5 flex flex-col gap-3 border border-zinc-700">
	<div class="flex items-start justify-between gap-2">
		<div>
			<h3 class="text-white font-semibold text-base leading-tight">{track.name}</h3>
			<p class="text-zinc-500 text-xs mt-0.5 truncate max-w-[200px]" title={track.filename}>
				{track.filename}
			</p>
		</div>
		<button
			onclick={handleDelete}
			disabled={deleting}
			class="text-zinc-500 hover:text-red-400 transition text-sm"
			aria-label="Delete track"
		>✕</button>
	</div>

	<div class="flex gap-4 text-xs text-zinc-400">
		<span>⏱ {formatDuration(track.duration_ms)}</span>
		<span>🎵 {track.note_count} notes</span>
	</div>

	<a
		href="/session/{track.id}"
		class="mt-1 block text-center bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg py-2 text-sm font-medium transition"
	>
		Practice
	</a>
	<a
		href="/tracks/{track.id}/mapping"
		class="block text-center text-zinc-400 hover:text-white text-xs transition"
	>
		Edit Mapping
	</a>

	<!-- Pilot tracks section -->
	<button
		onclick={togglePilots}
		class="flex items-center gap-1 text-xs text-zinc-500 hover:text-zinc-300 transition text-left"
	>
		<span>{showPilots ? '▾' : '▸'}</span>
		<span>Pilot Tracks {pilotsLoaded ? `(${pilots.length})` : ''}</span>
	</button>

	{#if showPilots}
		<div class="flex flex-col gap-2 border-t border-zinc-700 pt-3">
			{#if pilotsLoading}
				<p class="text-zinc-500 text-xs">Loading…</p>
			{:else}
				{#each pilots as pilot (pilot.id)}
					<div class="flex items-center gap-2 text-xs">
						<span class="flex-1 text-zinc-300 truncate" title={pilot.name}>{pilot.name}</span>
						<span class="text-zinc-500 shrink-0">{formatSize(pilot.file_size)}</span>
						<button
							onclick={() => handleDeletePilot(pilot.id)}
							class="text-zinc-600 hover:text-red-400 transition shrink-0"
							aria-label="Delete pilot"
						>✕</button>
					</div>
				{/each}

				<!-- Upload form -->
				<div class="flex flex-col gap-2 mt-1">
					<input
						type="text"
						placeholder="Name"
						bind:value={uploadName}
						class="bg-zinc-700 border border-zinc-600 text-white text-xs rounded px-2 py-1 placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
					/>
					<input
						id="pilot-file-{track.id}"
						type="file"
						accept="audio/*"
						onchange={handleFileChange}
						class="text-xs text-zinc-400 file:mr-2 file:py-1 file:px-2 file:rounded file:border-0 file:text-xs file:bg-zinc-700 file:text-zinc-300 hover:file:bg-zinc-600 file:cursor-pointer"
					/>
					{#if uploading}
						<div class="h-1.5 bg-zinc-700 rounded-full overflow-hidden">
							<div
								class="h-full bg-indigo-500 rounded-full transition-[width] duration-100"
								style="width: {uploadProgress}%"
							></div>
						</div>
					{:else}
						<button
							onclick={handleUpload}
							disabled={!uploadFile || !uploadName.trim()}
							class="text-xs bg-zinc-700 hover:bg-zinc-600 disabled:opacity-40 disabled:cursor-not-allowed text-zinc-200 rounded px-3 py-1 transition"
						>
							Upload
						</button>
					{/if}
					{#if uploadError}
						<p class="text-red-400 text-xs">{uploadError}</p>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>
