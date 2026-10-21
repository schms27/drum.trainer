<script lang="ts">
	import { Timer, Volume2, VolumeX } from '@lucide/svelte';
	import type { PilotPlayer } from '$lib/audio/PilotPlayer';
	import type { PilotTrack } from '$lib/api/pilots';
	import { levelToGain } from '$lib/audio/mix';
	import { sessionMix } from '$lib/audio/session-mix.svelte';
	import VolumeSlider from '$lib/components/ui/VolumeSlider.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	// Levels for the metronome and the backing tracks ("pilot tracks" in the
	// API). They work during a take too: the metronome reads sessionMix for each
	// click, and backing tracks are ramped to the new gain right away.
	let {
		pilots,
		player,
		ready = true,
		kitName = null,
		onChange,
		onSaveAsKitDefault,
		onResetToKitDefault
	}: {
		pilots: PilotTrack[];
		player: PilotPlayer | null;
		ready?: boolean;
		/** Name of the active kit profile, which holds the default levels. */
		kitName?: string | null;
		/** Called after any level changed (the page saves it for the track). */
		onChange: () => void;
		onSaveAsKitDefault?: () => Promise<void>;
		onResetToKitDefault?: () => void;
	} = $props();

	let mutedIds = $state<Set<number>>(new Set());
	let savingDefault = $state(false);
	let savedDefault = $state(false);

	function toggleMute(id: number) {
		const nowMuted = !mutedIds.has(id);
		player?.setMuted(id, nowMuted);
		const next = new Set(mutedIds);
		if (nowMuted) next.add(id);
		else next.delete(id);
		mutedIds = next;
	}

	function setPilotLevel(id: number, level: number) {
		sessionMix.pilots[id] = level;
		player?.setVolume(id, levelToGain(level));
		onChange();
	}

	async function saveDefault() {
		if (!onSaveAsKitDefault) return;
		savingDefault = true;
		try {
			await onSaveAsKitDefault();
			savedDefault = true;
			setTimeout(() => (savedDefault = false), 2000);
		} finally {
			savingDefault = false;
		}
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-3">
		<span class="w-9 h-9 rounded-lg flex items-center justify-center bg-control text-fg-muted shrink-0" aria-hidden="true">
			<Timer size={16} />
		</span>
		<span class="w-28 sm:w-40 shrink-0 text-sm text-fg truncate">Metronome</span>
		<VolumeSlider
			id="mix-metronome"
			label="Metronome volume"
			bind:value={sessionMix.metronome}
			oninput={() => onChange()}
		/>
	</div>

	{#each pilots as pilot (pilot.id)}
		{@const muted = mutedIds.has(pilot.id)}
		<div class="flex items-center gap-3">
			<button
				type="button"
				onclick={() => toggleMute(pilot.id)}
				disabled={!player}
				aria-pressed={muted}
				aria-label={muted ? `Unmute ${pilot.name}` : `Mute ${pilot.name}`}
				class="w-9 h-9 shrink-0 rounded-lg flex items-center justify-center transition-colors cursor-pointer
					disabled:opacity-40 disabled:cursor-not-allowed
					{muted ? 'bg-control text-fg-subtle hover:bg-control-hover' : 'bg-accent-soft text-accent-text'}"
			>
				{#if muted}<VolumeX size={16} aria-hidden="true" />{:else}<Volume2 size={16} aria-hidden="true" />{/if}
			</button>
			<span class="w-28 sm:w-40 shrink-0 text-sm text-fg truncate" title={pilot.name}>{pilot.name}</span>
			<VolumeSlider
				id="mix-pilot-{pilot.id}"
				label="Volume of {pilot.name}"
				value={sessionMix.pilots[pilot.id] ?? 1}
				oninput={(v) => setPilotLevel(pilot.id, v)}
			/>
		</div>
	{/each}

	<div class="flex flex-wrap items-center gap-x-3 gap-y-2 mt-1">
		<p class="text-xs text-fg-subtle flex-1 min-w-[12rem]">
			{#if !ready}
				Loading audio…
			{:else if pilots.length > 0 && !player}
				Backing tracks start together with the track. Levels can be changed while you record and are saved for this track.
			{:else if pilots.length > 0}
				Playing along with your recording.
			{:else}
				Levels can be changed while you record and are saved for this track.
			{/if}
		</p>
		{#if kitName && onSaveAsKitDefault}
			<Button size="sm" variant="ghost" onclick={saveDefault} disabled={savingDefault} title="Use these levels for tracks without their own mix">
				{savedDefault ? 'Saved' : savingDefault ? 'Saving…' : `Save as ${kitName} default`}
			</Button>
		{/if}
		{#if onResetToKitDefault}
			<Button size="sm" variant="ghost" onclick={onResetToKitDefault}>
				Reset to {kitName ? 'kit' : 'full'} levels
			</Button>
		{/if}
	</div>
</div>
