<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import { Minus, Plus, Redo2, Save, Trash, Undo2 } from '@lucide/svelte';
	import { replaceTrackNotes } from '$lib/api/tracks';
	import { NAMED_NOTES, noteLabel } from '$lib/drums';
	import {
		DIVISIONS,
		History,
		applyNoteActions,
		barsFor,
		moveNotes,
		newNote,
		repeatBars,
		sortNotes,
		stepMs as gridStep,
		withIds,
		type Division,
		type EditorNote,
		type NoteAction
	} from '$lib/editor/notes';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import VolumeSlider from '$lib/components/ui/VolumeSlider.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import BulkReplace from '$lib/components/editor/BulkReplace.svelte';
	import NoteGrid from '$lib/components/editor/NoteGrid.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	let track = $state(data.track);

	const barMs = $derived((60000 / track.bpm) * track.time_sig_numerator);
	const lastNoteMs = (list: { time_ms: number }[]) => list.reduce((m, n) => Math.max(m, n.time_ms), 0);
	const initialBars = (t: typeof track) => barsFor(Math.max(t.duration_ms, lastNoteMs(t.notes) + 1), t.bpm, t.time_sig_numerator);

	// ── Working copy and history ────────────────────────────────────────────
	// svelte-ignore state_referenced_locally
	let notes = $state<EditorNote[]>(withIds(track.notes));
	// svelte-ignore state_referenced_locally
	let bars = $state(initialBars(track));
	// svelte-ignore state_referenced_locally
	let savedKey = $state(keyOf(notes, bars));
	const history = new History();
	let historyVersion = $state(0); // bumps so canUndo/canRedo re-render
	let selected = $state<Set<number>>(new Set());
	let extraRows = $state<number[]>([]);

	function keyOf(list: EditorNote[], barCount: number) {
		return JSON.stringify([barCount, sortNotes(list).map((n) => [n.note_number, Math.round(n.time_ms * 1000), n.velocity])]);
	}
	const dirty = $derived(keyOf(notes, bars) !== savedKey);

	function change(next: EditorNote[]) {
		history.push(notes);
		historyVersion++;
		notes = next;
	}
	function undo() {
		const prev = history.undo(notes);
		if (prev) {
			notes = prev;
			historyVersion++;
			selected = new Set();
		}
	}
	function redo() {
		const next = history.redo(notes);
		if (next) {
			notes = next;
			historyVersion++;
			selected = new Set();
		}
	}
	const canUndo = $derived(historyVersion >= 0 && history.canUndo);
	const canRedo = $derived(historyVersion >= 0 && history.canRedo);

	// ── Grid ────────────────────────────────────────────────────────────────
	let division = $state<Division>('1/16');
	let pxPerBeat = $state(64);
	const step = $derived(gridStep(track.bpm, division));
	// Cymbals at the top, kick at the bottom: roughly descending note numbers.
	const rows = $derived([...new Set([...notes.map((n) => n.note_number), ...extraRows])].sort((a, b) => b - a));
	let addRowNote = $state('');

	function addRow() {
		const n = Number(addRowNote);
		if (Number.isInteger(n) && n >= 0 && n <= 127 && !rows.includes(n)) extraRows = [...extraRows, n];
		addRowNote = '';
	}

	function setBars(value: number) {
		const next = Math.max(1, Math.min(999, Math.round(value) || 1));
		const end = next * barMs;
		const beyond = notes.filter((n) => n.time_ms >= end);
		if (beyond.length && !confirm(`Shortening to ${next} bars removes ${beyond.length} hits after the end. Continue?`)) return;
		if (beyond.length) change(notes.filter((n) => n.time_ms < end));
		bars = next;
	}

	// ── Repeat bars ─────────────────────────────────────────────────────────
	let repeatFrom = $state(1);
	let repeatTo = $state(1);
	let repeatTimes = $state(1);
	function applyRepeat() {
		const from = Math.max(1, Math.round(repeatFrom));
		const to = Math.min(bars, Math.max(from, Math.round(repeatTo)));
		const times = Math.max(1, Math.min(64, Math.round(repeatTimes)));
		if (bars + (to - from + 1) * times > 999) {
			error = 'That would make the track longer than 999 bars.';
			return;
		}
		const res = repeatBars(notes, barMs, from, to, times);
		change(res.notes);
		bars += res.addedBars;
	}

	// ── Selection ───────────────────────────────────────────────────────────
	const selectedNotes = $derived(notes.filter((n) => selected.has(n.id)));
	let selectionVelocity = $state(100);
	$effect(() => {
		if (selectedNotes.length) selectionVelocity = selectedNotes[0].velocity;
	});

	function toggleSelect(id: number) {
		const next = new Set(selected);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		selected = next;
	}
	function setSelectionVelocity(level: number) {
		const v = Math.max(1, Math.round(level * 127));
		change(notes.map((n) => (selected.has(n.id) ? { ...n, velocity: v } : n)));
	}
	function deleteSelection() {
		change(notes.filter((n) => !selected.has(n.id)));
		selected = new Set();
	}

	function onKey(e: KeyboardEvent) {
		const el = e.target as HTMLElement;
		if (el.closest('input, select, textarea')) return;
		const mod = e.ctrlKey || e.metaKey;
		if (mod && e.key.toLowerCase() === 'z') {
			e.preventDefault();
			if (e.shiftKey) redo();
			else undo();
		} else if (mod && e.key.toLowerCase() === 'y') {
			e.preventDefault();
			redo();
		} else if ((e.key === 'Delete' || e.key === 'Backspace') && selected.size) {
			e.preventDefault();
			deleteSelection();
		} else if (e.key === 'Escape') {
			selected = new Set();
		}
	}

	// ── Save / discard ──────────────────────────────────────────────────────
	let saving = $state(false);
	let error = $state('');
	let savedMessage = $state(false);

	async function save() {
		saving = true;
		error = '';
		try {
			const lengthChanged = bars !== initialBars(track);
			track = await replaceTrackNotes(
				track.id,
				sortNotes(notes).map(({ note_number, time_ms, velocity }) => ({ note_number, time_ms, velocity })),
				lengthChanged ? bars * barMs : undefined
			);
			savedKey = keyOf(notes, bars);
			savedMessage = true;
			setTimeout(() => (savedMessage = false), 2500);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not save the notes';
		} finally {
			saving = false;
		}
	}

	function discard() {
		if (!confirm('Discard all changes since the last save?')) return;
		change(withIds(track.notes));
		bars = initialBars(track);
		savedKey = keyOf(notes, bars);
		selected = new Set();
		extraRows = [];
	}

	beforeNavigate((nav) => {
		if (dirty && !saving && !confirm('You have unsaved changes. Leave the editor anyway?')) nav.cancel();
	});

	const playHits = $derived(sortNotes(notes));
	const selectClass =
		'bg-control text-fg text-sm rounded-lg px-2 h-9 border border-line focus:outline-none focus:ring-2 focus:ring-accent';
