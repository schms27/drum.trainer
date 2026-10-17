<script lang="ts">
	import { Upload } from '@lucide/svelte';
	import { uploadTrack } from '$lib/api/tracks';
	import type { TrackWithNotes } from '$lib/types/track';
	import Notice from '$lib/components/ui/Notice.svelte';

	let { onUploaded }: { onUploaded: (track: TrackWithNotes) => void } = $props();

	let dragging = $state(false);
	let uploading = $state(false);
	let progress = $state(0);
	let error = $state('');
	// MIDI channel 10 is where General MIDI puts drums; most files use it.
	let channel = $state(10);

	async function handleFile(file: File) {
		if (!/\.midi?$/i.test(file.name)) {
			error = 'That is not a MIDI file. Choose a file ending in .mid or .midi.';
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
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) handleFile(file);
		input.value = '';
	}
</script>

<div class="flex flex-col gap-4">
	<label
		class="flex flex-col items-center justify-center gap-2 text-center border-2 border-dashed rounded-xl px-6 py-10 cursor-pointer transition-colors
			{dragging ? 'border-accent bg-accent-soft' : 'border-line-strong hover:border-accent'}"
		ondragover={(e) => {
			e.preventDefault();
			dragging = true;
		}}
		ondragleave={() => (dragging = false)}
		ondrop={onDrop}
	>
		{#if uploading}
			<p class="text-fg">Uploading… {progress}%</p>
			<div class="w-full max-w-xs bg-control rounded-full h-2" role="progressbar" aria-valuenow={progress} aria-valuemin={0} aria-valuemax={100}>
				<div class="bg-accent h-2 rounded-full transition-all" style="width:{progress}%"></div>
			</div>
		{:else}
			<Upload size={28} class="text-fg-subtle" aria-hidden="true" />
			<p class="text-fg font-medium">Drop a MIDI file here, or click to choose one</p>
			<p class="text-fg-subtle text-sm">.mid or .midi</p>
			<input type="file" accept=".mid,.midi" class="sr-only" onchange={onFileInput} disabled={uploading} />
		{/if}
	</label>

	{#if error}
		<Notice variant="error">{error}</Notice>
	{/if}

	<details class="text-sm text-fg-muted">
		<summary class="cursor-pointer w-fit hover:text-fg">Advanced: drum channel</summary>
		<div class="mt-3 flex flex-wrap items-center gap-3">
			<label for="drum-channel" class="text-fg">Read drum notes from MIDI channel</label>
			<select
				id="drum-channel"
				bind:value={channel}
				class="bg-control text-fg text-sm rounded-lg px-2 h-9 border border-line focus:outline-none focus:ring-2 focus:ring-accent"
			>
				{#each Array.from({ length: 16 }, (_, i) => i + 1) as ch}
					<option value={ch}>{ch}{ch === 10 ? ' (standard for drums)' : ''}</option>
				{/each}
			</select>
			<p class="w-full text-fg-subtle">
				Only change this if the upload says no notes were found — some files put drums on another channel.
			</p>
		</div>
	</details>
</div>
