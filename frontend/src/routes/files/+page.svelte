<script lang="ts">
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface Directory {
		id: number;
		path: string;
		project_id: number;
		project_name: string;
	}

	let directories = $state<Directory[]>([]);
	let loading = $state(true);

	const breadcrumbSegments = [
		{ label: '', href: '/', icon: 'banner' },
		{ label: 'Files', href: '/files' }
	];

	onMount(async () => {
		await loadDirectories();
	});

	async function loadDirectories() {
		try {
			const res = await fetch('/api/base-directories');
			if (res.ok) {
				directories = await res.json();
			}
		} catch (err) {
			console.error('Failed to load directories:', err);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Files - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<Breadcrumb segments={breadcrumbSegments} />

	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold text-vanna-navy font-serif">File Browser</h1>
		<p class="mt-2 text-slate-500">Browse and edit files in your project directories</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center min-h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div>
		</div>
	{:else if directories.length === 0}
		<Card class="text-center py-12">
			<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
				<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
				</svg>
			</div>
			<h3 class="text-xl font-semibold text-vanna-navy mb-2">No Directories</h3>
			<p class="text-slate-500 mb-4">No base directories have been configured yet.</p>
			<Button href="/" variant="primary">
				Back to Dashboard
			</Button>
		</Card>
	{:else}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each directories as dir}
				<Card class="hover:shadow-md transition-shadow" hover={true}>
					<a href="/files/{dir.id}" class="block">
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 bg-vanna-teal/10 rounded-lg flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
								</svg>
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-sm text-vanna-teal hover:underline mb-1">
									{dir.project_name || 'Unknown Project'}
								</p>
								<p class="font-mono text-sm text-vanna-navy truncate" title={dir.path}>
									{dir.path}
								</p>
							</div>
							<svg class="w-5 h-5 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
							</svg>
						</div>
					</a>
				</Card>
			{/each}
		</div>
	{/if}
</div>
