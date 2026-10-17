<script lang="ts">
	import { page } from '$app/state';
	import { CircleCheck } from '@lucide/svelte';
	import { submitReport, type ReportKind } from '$lib/api/reports';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	const user = page.data.user;
	const prefilled = page.url.searchParams.get('token');

	let link = $state(prefilled ? `${location.origin}/s/${prefilled}` : '');
	let kind = $state<ReportKind>('copyright');
	let name = $state(user?.name ?? '');
	let email = $state(user?.email ?? '');
	let work = $state('');
	let details = $state('');
	let goodFaith = $state(false);

	let busy = $state(false);
	let error = $state('');
	let submitted = $state(false);

	const isCopyright = $derived(kind === 'copyright');
	const canSubmit = $derived(
		link.trim() !== '' &&
			name.trim() !== '' &&
			email.trim() !== '' &&
			details.trim().length >= 10 &&
			(!isCopyright || (work.trim() !== '' && goodFaith))
	);

	const field = 'w-full rounded-lg border border-line bg-canvas px-3 py-2 text-sm text-fg';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			await submitReport({
				token: link,
				kind,
				reporter_name: name,
				reporter_email: email,
				work_identified: isCopyright ? work : '',
				details,
				good_faith: isCopyright && goodFaith
			});
			submitted = true;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Could not send the report.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>Report a shared track — Drum Trainer</title></svelte:head>

<PageHeader title="Report a shared track" subtitle="Tell us about a shared track that infringes your rights or breaks our terms." />

{#if submitted}
	<Card class="max-w-2xl">
		<div class="flex gap-3">
			<CircleCheck size={20} class="text-good shrink-0 mt-0.5" aria-hidden="true" />
			<div class="text-sm">
				<p class="font-medium text-fg">Thanks, we’ve received your report.</p>
				<p class="text-fg-muted mt-1">
					We’ll review it and may contact you at {email}. If the track infringes your rights, we turn off the link
					and remove all copies made from it.
				</p>
			</div>
		</div>
	</Card>
{:else}
	<form onsubmit={submit} class="max-w-2xl flex flex-col gap-5">
		<Card>
			<div class="flex flex-col gap-4">
				<label class="flex flex-col gap-1.5 text-sm">
					<span class="font-medium">Share link</span>
					<input bind:value={link} required placeholder="https://…/s/…" class={field} />
				</label>

				<fieldset class="flex flex-col gap-2 text-sm">
					<legend class="font-medium mb-1.5">What’s the problem?</legend>
					<label class="flex items-center gap-2"><input type="radio" bind:group={kind} value="copyright" class="accent-(--accent)" /> It infringes my copyright (or one I represent)</label>
					<label class="flex items-center gap-2"><input type="radio" bind:group={kind} value="abuse" class="accent-(--accent)" /> It’s abusive or offensive</label>
					<label class="flex items-center gap-2"><input type="radio" bind:group={kind} value="other" class="accent-(--accent)" /> Something else</label>
				</fieldset>

				{#if isCopyright}
					<label class="flex flex-col gap-1.5 text-sm">
						<span class="font-medium">Which work is infringed?</span>
						<input bind:value={work} maxlength="500" placeholder="Song title, artist, and who holds the rights" class={field} />
					</label>
				{/if}

				<label class="flex flex-col gap-1.5 text-sm">
					<span class="font-medium">Details</span>
					<textarea bind:value={details} rows="5" maxlength="5000" class={field}
						placeholder={isCopyright ? 'How the shared track copies the work, and your relationship to the rights holder' : 'What’s wrong with this track'}
					></textarea>
				</label>
			</div>
		</Card>

		<Card title="Your contact details" description="We use these only to handle this report.">
			<div class="grid gap-4 sm:grid-cols-2">
				<label class="flex flex-col gap-1.5 text-sm">
					<span class="font-medium">Name</span>
					<input bind:value={name} required maxlength="200" autocomplete="name" class={field} />
				</label>
				<label class="flex flex-col gap-1.5 text-sm">
					<span class="font-medium">Email</span>
					<input bind:value={email} required type="email" autocomplete="email" class={field} />
				</label>
			</div>
		</Card>

		{#if isCopyright}
			<label class="flex items-start gap-3 text-sm cursor-pointer">
				<input type="checkbox" bind:checked={goodFaith} class="mt-0.5 accent-(--accent)" />
				<span>
					I believe in good faith that this use isn’t authorised by the rights holder or the law, and the information
					in this report is accurate. I am the rights holder or authorised to act for them.
				</span>
			</label>
		{/if}

		{#if error}<Notice variant="error">{error}</Notice>{/if}

		<div class="flex items-center justify-between gap-4">
			<a href="/copyright" class="text-sm text-fg-muted hover:text-fg hover:underline">How we handle reports</a>
			<Button type="submit" variant="primary" disabled={busy || !canSubmit}>Send report</Button>
		</div>
	</form>
{/if}
