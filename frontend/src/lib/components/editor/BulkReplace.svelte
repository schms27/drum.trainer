<script lang="ts">
	import { ArrowRight } from '@lucide/svelte';
	import { NAMED_NOTES, drumGroup, noteLabel } from '$lib/drums';
	import { applyNoteActions, countByNote, type EditorNote, type NoteAction } from '$lib/editor/notes';
	import Button from '$lib/components/ui/Button.svelte';

	// Replace or delete every hit of a note number across the whole track,
	// e.g. to play a song on a smaller kit. Shows the result before applying.
	let { notes, onApply }: { notes: EditorNote[]; onApply: (actions: Map<number, NoteAction>) => void } = $props();

	const used = $derived([...countByNote(notes).entries()].sort((a, b) => a[0] - b[0]));
	// Per source note: '' = keep, 'delete', a note number, or 'custom' (typed in).
	let choice = $state<Record<number, string>>({});
	let custom = $state<Record<number, number>>({});

	const actions = $derived.by(() => {
		const m = new Map<number, NoteAction>();
		for (const [note] of used) {
			const c = choice[note] ?? '';
			if (c === 'delete') m.set(note, 'delete');
			else if (c === 'custom') {
				const v = custom[note];
				if (Number.isInteger(v) && v >= 0 && v <= 127 && v !== note) m.set(note, v);
			} else if (c !== '' && Number(c) !== note) m.set(note, Number(c));
		}
		return m;
	});
	const after = $derived(actions.size ? countByNote(applyNoteActions(notes, actions)) : null);
	// Rows: every note before or after the change.
	const previewRows = $derived(
		after ? [...new Set([...used.map(([n]) => n), ...after.keys()])].sort((a, b) => a - b) : []
	);
	const targets = $derived([...new Set([...NAMED_NOTES, ...used.map(([n]) => n)])].sort((a, b) => a - b));

	function apply() {
		onApply(actions);
		choice = {};
		custom = {};
	}

	const selectClass =
		'bg-control text-fg text-sm rounded-lg px-2 h-9 border border-line max-w-full focus:outline-none focus:ring-2 focus:ring-accent';
</script>

<div class="flex flex-col gap-4">
	<ul class="flex flex-col gap-2">
		{#each used as [note, count] (note)}
			<li class="flex flex-wrap items-center gap-x-3 gap-y-1">
				<span class="w-44 min-w-0 text-sm">
					<span class="text-fg">{noteLabel(note)}</span>
					<span class="text-fg-subtle font-mono text-xs">#{note}</span>
					<span class="text-fg-subtle text-xs"> · {count}×</span>
				</span>
				<ArrowRight size={14} class="text-fg-subtle shrink-0" aria-hidden="true" />
				<label class="sr-only" for="bulk-{note}">What to do with {noteLabel(note)} (#{note})</label>
				<select
					id="bulk-{note}"
					class={selectClass}
					value={choice[note] ?? ''}
					onchange={(e) => (choice[note] = e.currentTarget.value)}
				>
					<option value="">Keep</option>
					{#each targets as t (t)}
						{#if t !== note}<option value={String(t)}>{noteLabel(t)} #{t} · {drumGroup(t)}</option>{/if}
					{/each}
					<option value="custom">Other note number…</option>
					<option value="delete">Delete all</option>
				</select>
				{#if choice[note] === 'custom'}
					<label class="sr-only" for="bulk-custom-{note}">Note number for {noteLabel(note)}</label>
					<input
						id="bulk-custom-{note}"
						type="number"
						min="0"
						max="127"
						bind:value={custom[note]}
						placeholder="0–127"
						class="w-20 h-9 bg-control border border-line rounded-lg px-2 text-sm text-fg focus:outline-none focus:ring-2 focus:ring-accent"
					/>
				{/if}
			</li>
		{/each}
	</ul>

	{#if after}
		<div class="border border-line rounded-lg overflow-hidden" aria-live="polite">
			<table class="w-full text-sm">
				<caption class="sr-only">Hits per note before and after</caption>
				<thead>
					<tr class="text-xs text-fg-muted border-b border-line">
						<th scope="col" class="text-left font-medium px-3 py-2">Note</th>
						<th scope="col" class="text-right font-medium px-3 py-2">Before</th>
						<th scope="col" class="text-right font-medium px-3 py-2">After</th>
					</tr>
				</thead>
				<tbody>
					{#each previewRows as n (n)}
						{@const b = used.find(([u]) => u === n)?.[1] ?? 0}
						{@const a = after.get(n) ?? 0}
						<tr class="border-b border-line-subtle last:border-0">
							<th scope="row" class="text-left font-normal px-3 py-1.5 text-fg">
								{noteLabel(n)} <span class="text-fg-subtle font-mono text-xs">#{n}</span>
							</th>
							<td class="text-right px-3 py-1.5 tabular-nums text-fg-muted">{b}</td>
							<td class="text-right px-3 py-1.5 tabular-nums {a !== b ? 'text-accent-text font-medium' : 'text-fg-muted'}">{a}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<p class="text-xs text-fg-subtle">Hits that end up on the same note at the same time are merged into one.</p>
	{/if}

	<div>
		<Button variant="secondary" onclick={apply} disabled={!after}>Apply to track</Button>
	</div>
</div>
