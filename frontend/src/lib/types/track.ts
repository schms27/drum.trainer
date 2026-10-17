export interface NoteEvent {
	note_number: number;
	time_ms: number;
	velocity: number;
	instrument: string;
}

export interface Track {
	id: number;
	name: string;
	filename: string;
	file_size: number;
	duration_ms: number;
	note_count: number;
	ppq: number;
	bpm: number;
	time_sig_numerator: number;
	practice_bpm: number | null;
	created_at: string;
	/** Practice progress on this track (list and detail endpoints). */
	stats?: TrackStats;
}

export interface TrackStats {
	sessions: number;
	best_hit_rate: number | null;
	last_hit_rate: number | null;
	last_practiced_at: string | null;
	/** Hit rates of the latest takes (up to 10), oldest first. */
	recent_hit_rates: number[];
}

export interface TrackWithNotes extends Track {
	notes: NoteEvent[];
}

export interface NoteMapping {
	note_number: number;
	instrument: string;
}
