// Light/dark theme preference. "system" follows the OS setting.
// The initial theme is applied by the inline script in app.html; this module
// takes over once the app has loaded.

export type ThemePreference = 'system' | 'light' | 'dark';

const STORAGE_KEY = 'drum-trainer:theme';

function readPreference(): ThemePreference {
	try {
		const v = localStorage.getItem(STORAGE_KEY);
		if (v === 'light' || v === 'dark') return v;
	} catch {
		// Storage unavailable (private mode etc.) — fall back to system.
	}
	return 'system';
}

const media = typeof window !== 'undefined' ? window.matchMedia('(prefers-color-scheme: dark)') : null;

let preference = $state<ThemePreference>(typeof window !== 'undefined' ? readPreference() : 'system');
let systemDark = $state(media?.matches ?? false);

media?.addEventListener('change', (e) => {
	systemDark = e.matches;
	apply();
});

function resolved(): 'light' | 'dark' {
	if (preference === 'system') return systemDark ? 'dark' : 'light';
	return preference;
}

function apply() {
	if (typeof document !== 'undefined') document.documentElement.dataset.theme = resolved();
}

export const theme = {
	get preference() {
		return preference;
	},
	get resolved() {
		return resolved();
	},
	set(value: ThemePreference) {
		preference = value;
		try {
			if (value === 'system') localStorage.removeItem(STORAGE_KEY);
			else localStorage.setItem(STORAGE_KEY, value);
		} catch {
			// Not persisted, but still applied for this visit.
		}
		apply();
	}
};
