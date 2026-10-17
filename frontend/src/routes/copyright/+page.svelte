<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const contact = $derived(data.site.contact_email);
</script>

<svelte:head><title>Copyright policy — Drum Trainer</title></svelte:head>

<PageHeader title="Copyright policy" subtitle="How we handle reports about shared tracks" />

<article class="max-w-2xl flex flex-col gap-6 text-sm text-fg leading-relaxed [&_h2]:text-base [&_h2]:font-semibold [&_h2]:mb-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_ol]:flex [&_ol]:flex-col [&_ol]:gap-1 [&_ul]:list-disc [&_ul]:pl-5 [&_ul]:flex [&_ul]:flex-col [&_ul]:gap-1">
	<section>
		<h2>What we host</h2>
		<p>
			Drum Trainer users can share the drum notes of their tracks through unlisted links. We never share the
			original MIDI files or any audio. The people who share a track confirm they have the right to, and are
			responsible for it under our <a href="/terms" class="text-accent-text hover:underline">terms of use</a>.
		</p>
	</section>

	<section>
		<h2>Reporting infringement</h2>
		<p>If you hold the rights to a work, or act for someone who does, send a report with:</p>
		<ul>
			<li>the share link of the track;</li>
			<li>the work you believe is infringed;</li>
			<li>your name and an email address we can reach you at;</li>
			<li>a statement that you report in good faith and the information is accurate.</li>
		</ul>
		<div class="mt-3"><Button href="/report" variant="primary">Send a report</Button></div>
		{#if contact}
			<p class="mt-3 text-fg-muted">You can also write to <a href="mailto:{contact}" class="text-accent-text hover:underline">{contact}</a>.</p>
		{/if}
	</section>

	<section>
		<h2>What happens next</h2>
		<ol>
			<li>We review every report promptly.</li>
			<li>If it’s well founded, we turn off the share link and remove the track and every copy made from it. Their owners see the reason we give.</li>
			<li>Tracks with the same notes can’t be shared again afterwards.</li>
			<li>If we decide a report isn’t well founded, we close it and the track stays available.</li>
		</ol>
	</section>

	<section>
		<h2>If your track was removed</h2>
		<p>
			You’ll see the reason on the track’s page. If you think we made a mistake, for example because you hold a
			licence, contact us with the details{#if contact}&nbsp;at <a href="mailto:{contact}" class="text-accent-text hover:underline">{contact}</a>{/if}.
			If we agree, we restore the track and its copies.
		</p>
	</section>

	<section>
		<h2>Repeat infringers</h2>
		<p>
			Each upheld report counts as a strike against the person who shared the track. After three strikes they can no
			longer share tracks, and all their existing share links are turned off.
		</p>
	</section>
</article>
