<script lang="ts">
	import { Trash, Save } from '@lucide/svelte';
	import { createKit, updateKit, type KitProfile } from '$lib/api/kits';
	import { listenToInput } from '$lib/midi/input';
	import { INSTRUMENTS, drumGroup, noteLabel } from '$lib/drums';
	import MidiDeviceSelector from '$lib/components/recorder/MidiDeviceSelector.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	let {
		kit = null,
		onSaved,
		onCancel
	}: {
		/** The kit to edit, or null to create a new one. */
		kit?: KitProfile | null;
		onSaved: (kit: KitProfile) => void;
		onCancel: () => void;
	} = $props();

	interface Row {
		note: number;
		instrument: string;
		velocity: number | null;
	}

	// svelte-ignore state_referenced_locally
	let name = $state(kit?.name ?? 'My kit');
	// svelte-ignore state_referenced_locally
	let rows = $state<Row[]>(
		(kit?.mappings ?? []).map((m) => ({ note: m.note_number, instrument: m.instrument, velocity: null }))
	);
	let deviceId = $state('');
	let lastNote = $state<number | null>(null);
	let flashTimer: ReturnType<typeof setTimeout> | undefined;
	let saving = $state(false);
	let error = $state('');

	// Learn mode: every pad hit adds (or highlights) its row.
	$effect(() => {
		const id = deviceId;
		if (!id) return;
		let stop: (() => void) | undefined;
		let cancelled = false;
		listenToInput(id, (hit) => {
			const existing = rows.find((r) => r.note === hit.note);
			if (existing) existing.velocity = hit.velocity;
			else rows = [...rows, { note: hit.note, instrument: drumGroup(hit.note), velocity: hit.velocity }].sort((a, b) => a.note - b.note);
			lastNote = hit.note;
			clearTimeout(flashTimer);
			flashTimer = setTimeout(() => (lastNote = null), 400);
		})
			.then((s) => {
				if (cancelled) s();
				else stop = s;
			})
			.catch(() => {
				// The device selector reports missing devices.
			});
		return () => {
			cancelled = true;
			stop?.();
			clearTimeout(flashTimer);
		};
	});

	async function save() {
		saving = true;
		error = '';
		const mappings = rows.map((r) => ({ note_number: r.note, instrument: r.instrument }));
		try {
			const saved = kit ? await updateKit(kit.id, name.trim(), mappings) : await createKit(name.trim(), mappings);
			onSaved(saved);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not save the kit';
		} finally {
			saving = false;
		}
	}
</script>

<div class="flex flex-col gap-5">
	<div class="flex flex-col gap-1.5">
		<label for="kit-name" class="text-sm font-medium text-fg">Name</label>
		<input
			id="kit-name"
			bind:value={name}
			maxlength="60"
			placeholder="e.g. Roland TD-17"
			class="bg-control text-fg rounded-lg px-3 h-10 text-sm border border-line max-w-sm focus:outline-none focus:ring-2 focus:ring-accent"
		/>
	</div>

	<div class="flex flex-col gap-2">
		<p class="text-sm font-medium text-fg">Learn your pads</p>
		<p class="text-sm text-fg-muted">
			Choose your kit, then hit each pad once — including rims, hi-hat pedal and cymbal edges. Each pad
			appears below with its standard instrument; change it where your kit differs.
		</p>
		<MidiDeviceSelector bind:selectedId={deviceId} />
	</div>

	{#if rows.length === 0}
		<p class="text-sm text-fg-subtle border border-dashed border-line-strong rounded-lg px-4 py-6 text-center" aria-live="polite">
			{deviceId ? 'Waiting for your first hit…' : 'Connect your kit to start learning pads.'}
		</p>
	{:else}
		<div class="border border-line rounded-lg overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="text-xs text-fg-muted border-b border-line">
						<th scope="col" class="text-left font-medium px-3 py-2">Pad (MIDI note)</th>
						<th scope="col" class="text-left font-medium px-3 py-2">Counts as</th>
						<th scope="col" class="px-3 py-2"><span class="sr-only">Remove</span></th>
					</tr>
				</thead>
				<tbody aria-live="polite">
					{#each rows as row (row.note)}
						<tr class="border-b border-line-subtle last:border-0 transition-colors {lastNote === row.note ? 'bg-accent-soft' : ''}">
							<th scope="row" class="text-left font-normal px-3 py-2">
								<span class="text-fg">{noteLabel(row.note)}</span>
								<span class="text-fg-subtle font-mono text-xs ml-1">#{row.note}</span>
								{#if row.velocity !== null}<span class="text-fg-subtle text-xs ml-2">velocity {row.velocity}</span>{/if}
							</th>
							<td class="px-3 py-2">
								<label class="sr-only" for="kit-instr-{row.note}">Instrument for note {row.note}</label>
								<select
									id="kit-instr-{row.note}"
									bind:value={row.instrument}
									class="bg-control text-fg border border-line rounded-lg px-2 h-9 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
								>
									{#each INSTRUMENTS as instr}
										<option value={instr}>{instr}</option>
									{/each}
								</select>
							</td>
							<td class="px-3 py-2 text-right">
								<button
									type="button"
									onclick={() => (rows = rows.filter((r) => r.note !== row.note))}
									class="w-8 h-8 inline-flex items-center justify-center rounded-lg text-fg-subtle hover:text-bad hover:bg-control cursor-pointer"
									aria-label="Remove note {row.note}"
								>
									<Trash size={14} aria-hidden="true" />
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<p class="text-xs text-fg-subtle">Pads you don't list here use the standard General MIDI instrument.</p>
	{/if}

	{#if error}<Notice variant="error">{error}</Notice>{/if}

	<div class="flex gap-2">
		<Button variant="primary" onclick={save} disabled={saving || !name.trim()}>
			<Save size={16} aria-hidden="true" />
			{saving ? 'Saving…' : 'Save kit'}
		</Button>
		<Button variant="ghost" onclick={onCancel} disabled={saving}>Cancel</Button>
	</div>
</div>
