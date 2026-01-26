<script>
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	let directories = [];
	let gitStatusMap = new Map();
	let loading = true;
	let showDirtyOnly = false;
	let groupByProject = false;

	onMount(async () => {
		await loadDirectories();
		await loadAllGitStatus();
		// Refresh every 5 seconds
		const interval = setInterval(loadAllGitStatus, 5000);
		return () => clearInterval(interval);
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

	async function loadAllGitStatus() {
		for (const dir of directories) {
			try {
				const res = await fetch(`/api/git/status?path=${encodeURIComponent(dir.path)}`);
				if (res.ok) {
					gitStatusMap.set(dir.id, await res.json());
				}
			} catch (err) {
				console.error(`Failed to get git status for ${dir.path}:`, err);
			}
		}
		gitStatusMap = new Map(gitStatusMap); // trigger reactivity
	}

	function getChangeCount(status) {
		if (!status) return 0;
		return (status.stagedFiles?.length || 0) +
			(status.unstagedFiles?.length || 0) +
			(status.untrackedFiles?.length || 0);
	}

	function isDirty(status) {
		return getChangeCount(status) > 0;
	}

	$: filteredDirectories = showDirtyOnly
		? directories.filter(dir => isDirty(gitStatusMap.get(dir.id)))
		: directories;

	$: groupedDirectories = groupByProject
		? filteredDirectories.reduce((acc, dir) => {
			const project = dir.project_name || 'Unknown Project';
			if (!acc[project]) acc[project] = [];
			acc[project].push(dir);
			return acc;
		}, {})
		: null;

	$: dirtyCount = directories.filter(dir => isDirty(gitStatusMap.get(dir.id))).length;
</script>

<svelte:head>
	<title>Directories - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold text-vanna-navy font-serif">Directories</h1>
		<p class="mt-2 text-slate-500">View git status across all projects</p>
	</div>

	<!-- Filters -->
	<div class="flex items-center gap-4">
		<label class="flex items-center gap-2 cursor-pointer">
			<input
				type="checkbox"
				bind:checked={showDirtyOnly}
				class="rounded border-slate-300 text-vanna-teal focus:ring-vanna-teal"
			/>
			<span class="text-sm text-slate-600">Show only dirty ({dirtyCount})</span>
		</label>
		<label class="flex items-center gap-2 cursor-pointer">
			<input
				type="checkbox"
				bind:checked={groupByProject}
				class="rounded border-slate-300 text-vanna-teal focus:ring-vanna-teal"
			/>
			<span class="text-sm text-slate-600">Group by project</span>
		</label>
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
	{:else if groupByProject}
		<!-- Grouped view -->
		{#each Object.entries(groupedDirectories) as [projectName, dirs]}
			<div class="space-y-3">
				<h2 class="text-lg font-semibold text-vanna-navy flex items-center gap-2">
					<svg class="w-5 h-5 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
					</svg>
					{projectName}
				</h2>
				<div class="grid gap-4">
					{#each dirs as dir}
						{@const status = gitStatusMap.get(dir.id)}
						{@const changeCount = getChangeCount(status)}
						<Card class="hover:shadow-md transition-shadow">
							<div class="flex items-center justify-between">
								<div class="min-w-0 flex-1">
									<p class="font-mono text-sm text-vanna-navy truncate">{dir.path}</p>
									<div class="flex items-center gap-3 mt-2">
										{#if status}
											<!-- Branch -->
											<div class="flex items-center gap-1.5 text-sm">
												<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"/>
												</svg>
												<span class="text-slate-600">{status.currentBranch || 'main'}</span>
											</div>
											<!-- Ahead/Behind -->
											{#if status.ahead > 0}
												<span class="text-xs bg-vanna-teal/10 text-vanna-teal px-2 py-0.5 rounded-full">
													{status.ahead} ahead
												</span>
											{/if}
											{#if status.behind > 0}
												<span class="text-xs bg-vanna-orange/10 text-vanna-orange px-2 py-0.5 rounded-full">
													{status.behind} behind
												</span>
											{/if}
											<!-- Change count -->
											{#if changeCount > 0}
												<span class="text-xs bg-vanna-magenta/10 text-vanna-magenta px-2 py-0.5 rounded-full">
													{changeCount} change{changeCount !== 1 ? 's' : ''}
												</span>
											{:else}
												<span class="text-xs bg-green-100 text-green-600 px-2 py-0.5 rounded-full">
													Clean
												</span>
											{/if}
										{:else}
											<span class="text-xs text-slate-400">Loading...</span>
										{/if}
									</div>
								</div>
								<Button href="/git/{dir.id}" variant="ghost" size="sm">
									View Changes
									<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
									</svg>
								</Button>
							</div>
						</Card>
					{/each}
				</div>
			</div>
		{/each}
	{:else}
		<!-- Flat view -->
		<div class="grid gap-4">
			{#each filteredDirectories as dir}
				{@const status = gitStatusMap.get(dir.id)}
				{@const changeCount = getChangeCount(status)}
				<Card class="hover:shadow-md transition-shadow">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2 mb-1">
								<a href="/projects/{dir.project_id}" class="text-xs text-vanna-teal hover:underline">
									{dir.project_name || 'Unknown Project'}
								</a>
							</div>
							<p class="font-mono text-sm text-vanna-navy truncate">{dir.path}</p>
							<div class="flex items-center gap-3 mt-2">
								{#if status}
									<!-- Branch -->
									<div class="flex items-center gap-1.5 text-sm">
										<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"/>
										</svg>
										<span class="text-slate-600">{status.currentBranch || 'main'}</span>
									</div>
									<!-- Ahead/Behind -->
									{#if status.ahead > 0}
										<span class="text-xs bg-vanna-teal/10 text-vanna-teal px-2 py-0.5 rounded-full">
											{status.ahead} ahead
										</span>
									{/if}
									{#if status.behind > 0}
										<span class="text-xs bg-vanna-orange/10 text-vanna-orange px-2 py-0.5 rounded-full">
											{status.behind} behind
										</span>
									{/if}
									<!-- Change count -->
									{#if changeCount > 0}
										<span class="text-xs bg-vanna-magenta/10 text-vanna-magenta px-2 py-0.5 rounded-full">
											{changeCount} change{changeCount !== 1 ? 's' : ''}
										</span>
									{:else}
										<span class="text-xs bg-green-100 text-green-600 px-2 py-0.5 rounded-full">
											Clean
										</span>
									{/if}
								{:else}
									<span class="text-xs text-slate-400">Loading...</span>
								{/if}
							</div>
						</div>
						<Button href="/git/{dir.id}" variant="ghost" size="sm">
							View Changes
							<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
							</svg>
						</Button>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
