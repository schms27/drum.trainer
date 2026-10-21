import { apiFetch } from './client';

export type WeekStat = {
	week_start: string; // "2025-01-06" — Monday of that week
	session_count: number;
	minutes_played: number;
	tracks_practiced: number;
};

export type DayStat = {
	date: string; // "2026-10-21" in the browser's time zone
	session_count: number;
	minutes_played: number;
};

export type StatsData = {
	weekly: WeekStat[];
	/** Days with at least one take in the past year, oldest first. */
	daily: DayStat[];
	totals: {
		sessions: number;
		minutes_played: number;
		tracks_practiced: number;
	};
	/** Consecutive days with practice, ending today (or yesterday). */
	streak_days: number;
	practiced_today: boolean;
};

/** Fetch weekly stats, bucketed into weeks in the browser's time zone. */
export async function fetchStats(): Promise<StatsData> {
	const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
	return apiFetch<StatsData>(tz ? `/stats?tz=${encodeURIComponent(tz)}` : '/stats');
}
