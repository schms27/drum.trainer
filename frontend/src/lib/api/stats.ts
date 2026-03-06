import { apiFetch } from './client';

export type WeekStat = {
	week_start: string; // "2025-01-06" — Monday of that week
	session_count: number;
	minutes_played: number;
	tracks_practiced: number;
};

export type StatsData = {
	weekly: WeekStat[];
	totals: {
		sessions: number;
		minutes_played: number;
		tracks_practiced: number;
	};
};

export async function fetchStats(): Promise<StatsData> {
	return apiFetch<StatsData>('/stats');
}
