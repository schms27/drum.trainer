import { error } from '@sveltejs/kit';
import { listReports, type ReportStatus } from '$lib/api/reports';
import type { PageLoad } from './$types';

const STATUSES: ReportStatus[] = ['open', 'actioned', 'rejected'];

export const load: PageLoad = async ({ parent, url }) => {
	const { user } = await parent();
	if (!user?.is_admin) error(404, 'Not found');

	const param = url.searchParams.get('status') as ReportStatus | null;
	const status = param && STATUSES.includes(param) ? param : 'open';
	return { status, reports: await listReports(status) };
};
