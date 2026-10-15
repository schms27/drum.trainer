import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiFetch, ApiError } from '$lib/api/client';

describe('apiFetch', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	it('returns parsed JSON on 200', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			json: () => Promise.resolve({ data: 'hello' }),
		}));

		const result = await apiFetch<{ data: string }>('/test');
		expect(result).toEqual({ data: 'hello' });
	});

	it('prepends /api/v1 to the path', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			json: () => Promise.resolve({}),
		});
		vi.stubGlobal('fetch', mockFetch);

		await apiFetch('/tracks');

		const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
		expect(url).toBe('/api/v1/tracks');
	});

	it('sends Content-Type: application/json', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			json: () => Promise.resolve({}),
		});
		vi.stubGlobal('fetch', mockFetch);

		await apiFetch('/test');

		const [, init] = mockFetch.mock.calls[0] as [string, RequestInit];
		expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
	});

	it('returns undefined for 204 No Content', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			status: 204,
			json: () => Promise.resolve(null),
		}));

		const result = await apiFetch('/test');
		expect(result).toBeUndefined();
	});

	it('throws ApiError with status and message from body on error', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 404,
			json: () => Promise.resolve({ error: 'track not found' }),
		}));

		const err = (await apiFetch('/test').catch((e) => e)) as ApiError;
		expect(err).toBeInstanceOf(ApiError);
		expect(err.status).toBe(404);
		expect(err.message).toBe('track not found');
	});

	it('falls back to "HTTP <status>" when error body has no .error field', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 500,
			json: () => Promise.resolve({ something: 'else' }),
		}));

		const err = (await apiFetch('/test').catch((e) => e)) as ApiError;
		expect(err).toBeInstanceOf(ApiError);
		expect(err.message).toBe('HTTP 500');
	});

	it('falls back to "HTTP <status>" when error body is not JSON', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 503,
			json: () => Promise.reject(new SyntaxError('not JSON')),
		}));

		const err = (await apiFetch('/test').catch((e) => e)) as ApiError;
		expect(err).toBeInstanceOf(ApiError);
		expect(err.status).toBe(503);
		expect(err.message).toBe('HTTP 503');
	});

	it('merges caller-supplied headers with the default Content-Type', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			json: () => Promise.resolve({}),
		});
		vi.stubGlobal('fetch', mockFetch);

		await apiFetch('/test', { headers: { Authorization: 'Bearer tok' } });

		const [, init] = mockFetch.mock.calls[0] as [string, RequestInit];
		const headers = init.headers as Record<string, string>;
		expect(headers['Content-Type']).toBe('application/json');
		expect(headers['Authorization']).toBe('Bearer tok');
	});
});
