<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// Keep in sync with TermsVersion in backend/api/sharing.go.
	const version = '17 October 2026';
	const contact = $derived(data.site.contact_email);
</script>

<svelte:head><title>Terms of use — Drum Trainer</title></svelte:head>

<PageHeader title="Terms of use" subtitle="Last updated {version}" />

<article class="max-w-2xl flex flex-col gap-6 text-sm text-fg leading-relaxed [&_h2]:text-base [&_h2]:font-semibold [&_h2]:mb-2 [&_ul]:list-disc [&_ul]:pl-5 [&_ul]:flex [&_ul]:flex-col [&_ul]:gap-1">
	<section>
		<h2>Your tracks</h2>
		<p>
			You keep all rights to the tracks you upload. Tracks are private unless you share them. We store them only to
			run Drum Trainer for you, and you can download your own uploads again at any time.
		</p>
	</section>

	<section>
		<h2>Sharing</h2>
		<p>When you create a share link for a track:</p>
		<ul>
			<li>anyone with the link can see the track’s name, tempo and opening bars, and signed-in users can add a copy of its drum notes to their own library;</li>
			<li>only the drum notes are shared, not your original MIDI file, backing tracks or practice history;</li>
			<li>people who add a copy can practise it in Drum Trainer but can’t download it as a file;</li>
			<li>copies people have already added stay in their libraries if you reset the link or stop sharing.</li>
		</ul>
	</section>

	<section>
		<h2>Your responsibility for what you share</h2>
		<p>By sharing a track you confirm that:</p>
		<ul>
			<li>you wrote it yourself, the music is in the public domain, or you have permission from the rights holder to share it;</li>
			<li>sharing it doesn’t infringe anyone’s copyright or other rights.</li>
		</ul>
		<p class="mt-2">
			You give Drum Trainer permission to store the shared track and show it to people who have the link, only as far
			as needed to provide sharing. You’re responsible for what you share, and you agree to cover any costs we incur
			because a track you shared infringed someone’s rights.
		</p>
		<p class="mt-2">
			Transcribing a copyrighted song for your own practice is one thing; sharing that transcription needs the rights
			holder’s permission. When in doubt, keep the track private.
		</p>
	</section>

	<section>
		<h2>Removal</h2>
		<p>
			We may turn off a share link or remove a shared track, together with copies made from it, if we receive a
			credible notice that it infringes someone’s rights, or if it breaks these terms. Users who repeatedly share
			infringing tracks lose the ability to share.
		</p>
	</section>

	<section id="report">
		<h2>Reporting a copyright problem</h2>
		<p>
			If you believe a shared track infringes your copyright, use the
			<a href="/report" class="text-accent-text hover:underline">report form</a> — no account needed. Our
			<a href="/copyright" class="text-accent-text hover:underline">copyright policy</a> explains what happens next.
			{#if contact}
				You can also write to <a href="mailto:{contact}" class="text-accent-text hover:underline">{contact}</a>.
			{/if}
		</p>
	</section>
</article>
