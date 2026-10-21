import { describe, expect, it } from 'vitest';
import { buildCalendar, isoDate, levelFor } from './calendar';

describe('practice calendar', () => {
	it('maps minutes to shading levels', () => {
		expect([0, 1, 9.9, 10, 19, 20, 39, 40, 120].map(levelFor)).toEqual([0, 1, 1, 2, 2, 3, 3, 4, 4]);
	});

	it('lays out weeks Monday to Sunday ending with this week', () => {
		const today = new Date(2026, 9, 21); // Wednesday 21 Oct 2026
		const weeks = buildCalendar([{ date: '2026-10-20', session_count: 2, minutes_played: 25 }], today, 53, 'en');
		expect(weeks).toHaveLength(53);
		const last = weeks.at(-1)!;
		expect(last.days[0].date).toBe('2026-10-19'); // Monday
		expect(last.days[6].date).toBe('2026-10-25'); // Sunday
		expect(last.days[1]).toMatchObject({ minutes: 25, takes: 2, level: 3, future: false });
		expect(last.days[3].future).toBe(true); // Thursday is after today
		expect(weeks[0].days[0].date).toBe('2025-10-20');
	});

	it('never skips or repeats a day across DST changes and year ends', () => {
		const weeks = buildCalendar([], new Date(2026, 9, 21), 53);
		const dates = weeks.flatMap((w) => w.days.map((d) => d.date));
		expect(new Set(dates).size).toBe(dates.length);
		for (let i = 1; i < dates.length; i++) {
			const [y, m, d] = dates[i - 1].split('-').map(Number);
			expect(dates[i]).toBe(isoDate(new Date(y, m - 1, d + 1)));
		}
	});

	it('labels the week where a month starts', () => {
		// 8 weeks back starts on Mon 31 Aug: that column holds 1 Sep.
		const labels = (n: number) => buildCalendar([], new Date(2026, 9, 21), n, 'en').map((w) => w.monthLabel);
		expect(labels(8).filter(Boolean)).toEqual(['Sep', 'Oct']);
		// 7 weeks back starts on Mon 7 Sep, far enough from 1 Oct to name September.
		expect(labels(7)).toEqual(['Sep', '', '', 'Oct', '', '', '']);
		// 6 weeks back starts on Mon 14 Sep: "Sep" would collide with "Oct".
		expect(labels(6)).toEqual(['', '', 'Oct', '', '', '']);
	});
});
