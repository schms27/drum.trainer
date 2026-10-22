import type { HandleClientError } from '@sveltejs/kit';
import { ApiError, isNetworkError } from '$lib/api/client';

// Unexpected errors in load functions would otherwise all show as
// "Internal Error". Pass on what the API said, and mark connection problems
// so +error.svelte can show the offline page.
export const handleError: HandleClientError = ({ error, message }) => {
	if (isNetworkError(error)) return { message: (error as ApiError).message, offline: true };
	if (error instanceof ApiError) return { message: error.message };
	return { message };
};
