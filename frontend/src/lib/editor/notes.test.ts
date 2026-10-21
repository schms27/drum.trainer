import { describe, expect, it } from 'vitest';
import {
	History,
	applyNoteActions,
	barsFor,
	countByNote,
	mergeDuplicates,
	moveNotes,
	repeatBars,
	snap,
	stepMs,
	withIds,
	type NoteAction
} from './notes';

const n = (note_number: number, time_ms: number, velocity = 100) => ({ note_number, time_ms, velocity });
const strip = (notes: { note_number: number; time_ms: number; velocity: number }[]) =>
	notes.map(({ note_number, time_ms, velocity }) => ({ note_number, time_ms, velocity }));

describe('editor notes', () => {
	it('maps several cymbals to one and merges hits that land together', () => {
		// The simplification from WoE_Intro_Drums: #40 → #38, hi-hats/ride/crash → #30.
		const notes = withIds([n(42, 0, 70), n(49, 0, 120), n(40, 500), n(51, 500, 90), n(53, 500, 60), n(36, 0)]);
		const actions = new Map<number, NoteAction>([
			[40, 38],
			[42, 30],
			[49, 30],
			[51, 30],
			[53, 30]
		]);
		const out = applyNoteActions(notes, actions);
		expect(strip(out)).toEqual([n(30, 0, 120), n(36, 0), n(30, 500, 90), n(38, 500)]);
		expect(countByNote(out)).toEqual(
			new Map([
				[30, 2],
				[36, 1],
				[38, 1]
			])
		);
	});

	it('deletes every hit of a note number', () => {
		const out = applyNoteActions(withIds([n(45, 0), n(36, 0), n(45, 250)]), new Map([[45, 'delete']]));
		expect(strip(out)).toEqual([n(36, 0)]);
	});

	it('keeps separate hits that are just apart', () => {
		expect(mergeDuplicates(withIds([n(38, 0), n(38, 15)]))).toHaveLength(2);
	});

	it('snaps to the grid', () => {
		expect(stepMs(120, '1/16')).toBe(125);
		expect(stepMs(120, '1/8T')).toBeCloseTo(166.667, 2);
		expect(snap(190, 125)).toBe(250);
		expect(snap(-40, 125)).toBe(0);
	});

	it('moves notes in time and between rows, clamped to the rows', () => {
		const notes = withIds([n(36, 0), n(38, 500)]);
		const rows = [49, 38, 36];
		const moved = moveNotes(notes, new Set([notes[0].id]), 250, -1, rows);
		expect(strip(moved)).toEqual([n(38, 250), n(38, 500)]);
		const clamped = moveNotes(notes, new Set([notes[1].id]), -1000, -5, rows);
		expect(strip(clamped)).toEqual([n(36, 0), n(49, 0)]);
	});

	it('counts bars', () => {
		expect(barsFor(16000, 120, 4)).toBe(8);
		expect(barsFor(16001, 120, 4)).toBe(9);
		expect(barsFor(0, 120, 4)).toBe(1);
	});

	it('repeats a range of bars and pushes later bars back', () => {
		// 2 s bars: one hit in bar 1, one in bar 2, one in bar 3.
		const notes = withIds([n(36, 0), n(38, 2500), n(42, 4000)]);
		const { notes: out, addedBars } = repeatBars(notes, 2000, 1, 2, 2);
		expect(addedBars).toBe(4);
		expect(strip(out)).toEqual([n(36, 0), n(38, 2500), n(36, 4000), n(38, 6500), n(36, 8000), n(38, 10500), n(42, 12000)]);
		expect(repeatBars(notes, 2000, 3, 2, 1).addedBars).toBe(0);
	});

	it('undoes and redoes', () => {
		const h = new History();
		const a = withIds([n(36, 0)]);
		const b = withIds([n(36, 0), n(38, 500)]);
		h.push(a);
		expect(h.undo(b)).toBe(a);
		expect(h.redo(a)).toBe(b);
		expect(h.canRedo).toBe(false);
		h.undo(b);
		h.push(a); // a new change clears the redo stack
		expect(h.canRedo).toBe(false);
	});
});
