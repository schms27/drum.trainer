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
	/** Saved metronome level (0–1) for this track; null = the kit profile's default. */
	metronome_volume: number | null;
	created_at: string;
	visibility: Visibility;
	/** Link token while the track is shared; null when private. */
	share_token: string | null;
	rights_basis: RightsBasis | null;
	/** Set when this track is a copy of someone else's shared track. */
	source_track_id: number | null;
	/** True for every copy of someone else's track, even after the original is deleted. */
	copied: boolean;
	/** Removed by moderation (or copied from a removed track): no notes, no practice. */
	disabled: boolean;
	/** Why it was removed, as written by the moderator. */
	disabled_reason: string | null;
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

export type Visibility = 'private' | SharedVisibility;

/** 'unlisted': only people with the link; 'public': also listed in Explore. */
export type SharedVisibility = 'unlisted' | 'public';

export type RightsBasis = 'original' | 'licensed' | 'public_domain' | 'transcription';

/** What anyone with a share link sees: metadata and the opening bars only. */
export interface SharedTrack {
	name: string;
	duration_ms: number;
	note_count: number;
	bpm: number;
	time_sig_numerator: number;
	rights_basis: RightsBasis;
	visibility: SharedVisibility;
	shared_at: string;
	/** Note count per instrument group. */
	instruments: Record<string, number>;
	preview_notes: NoteEvent[];
	/** Length of the preview in ms. */
	preview_ms: number;
	is_owner: boolean;
	/** The viewer's existing copy, if they already added it. */
	copied_track_id: number | null;
}

/** A track in the Explore gallery. */
export interface PublicTrack {
	share_token: string;
	name: string;
	duration_ms: number;
	note_count: number;
	bpm: number;
	time_sig_numerator: number;
	rights_basis: RightsBasis;
	listed_at: string;
	/** How many people added it to their library. */
	copies: number;
}
