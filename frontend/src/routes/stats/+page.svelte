<script lang="ts">
	import WeeklyBars from '$lib/components/stats/WeeklyBars.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const { stats } = data;
	const { weekly, totals } = stats;

	function fmtMinutes(m: number) {
		if (m < 60) return m.toFixed(0) + ' min';
		const h = Math.floor(m / 60);
		const rem = Math.round(m % 60);
		return rem > 0 ? `${h}h ${rem}m` : `${h}h`;
	}
</script>

<div class="flex flex-col gap-8">
	<div>
		<h1 class="text-2xl font-bold text-white">Your Statistics</h1>
		<p class="text-zinc-400 text-sm mt-1">Past 12 months</p>
	</div>

	<!-- Totals -->
	<div class="grid grid-cols-3 gap-4">
		<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
			<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Sessions</p>
			<p class="text-3xl font-bold text-white">{totals.sessions}</p>
		</div>
		<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
			<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Time Played</p>
			<p class="text-3xl font-bold text-indigo-400">{fmtMinutes(totals.minutes_played)}</p>
		</div>
		<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
			<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Tracks Practiced</p>
			<p class="text-3xl font-bold text-white">{totals.tracks_practiced}</p>
		</div>
	</div>

	<!-- Weekly charts -->
	<WeeklyBars
		data={weekly}
		valueKey="minutes_played"
		label="Minutes played per week"
		color="#6366f1"
		unit="min"
	/>

	<WeeklyBars
		data={weekly}
		valueKey="session_count"
		label="Sessions per week"
		color="#22d3ee"
	/>

	<WeeklyBars
		data={weekly}
		valueKey="tracks_practiced"
		label="Distinct tracks per week"
		color="#a78bfa"
	/>
</div>
