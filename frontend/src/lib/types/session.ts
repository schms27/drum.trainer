export interface RecordedHit {
	note_number: number;
	time_ms: number;
	velocity: number;
}

export interface MatchedHit {
	target_time_ms: number;
	recorded_time_ms: number;
	offset_ms: number; // negative = early, positive = late
	note_number: number;
	instrument: string;
}

export interface InstrumentBreakdown {
	instrument: string;
	target_count: number;
	matched_count: number;
	missed_count: number;
	ghost_count: number;
	hit_rate: number;
	avg_offset_ms: number | null;
	std_offset_ms: number | null;
}

export interface ScoreResult {
	hit_rate: number;
	miss_rate: number;
	ghost_rate: number;
	avg_offset_ms: number;
	std_offset_ms: number;
	by_instrument: InstrumentBreakdown[];
	matched_hits: MatchedHit[];
}

export interface Session {
	id: number;
	track_id: number;
	started_at: string;
	ended_at: string;
	duration_ms: number;
	hit_count: number;
	created_at: string;
}

export interface SessionSummary {
	id: number;
	track_id: number;
	hit_count: number;
	hit_rate: number;
	avg_offset_ms: number;
	created_at: string;
}

export interface SessionWithScore {
	session: Session;
	score: ScoreResult;
}

export interface SubmitSessionRequest {
	track_id: number;
	started_at: string;
	ended_at: string;
	hits: RecordedHit[];
}
