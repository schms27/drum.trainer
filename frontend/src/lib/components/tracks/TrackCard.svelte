<script lang="ts">
	import type { Track } from '$lib/types/track';
	import { deleteTrack } from '$lib/api/tracks';

	let {
		track,
		onDeleted
	}: {
		track: Track;
		onDeleted: (id: number) => void;
	} = $props();

	let deleting = $state(false);

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
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
</div>
