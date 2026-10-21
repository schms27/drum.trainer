// Tag colours. The API accepts only these keys (backend models.TagColors).
// Chips mix the hue into the card colour for the background and into the
// text colour for the label, so they stay readable in light and dark mode.

export const TAG_COLORS = {
	gray: '#71717a',
	red: '#dc2626',
	orange: '#ea580c',
	amber: '#d97706',
	green: '#16a34a',
	teal: '#0d9488',
	sky: '#0284c7',
	blue: '#2563eb',
	violet: '#7c3aed',
	pink: '#db2777'
} as const;

export type TagColor = keyof typeof TAG_COLORS;

export const TAG_COLOR_KEYS = Object.keys(TAG_COLORS) as TagColor[];

export function tagHue(color: string): string {
	return TAG_COLORS[color as TagColor] ?? TAG_COLORS.gray;
}

/** Inline style for a tag chip. */
export function chipStyle(color: string): string {
	const hue = tagHue(color);
	return [
		`background: color-mix(in srgb, ${hue} 18%, var(--card))`,
		`color: color-mix(in srgb, ${hue} 55%, var(--fg))`,
		`border-color: color-mix(in srgb, ${hue} 40%, var(--card))`
	].join('; ');
}
