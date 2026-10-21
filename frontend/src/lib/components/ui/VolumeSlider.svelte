<script lang="ts">
	// A 0–1 level slider showing whole percent.
	let {
		id,
		label,
		value = $bindable(1),
		disabled = false,
		oninput
	}: {
		id: string;
		label: string;
		value?: number;
		disabled?: boolean;
		oninput?: (value: number) => void;
	} = $props();

	const pct = $derived(Math.round(value * 100));
</script>

<div class="flex items-center gap-3 min-w-0 flex-1">
	<label for={id} class="sr-only">{label}</label>
	<input
		{id}
		type="range"
		min="0"
		max="100"
		step="1"
		value={pct}
		{disabled}
		aria-valuetext="{pct} %"
		oninput={(e) => {
			value = Number(e.currentTarget.value) / 100;
			oninput?.(value);
		}}
		class="flex-1 min-w-0 accent-[var(--accent)] cursor-pointer disabled:cursor-not-allowed disabled:opacity-40"
	/>
	<span class="w-11 text-right text-xs tabular-nums text-fg-muted" aria-hidden="true">{pct} %</span>
</div>
