<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import { listenToInput } from '$lib/midi/input';
	import { playClick } from '$lib/midi/metronome';
	import { audioTimeToPerfMs, computeCalibration, type CalibrationResult } from '$lib/midi/timing';
	import { practiceSettings } from '$lib/practice-settings.svelte';

	let { open = $bindable(false), deviceId }: { open?: boolean; deviceId: string } = $props();

	const BPM = 90;
	const CLICKS = 16;
	const WARMUP = 4;
	const BEAT_S = 60 / BPM;

	type Phase = 'intro' | 'running' | 'done' | 'failed';

	let dialog: HTMLDialogElement;
	let phase = $state<Phase>('intro');
	let beat = $state(-1);
	let result = $state<CalibrationResult | null>(null);
	let error = $state('');

	let ctx: AudioContext | null = null;
	let stopListening: (() => void) | undefined;
	let ticker: ReturnType<typeof setInterval> | undefined;

	const current = $derived(deviceId ? practiceSettings.latencyFor(deviceId) : null);

	$effect(() => {
		if (open && !dialog.open) {
			phase = 'intro';
			result = null;
			error = '';
			dialog.showModal();
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});

	function cleanup() {
		clearInterval(ticker);
		stopListening?.();
		stopListening = undefined;
		ctx?.close();
		ctx = null;
		beat = -1;
	}

	async function run() {
		error = '';
		result = null;
		const taps: number[] = [];
		try {
			stopListening = await listenToInput(deviceId, (hit) => taps.push(hit.timeStamp));
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Could not open the MIDI device.';
			phase = 'failed';
			return;
		}

		ctx = new AudioContext();
		await ctx.resume();
		const first = ctx.currentTime + 0.6;
		const clickAudioTimes = Array.from({ length: CLICKS }, (_, i) => first + i * BEAT_S);
		clickAudioTimes.forEach((t, i) => playClick(ctx!, t, i % 4 === 0));
		phase = 'running';

		ticker = setInterval(() => {
			if (!ctx) return;
			beat = Math.floor((ctx.currentTime - first) / BEAT_S);
			if (ctx.currentTime > first + CLICKS * BEAT_S + 0.3) finish(clickAudioTimes, taps);
		}, 50);
	}

	function finish(clickAudioTimes: number[], taps: number[]) {
		if (!ctx) return;
		// Map the click times onto the MIDI clock now that the audio output
		// timestamp is settled.
		const clickPerf = clickAudioTimes.map((t) => audioTimeToPerfMs(ctx!, t));
		cleanup();
		result = computeCalibration(taps, clickPerf, BEAT_S * 1000, WARMUP);
		phase = result ? 'done' : 'failed';
		if (!result) error = 'Not enough hits lined up with the clicks. Try again and hit a pad on every click.';
	}

	function save() {
		if (!result) return;
		// Guard against nonsense values from a botched run.
		practiceSettings.setLatency(deviceId, Math.max(-100, Math.min(250, result.offsetMs)));
		open = false;
	}

	function close() {
		cleanup();
		open = false;
	}
</script>

<dialog
	bind:this={dialog}
	onclose={close}
	class="m-auto w-[calc(100%-2rem)] max-w-md rounded-xl border border-line bg-card text-fg p-0 shadow-xl backdrop:bg-black/50"
>
	<div class="p-6 flex flex-col gap-4">
		<h2 class="text-lg font-semibold">Calibrate timing</h2>

		{#if phase === 'intro'}
			<p class="text-sm text-fg-muted">
				Your kit, USB and speakers each add a small delay. To measure it, you'll hear {CLICKS} clicks at
				{BPM} BPM. Hit any pad exactly on each click — the first {WARMUP} are for warming up.
			</p>
			<p class="text-sm text-fg-muted">Use the same speakers or headphones you practise with.</p>
			{#if current !== null}
				<p class="text-sm text-fg">Current correction: <span class="font-medium tabular-nums">{current} ms</span></p>
			{/if}
		{:else if phase === 'running'}
			<p class="text-sm text-fg-muted">Hit a pad on every click.</p>
			<div class="flex flex-wrap gap-1.5" aria-hidden="true">
				{#each { length: CLICKS } as _, i}
					<span
						class="w-4 h-4 rounded-full transition-colors
							{i === beat ? 'bg-accent scale-125' : i < beat ? (i < WARMUP ? 'bg-control-hover' : 'bg-accent/50') : 'bg-control'}"
					></span>
				{/each}
			</div>
			<p class="text-sm text-fg-subtle" role="status">
				{beat < 0 ? 'Get ready…' : beat < WARMUP ? `Warm-up ${beat + 1} of ${WARMUP}` : `Click ${beat + 1} of ${CLICKS}`}
			</p>
		{:else if phase === 'done' && result}
			<p class="text-sm text-fg">
				Your hits arrive about <span class="font-semibold tabular-nums">{result.offsetMs} ms</span>
				{result.offsetMs >= 0 ? 'after' : 'before'} the click
				<span class="text-fg-subtle">(±{result.spreadMs} ms, {result.count} hits)</span>.
			</p>
			{#if result.spreadMs > 30}
				<Notice variant="warning">Your hits varied quite a bit. Consider trying again for a more reliable value.</Notice>
			{/if}
			<p class="text-sm text-fg-muted">This correction is applied to takes that start with a count-in.</p>
		{:else}
			<Notice variant="error">{error}</Notice>
		{/if}

		<div class="flex flex-wrap justify-end gap-2 mt-2">
			{#if phase === 'done'}
				<Button variant="ghost" onclick={() => (phase = 'intro')}>Try again</Button>
				<Button variant="primary" onclick={save}>Use {result?.offsetMs} ms</Button>
			{:else if phase === 'running'}
				<Button variant="ghost" onclick={close}>Cancel</Button>
			{:else}
				{#if current !== null}
					<Button
						variant="ghost"
						onclick={() => {
							practiceSettings.setLatency(deviceId, null);
							open = false;
						}}>Remove correction</Button
					>
				{/if}
				<Button variant="ghost" onclick={close}>Close</Button>
				<Button variant="primary" onclick={run} disabled={!deviceId}>{phase === 'failed' ? 'Try again' : 'Start'}</Button>
			{/if}
		</div>
	</div>
</dialog>
