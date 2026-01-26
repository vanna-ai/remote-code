<script>
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	let directories = [];
	let gitStatusMap = new Map();
	let devServerStatusMap = new Map();
	let loading = true;
	let showDirtyOnly = false;
	let groupByProject = false;
	let startingDevServer = new Set();
	let stoppingDevServer = new Set();

	onMount(async () => {
		await loadDirectories();
		await Promise.all([loadAllGitStatus(), loadAllDevServerStatus()]);
		// Refresh every 5 seconds
		const interval = setInterval(() => {
			loadAllGitStatus();
			loadAllDevServerStatus();
		}, 5000);
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

	async function loadAllDevServerStatus() {
		for (const dir of directories) {
			try {
				const res = await fetch(`/api/base-directories/${dir.id}/dev-server`);
				if (res.ok) {
					devServerStatusMap.set(dir.id, await res.json());
				}
			} catch (err) {
				console.error(`Failed to get dev server status for ${dir.path}:`, err);
			}
		}
		devServerStatusMap = new Map(devServerStatusMap); // trigger reactivity
	}

	async function startDevServer(dirId) {
		startingDevServer.add(dirId);
		startingDevServer = new Set(startingDevServer);
		try {
			const res = await fetch(`/api/base-directories/${dirId}/dev-server`, {
				method: 'POST'
			});
			if (res.ok) {
				const status = await res.json();
				devServerStatusMap.set(dirId, status);
				devServerStatusMap = new Map(devServerStatusMap);
			} else {
				const error = await res.text();
				alert('Failed to start dev server: ' + error);
			}
		} catch (err) {
			console.error('Failed to start dev server:', err);
			alert('Failed to start dev server');
		} finally {
			startingDevServer.delete(dirId);
			startingDevServer = new Set(startingDevServer);
		}
	}

	async function stopDevServer(dirId) {
		stoppingDevServer.add(dirId);
		stoppingDevServer = new Set(stoppingDevServer);
		try {
			const res = await fetch(`/api/base-directories/${dirId}/dev-server`, {
				method: 'DELETE'
			});
			if (res.ok) {
				devServerStatusMap.set(dirId, { running: false });
				devServerStatusMap = new Map(devServerStatusMap);
			} else {
				alert('Failed to stop dev server');
			}
		} catch (err) {
			console.error('Failed to stop dev server:', err);
			alert('Failed to stop dev server');
		} finally {
			stoppingDevServer.delete(dirId);
			stoppingDevServer = new Set(stoppingDevServer);
		}
	}

	function hasDevServerCommands(dir) {
		return dir.dev_server_setup_commands && dir.dev_server_setup_commands.trim() !== '';
	}

	function isDevServerRunning(dirId) {
		const status = devServerStatusMap.get(dirId);
		return status?.running === true;
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
						{@const devServerRunning = isDevServerRunning(dir.id)}
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
										<!-- Dev Server Status -->
										{#if devServerRunning}
											<span class="text-xs bg-green-100 text-green-600 px-2 py-0.5 rounded-full flex items-center gap-1">
												<span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
												Dev Server
											</span>
										{/if}
									</div>
								</div>
								<div class="flex items-center gap-2">
									<!-- Dev Server Button -->
									{#if hasDevServerCommands(dir)}
										{#if devServerRunning}
											<Button
												variant="ghost"
												size="sm"
												onclick={() => stopDevServer(dir.id)}
												disabled={stoppingDevServer.has(dir.id)}
											>
												{#if stoppingDevServer.has(dir.id)}
													<svg class="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
														<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
														<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
													</svg>
												{:else}
													<svg class="w-4 h-4 mr-1 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
													</svg>
												{/if}
												Stop
											</Button>
											<Button href="/dev-server/{dir.id}" variant="ghost" size="sm">
												<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
												</svg>
												Terminal
											</Button>
										{:else}
											<Button
												variant="ghost"
												size="sm"
												onclick={() => startDevServer(dir.id)}
												disabled={startingDevServer.has(dir.id)}
											>
												{#if startingDevServer.has(dir.id)}
													<svg class="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
														<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
														<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
													</svg>
													Starting...
												{:else}
													<svg class="w-4 h-4 mr-1 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
													</svg>
													Start Dev
												{/if}
											</Button>
										{/if}
									{/if}
									<Button href="/files/{dir.id}" variant="ghost" size="sm">
										<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
										</svg>
										Files
									</Button>
									<Button href="/git/{dir.id}" variant="ghost" size="sm">
										Git
										<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
										</svg>
									</Button>
								</div>
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
				{@const devServerRunning = isDevServerRunning(dir.id)}
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
								<!-- Dev Server Status -->
								{#if devServerRunning}
									<span class="text-xs bg-green-100 text-green-600 px-2 py-0.5 rounded-full flex items-center gap-1">
										<span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
										Dev Server
									</span>
								{/if}
							</div>
						</div>
						<div class="flex items-center gap-2">
							<!-- Dev Server Button -->
							{#if hasDevServerCommands(dir)}
								{#if devServerRunning}
									<Button
										variant="ghost"
										size="sm"
										onclick={() => stopDevServer(dir.id)}
										disabled={stoppingDevServer.has(dir.id)}
									>
										{#if stoppingDevServer.has(dir.id)}
											<svg class="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
										{:else}
											<svg class="w-4 h-4 mr-1 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
											</svg>
										{/if}
										Stop
									</Button>
									<Button href="/dev-server/{dir.id}" variant="ghost" size="sm">
										<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
										</svg>
										Terminal
									</Button>
								{:else}
									<Button
										variant="ghost"
										size="sm"
										onclick={() => startDevServer(dir.id)}
										disabled={startingDevServer.has(dir.id)}
									>
										{#if startingDevServer.has(dir.id)}
											<svg class="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											Starting...
										{:else}
											<svg class="w-4 h-4 mr-1 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
											</svg>
											Start Dev
										{/if}
									</Button>
								{/if}
							{/if}
							<Button href="/files/{dir.id}" variant="ghost" size="sm">
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
								</svg>
								Files
							</Button>
							<Button href="/git/{dir.id}" variant="ghost" size="sm">
								Git
								<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
								</svg>
							</Button>
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
