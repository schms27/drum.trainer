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
	created_at: string;
}

export interface TrackWithNotes extends Track {
	notes: NoteEvent[];
}
