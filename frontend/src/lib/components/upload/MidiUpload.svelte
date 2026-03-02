<script lang="ts">
	import { uploadTrack } from '$lib/api/tracks';
	import type { TrackWithNotes } from '$lib/types/track';

	let {
		onUploaded,
		channel = 10
	}: {
		onUploaded: (track: TrackWithNotes) => void;
		channel?: number;
	} = $props();

	let dragging = $state(false);
	let uploading = $state(false);
	let progress = $state(0);
	let error = $state('');

	async function handleFile(file: File) {
		if (!file.name.toLowerCase().endsWith('.mid') && !file.name.toLowerCase().endsWith('.midi')) {
			error = 'Only .mid / .midi files are supported.';
			return;
		}
		error = '';
		uploading = true;
		progress = 0;
		try {
			const track = await uploadTrack(file, channel, (pct) => (progress = pct));
			onUploaded(track);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const file = e.dataTransfer?.files[0];
		if (file) handleFile(file);
	}

	function onFileInput(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (file) handleFile(file);
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="border-2 border-dashed rounded-xl p-10 text-center transition-colors
	       {dragging ? 'border-indigo-400 bg-indigo-950/20' : 'border-zinc-600 hover:border-zinc-400'}"
	ondragover={(e) => { e.preventDefault(); dragging = true; }}
	ondragleave={() => (dragging = false)}
	ondrop={onDrop}
	role="region"
	aria-label="MIDI file drop zone"
>
	{#if uploading}
		<p class="text-zinc-300 mb-3">Uploading… {progress}%</p>
		<div class="w-full bg-zinc-700 rounded-full h-2">
			<div class="bg-indigo-500 h-2 rounded-full transition-all" style="width:{progress}%"></div>
		</div>
	{:else}
		<p class="text-zinc-300 mb-1 text-lg">Drop a MIDI file here</p>
		<p class="text-zinc-500 text-sm mb-4">or click to browse</p>
		<label
			class="cursor-pointer inline-block px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm transition"
		>
			Choose file
			<input type="file" accept=".mid,.midi" class="hidden" onchange={onFileInput} />
		</label>
	{/if}

	{#if error}
		<p class="mt-3 text-red-400 text-sm">{error}</p>
	{/if}
</div>
