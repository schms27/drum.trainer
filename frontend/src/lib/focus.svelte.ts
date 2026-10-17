// Focus mode: while a take is being recorded the page chrome (header,
// footer) is hidden so the drummer sees only what matters.

let active = $state(false);

export const focusMode = {
	get active() {
		return active;
	},
	set active(v: boolean) {
		active = v;
	}
};
