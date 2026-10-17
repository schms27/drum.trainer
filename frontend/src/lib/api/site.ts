import { apiFetch } from './client';

/** Public site settings configured on the backend. */
export interface SiteInfo {
	/** Address for copyright notices; empty when not configured. */
	contact_email: string;
}

export async function getSiteInfo(): Promise<SiteInfo> {
	try {
		return await apiFetch<SiteInfo>('/site');
	} catch {
		return { contact_email: '' };
	}
}
