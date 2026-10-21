<script lang="ts">
	import { buildCalendar, LEVEL_LIMITS, type CalendarDay, type DayStat } from '$lib/calendar';

	// One tile per day for the past year, darker the longer you played.
	let { daily }: { daily: DayStat[] } = $props();

	const weeks = $derived(buildCalendar(daily, new Date()));
	const practisedDays = $derived(daily.filter((d) => d.minutes_played > 0).length);
	const totalMinutes = $derived(Math.round(daily.reduce((m, d) => m + d.minutes_played, 0)));

	// Monday-first weekday names; only Mon, Wed and Fri are labelled.
	const weekdayLabels = $derived(
		Array.from({ length: 7 }, (_, i) =>
			i % 2 === 0 && i < 6 ? new Date(2024, 0, 1 + i).toLocaleString(undefined, { weekday: 'short' }) : ''
		)
	);

	// Tiles mix the accent text colour (lighter in dark mode) into the
	// empty-tile colour, in evenly spaced steps.
	const SHADE = ['', '25%', '50%', '75%', '100%'];
	function tileStyle(day: CalendarDay) {
		if (day.future) return 'background: transparent';
		if (day.level === 0) return 'background: var(--control)';
		return `background: color-mix(in srgb, var(--accent-text) ${SHADE[day.level]}, var(--control))`;
	}

	function describe(day: CalendarDay) {
		const date = new Date(`${day.date}T12:00:00`).toLocaleDateString(undefined, {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
		if (day.takes === 0) return `${date}: no practice`;
		const min = Math.max(1, Math.round(day.minutes));
		return `${date}: ${min} min, ${day.takes} ${day.takes === 1 ? 'take' : 'takes'}`;
	}

	// Start scrolled to today on narrow screens.
	let scroller: HTMLDivElement;
	$effect(() => {
		void weeks;
		if (scroller) scroller.scrollLeft = scroller.scrollWidth;
	});

	const legend = [
		{ level: 0, label: 'No practice' },
		{ level: 1, label: `Under ${LEVEL_LIMITS[0]} min` },
		{ level: 2, label: `${LEVEL_LIMITS[0]}–${LEVEL_LIMITS[1]} min` },
		{ level: 3, label: `${LEVEL_LIMITS[1]}–${LEVEL_LIMITS[2]} min` },
		{ level: 4, label: `${LEVEL_LIMITS[2]}+ min` }
	] as const;
</script>

<div class="bg-card border border-line rounded-xl p-5 flex flex-col gap-3">
	<div class="flex flex-wrap items-baseline justify-between gap-2">
		<h2 class="font-semibold text-fg">Practice calendar</h2>
		<p class="text-xs text-fg-subtle">
			{practisedDays} {practisedDays === 1 ? 'day' : 'days'} with practice · {totalMinutes} min in the past year
		</p>
	</div>

	<div class="flex gap-2">
		<div class="grid grid-rows-7 gap-[3px] pt-5 shrink-0 text-[10px] leading-none text-fg-subtle" aria-hidden="true">
			{#each weekdayLabels as label, i (i)}
				<span class="h-3 flex items-center">{label}</span>
			{/each}
		</div>
		<div bind:this={scroller} class="overflow-x-auto min-w-0 pb-1">
			<div
				class="inline-grid grid-flow-col gap-[3px]"
				role="img"
				aria-label="Practice calendar: {practisedDays} days with practice, {totalMinutes} minutes in the past year"
			>
				{#each weeks as week (week.days[0].date)}
					<div class="grid grid-rows-[1rem_repeat(7,0.75rem)] gap-[3px]">
						<span class="text-[10px] leading-none text-fg-subtle whitespace-nowrap w-3 overflow-visible" aria-hidden="true">
							{week.monthLabel}
						</span>
						{#each week.days as day (day.date)}
							<span class="w-3 h-3 rounded-[3px]" style={tileStyle(day)} title={day.future ? undefined : describe(day)}></span>
						{/each}
					</div>
				{/each}
			</div>
		</div>
	</div>

	<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-fg-subtle">
		{#each legend as l (l.level)}
			<span class="inline-flex items-center gap-1.5">
				<span class="w-3 h-3 rounded-[3px]" style={tileStyle({ level: l.level, future: false } as CalendarDay)}></span>
				{l.label}
			</span>
		{/each}
	</div>
</div>
