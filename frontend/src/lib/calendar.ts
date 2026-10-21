// Layout for the practice calendar: one column per week (Monday first), one
// row per weekday, covering the trailing year up to today.

export interface DayStat {
	date: string; // "2026-10-21", in the user's time zone
	session_count: number;
	minutes_played: number;
}

export interface CalendarDay {
	date: string;
	minutes: number;
	takes: number;
	/** 0 = no practice, 1–4 = increasingly long. */
	level: 0 | 1 | 2 | 3 | 4;
	/** After today: drawn empty, without a tooltip. */
	future: boolean;
}

export interface CalendarWeek {
	days: CalendarDay[]; // 7 entries, Monday to Sunday
	/** Short month name when this week starts a new month, else ''. */
	monthLabel: string;
}

/** Upper bounds (minutes, exclusive) of levels 1–3; anything longer is level 4. */
export const LEVEL_LIMITS = [10, 20, 40] as const;

export function levelFor(minutes: number): CalendarDay['level'] {
	if (minutes <= 0) return 0;
	if (minutes < LEVEL_LIMITS[0]) return 1;
	if (minutes < LEVEL_LIMITS[1]) return 2;
	if (minutes < LEVEL_LIMITS[2]) return 3;
	return 4;
}

/** "YYYY-MM-DD" of a local date. */
export function isoDate(d: Date): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/**
 * Weeks from the Monday `weeks - 1` weeks before today's week up to the
 * Sunday of today's week. Dates are built from local calendar fields, so a
 * DST change never skips or repeats a day.
 */
export function buildCalendar(daily: DayStat[], today: Date, weeks = 53, locale?: string): CalendarWeek[] {
	const byDate = new Map(daily.map((d) => [d.date, d]));
	const todayKey = isoDate(today);
	const weekday = (today.getDay() + 6) % 7; // Monday = 0
	const start = new Date(today.getFullYear(), today.getMonth(), today.getDate() - weekday - (weeks - 1) * 7);

	const out: CalendarWeek[] = [];
	for (let w = 0; w < weeks; w++) {
		const days: CalendarDay[] = [];
		let monthLabel = '';
		for (let i = 0; i < 7; i++) {
			const d = new Date(start.getFullYear(), start.getMonth(), start.getDate() + w * 7 + i);
			const key = isoDate(d);
			const stat = byDate.get(key);
			const minutes = stat?.minutes_played ?? 0;
			days.push({ date: key, minutes, takes: stat?.session_count ?? 0, level: levelFor(minutes), future: key > todayKey });
			// Label the week that holds the 1st of a month.
			if (d.getDate() === 1) monthLabel = d.toLocaleString(locale, { month: 'short' });
		}
		out.push({ days, monthLabel });
	}
	// The first column names its month too, unless the next label is so close
	// that the two would overlap.
	if (out.length && !out[0].monthLabel && !out[1]?.monthLabel && !out[2]?.monthLabel) {
		const [y, m, d] = out[0].days[0].date.split('-').map(Number);
		out[0].monthLabel = new Date(y, m - 1, d).toLocaleString(locale, { month: 'short' });
	}
	return out;
}
