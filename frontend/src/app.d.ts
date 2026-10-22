// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Error {
			message: string;
			/** The server could not be reached (set in hooks.client.ts). */
			offline?: boolean;
		}
	}
}

export {};
