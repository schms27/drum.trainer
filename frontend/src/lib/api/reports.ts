import { apiFetch } from './client';

export type ReportKind = 'copyright' | 'abuse' | 'other';
export type ReportStatus = 'open' | 'actioned' | 'rejected';

export interface NewReport {
	/** Share token or full share link. */
	token: string;
	kind: ReportKind;
	reporter_name: string;
	reporter_email: string;
	work_identified: string;
	details: string;
	good_faith: boolean;
}

export async function submitReport(report: NewReport): Promise<number> {
	const data = await apiFetch<{ id: number }>('/reports', { method: 'POST', body: JSON.stringify(report) });
	return data.id;
}

/** A report as the admin queue shows it. */
export interface AdminReport {
	id: number;
	track_id: number;
	track_name: string;
	share_token: string;
	reporter_name: string;
	reporter_email: string;
	kind: ReportKind;
	work_identified: string;
	details: string;
	good_faith: boolean;
	status: ReportStatus;
	resolution_note: string | null;
	created_at: string;
	resolved_at: string | null;
	track_exists: boolean;
	track_disabled: boolean;
	rights_basis: string | null;
	copies: number;
	owner_name: string | null;
	owner_email: string | null;
	owner_strikes: number;
	owner_banned: boolean;
}

export interface ModerationResult {
	tracks_affected: number;
	owner_strikes: number;
	owner_banned: boolean;
}

export async function listReports(status: ReportStatus | 'all'): Promise<AdminReport[]> {
	const data = await apiFetch<{ reports: AdminReport[] }>(`/admin/reports?status=${status}`);
	return data.reports;
}

export async function rejectReport(id: number, note: string): Promise<void> {
	await apiFetch(`/admin/reports/${id}/reject`, { method: 'POST', body: JSON.stringify({ note }) });
}

export async function disableTrack(trackId: number, reason: string, strike: boolean): Promise<ModerationResult> {
	return apiFetch<ModerationResult>(`/admin/tracks/${trackId}/disable`, {
		method: 'POST',
		body: JSON.stringify({ reason, strike })
	});
}

export async function restoreTrack(trackId: number, note: string, revokeStrike: boolean): Promise<ModerationResult> {
	return apiFetch<ModerationResult>(`/admin/tracks/${trackId}/restore`, {
		method: 'POST',
		body: JSON.stringify({ note, revoke_strike: revokeStrike })
	});
}