</script>

<svelte:head><title>Edit {track.name} — Drum Trainer</title></svelte:head>
<svelte:window onkeydown={onKey} />

<PageHeader
	title="Edit notes"
	subtitle="{track.name} · {Math.round(track.bpm)} BPM · {notes.length} notes"
	back={{ href: `/tracks/${track.id}`, label: track.name }}
>
	{#snippet actions()}
		<Button variant="ghost" onclick={discard} disabled={!dirty || saving}>Discard changes</Button>
		<Button variant="primary" onclick={save} disabled={!dirty || saving}>
			<Save size={16} aria-hidden="true" />
			{saving ? 'Saving…' : savedMessage ? 'Saved' : 'Save'}
		</Button>
	{/snippet}
</PageHeader>

{#if error}<Notice variant="error">{error}</Notice>{/if}
{#if track.share_token}
	<Notice variant="warning">
		This track is shared. Saving changes what people see on the share link; copies people already added keep their notes.
	</Notice>
{/if}
{#if (track.stats?.sessions ?? 0) > 0}
	<p class="text-sm text-fg-muted">
		Your earlier takes were scored against the old notes. Progress marks where the notes changed.
	</p>
{/if}

<Card title="Simplify the kit" description="Move every hit of a drum to another one, or remove it — for example to play the track with just kick, snare and one cymbal.">
	<BulkReplace {notes} onApply={(actions: Map<number, NoteAction>) => change(applyNoteActions(notes, actions))} />
</Card>

<Card title="Notes" description="Click an empty spot to add a hit, click a hit to remove it. Shift-click to select hits, drag to move them.">
	<div class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
			<div class="flex items-center gap-1">
				<Button size="sm" variant="ghost" onclick={undo} disabled={!canUndo} aria-label="Undo" title="Undo (Ctrl+Z)">
					<Undo2 size={16} aria-hidden="true" />
				</Button>
				<Button size="sm" variant="ghost" onclick={redo} disabled={!canRedo} aria-label="Redo" title="Redo (Ctrl+Y)">
					<Redo2 size={16} aria-hidden="true" />
				</Button>
			</div>
			<label class="flex items-center gap-2 text-sm text-fg-muted">
				Snap
				<select class={selectClass} bind:value={division}>
					{#each DIVISIONS as d (d.value)}<option value={d.value}>{d.label}</option>{/each}
				</select>
			</label>
			<div class="flex items-center gap-1 text-sm text-fg-muted">
				<span>Zoom</span>
				<Button size="sm" variant="ghost" onclick={() => (pxPerBeat = Math.max(24, pxPerBeat / 1.25))} aria-label="Zoom out">
					<Minus size={14} aria-hidden="true" />
				</Button>
				<Button size="sm" variant="ghost" onclick={() => (pxPerBeat = Math.min(240, pxPerBeat * 1.25))} aria-label="Zoom in">
					<Plus size={14} aria-hidden="true" />
				</Button>
			</div>
			<label class="flex items-center gap-2 text-sm text-fg-muted">
				Bars
				<input
					type="number"
					min="1"
					max="999"
					value={bars}
					onchange={(e) => setBars(Number(e.currentTarget.value))}
					class="w-16 h-9 bg-control border border-line rounded-lg px-2 text-sm text-fg tabular-nums focus:outline-none focus:ring-2 focus:ring-accent"
				/>
			</label>
			<div class="flex items-center gap-2">
				<label class="sr-only" for="add-row">Add a row for a drum</label>
				<select id="add-row" class={selectClass} bind:value={addRowNote} onchange={addRow}>
					<option value="">Add a drum…</option>
					{#each NAMED_NOTES.filter((n) => !rows.includes(n)) as n (n)}
						<option value={String(n)}>{noteLabel(n)} #{n}</option>
					{/each}
				</select>
			</div>
			<div class="sm:ml-auto"><PlayButton hits={playHits} label="Play" /></div>
		</div>

		{#if rows.length === 0}
			<p class="text-sm text-fg-subtle">No notes yet. Add a drum to start.</p>
		{:else}
			<NoteGrid
				{notes}
				{rows}
				bpm={track.bpm}
				beatsPerBar={track.time_sig_numerator}
				{bars}
				stepMs={step}
				{pxPerBeat}
				{selected}
				onAdd={(n, t) => change([...notes, newNote(n, t)])}
				onRemove={(id) => {
					change(notes.filter((n) => n.id !== id));
					if (selected.has(id)) toggleSelect(id);
				}}
				onMove={(ids, dMs, dRows) => change(moveNotes(notes, ids, dMs, dRows, rows))}
				onToggleSelect={toggleSelect}
			/>
		{/if}

		<div class="flex flex-wrap items-center gap-2 text-sm text-fg-muted">
			<span>Repeat bars</span>
			<label class="sr-only" for="repeat-from">First bar to repeat</label>
			<input id="repeat-from" type="number" min="1" max={bars} bind:value={repeatFrom} class="w-16 h-8 bg-control border border-line rounded-lg px-2 text-fg tabular-nums" />
			<span>to</span>
			<label class="sr-only" for="repeat-to">Last bar to repeat</label>
			<input id="repeat-to" type="number" min="1" max={bars} bind:value={repeatTo} class="w-16 h-8 bg-control border border-line rounded-lg px-2 text-fg tabular-nums" />
			<label class="sr-only" for="repeat-times">How many times</label>
			<input id="repeat-times" type="number" min="1" max="64" bind:value={repeatTimes} class="w-14 h-8 bg-control border border-line rounded-lg px-2 text-fg tabular-nums" />
			<span>more {repeatTimes === 1 ? 'time' : 'times'}</span>
			<Button size="sm" variant="ghost" onclick={applyRepeat}>Repeat</Button>
		</div>

		{#if selectedNotes.length}
			<div class="flex flex-wrap items-center gap-3 bg-control/50 rounded-lg px-3 py-2" aria-live="polite">
				<span class="text-sm text-fg">{selectedNotes.length} selected</span>
				<span class="text-sm text-fg-muted">Velocity</span>
				<div class="w-48">
					<VolumeSlider
						id="selection-velocity"
						label="Velocity of the selected hits"
						value={selectionVelocity / 127}
						oninput={setSelectionVelocity}
					/>
				</div>
				<Button size="sm" variant="ghost" onclick={deleteSelection}><Trash size={14} aria-hidden="true" /> Delete</Button>
				<Button size="sm" variant="ghost" onclick={() => (selected = new Set())}>Clear selection</Button>
			</div>
		{/if}
	</div>
</Card>

{#if !dirty && savedMessage}
	<p class="text-sm text-fg-muted">
		Saved. <button type="button" class="text-accent-text hover:underline cursor-pointer" onclick={() => goto(`/session/${track.id}`)}>Practise it now</button>
	</p>
{/if}
