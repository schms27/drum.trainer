<script lang="ts">
	import WeeklyBars from '$lib/components/stats/WeeklyBars.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import ProgressTabs from '$lib/components/layout/ProgressTabs.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	const { stats } = data;
	const { weekly, totals } = stats;

	function fmtMinutes(m: number) {
		if (m < 60) return m.toFixed(0) + ' min';
		const h = Math.floor(m / 60);
		const rem = Math.round(m % 60);
		return rem > 0 ? `${h}h ${rem}m` : `${h}h`;
	}
</script>

<svelte:head><title>Practice time — Drum Trainer</title></svelte:head>

<PageHeader title="Progress" subtitle="How much you practised over the past 12 months." />
<ProgressTabs />

<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Time played</p>
		<p class="text-3xl font-bold mt-1 text-fg tabular-nums">{fmtMinutes(totals.minutes_played)}</p>
	</div>
	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Takes recorded</p>
		<p class="text-3xl font-bold mt-1 text-fg tabular-nums">{totals.sessions}</p>
	</div>
	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Different tracks</p>
		<p class="text-3xl font-bold mt-1 text-fg tabular-nums">{totals.tracks_practiced}</p>
	</div>
</div>

<WeeklyBars data={weekly} valueKey="minutes_played" label="Minutes played per week" unit="min" />
<WeeklyBars data={weekly} valueKey="session_count" label="Takes per week" />
<WeeklyBars data={weekly} valueKey="tracks_practiced" label="Different tracks per week" />
