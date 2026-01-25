<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';

	$: directoryId = $page.params.directoryId;

	let directory = null;
	let gitStatus = null;
	let loading = true;
	let error = null;
	let selectedFile = null;
	let diffContent = '';
	let diffLoading = false;
	let commitMessage = '';
	let committing = false;
	let pushing = false;
	let stagingFile = null;
	let unstagingFile = null;

	$: breadcrumbSegments = [
		{ label: "", href: "/", icon: "banner" },
		{ label: "Git Review", href: `/git/${directoryId}` }
	];

	onMount(async () => {
		await loadDirectoryAndStatus();
	});

	async function loadDirectoryAndStatus() {
		try {
			loading = true;
			error = null;

			// Fetch the directory info
			const dirResponse = await fetch(`/api/base-directories/${directoryId}`);
			if (!dirResponse.ok) {
				throw new Error('Failed to load directory');
			}
			directory = await dirResponse.json();

			// Fetch git status for this directory
			await refreshGitStatus();
		} catch (err) {
			console.error('Failed to load:', err);
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function refreshGitStatus() {
		if (!directory) return;

		try {
			const statusResponse = await fetch(`/api/git/status?path=${encodeURIComponent(directory.path)}`);
			if (statusResponse.ok) {
				gitStatus = await statusResponse.json();
			}
		} catch (err) {
			console.error('Failed to get git status:', err);
		}
	}

	async function viewDiff(file, staged = false, untracked = false) {
		selectedFile = { ...file, staged, untracked };
		diffLoading = true;
		diffContent = '';

		try {
			let url = `/api/git/diff?path=${encodeURIComponent(directory.path)}&file=${encodeURIComponent(file.path)}`;
			if (untracked) {
				url += '&untracked=true';
			} else {
				url += `&staged=${staged}`;
			}
			const response = await fetch(url);
			if (response.ok) {
				const data = await response.json();
				diffContent = data.diff || '(No changes to display)';
			}
		} catch (err) {
			console.error('Failed to get diff:', err);
			diffContent = 'Failed to load diff';
		} finally {
			diffLoading = false;
		}
	}

	async function stageFile(file) {
		stagingFile = file.path;
		try {
			const response = await fetch('/api/git/add', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					path: directory.path,
					file: file.path
				})
			});
			if (response.ok) {
				await refreshGitStatus();
				// If viewing this file's diff, refresh it
				if (selectedFile && selectedFile.path === file.path) {
					await viewDiff(file, true);
				}
			}
		} catch (err) {
			console.error('Failed to stage file:', err);
		} finally {
			stagingFile = null;
		}
	}

	async function unstageFile(file) {
		unstagingFile = file.path;
		try {
			const response = await fetch('/api/git/unstage', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					path: directory.path,
					file: file.path
				})
			});
			if (response.ok) {
				await refreshGitStatus();
				// If viewing this file's diff, refresh it
				if (selectedFile && selectedFile.path === file.path) {
					await viewDiff(file, false);
				}
			}
		} catch (err) {
			console.error('Failed to unstage file:', err);
		} finally {
			unstagingFile = null;
		}
	}

	async function stageAll() {
		stagingFile = '__all__';
		try {
			const response = await fetch('/api/git/add', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					path: directory.path,
					all: true
				})
			});
			if (response.ok) {
				await refreshGitStatus();
			}
		} catch (err) {
			console.error('Failed to stage all:', err);
		} finally {
			stagingFile = null;
		}
	}

	async function unstageAll() {
		unstagingFile = '__all__';
		try {
			// Unstage each file individually
			for (const file of gitStatus.stagedFiles || []) {
				await fetch('/api/git/unstage', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						path: directory.path,
						file: file.path
					})
				});
			}
			await refreshGitStatus();
		} catch (err) {
			console.error('Failed to unstage all:', err);
		} finally {
			unstagingFile = null;
		}
	}

	async function commit() {
		if (!commitMessage.trim()) return;

		committing = true;
		try {
			const response = await fetch('/api/git/commit', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					path: directory.path,
					message: commitMessage
				})
			});
			if (response.ok) {
				commitMessage = '';
				await refreshGitStatus();
				selectedFile = null;
				diffContent = '';
			} else {
				const data = await response.json();
				alert('Commit failed: ' + (data.error || 'Unknown error'));
			}
		} catch (err) {
			console.error('Failed to commit:', err);
			alert('Failed to commit');
		} finally {
			committing = false;
		}
	}

	async function push() {
		pushing = true;
		try {
			const response = await fetch('/api/git/push', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					path: directory.path
				})
			});
			if (response.ok) {
				await refreshGitStatus();
			} else {
				const data = await response.json();
				alert('Push failed: ' + (data.error || 'Unknown error'));
			}
		} catch (err) {
			console.error('Failed to push:', err);
			alert('Failed to push');
		} finally {
			pushing = false;
		}
	}

	function getFileStatusColor(file) {
		const x = file.x || '';
		const y = file.y || '';

		if (x === 'A' || y === 'A') return 'text-vanna-teal';
		if (x === 'D' || y === 'D') return 'text-white';
		if (x === 'M' || y === 'M') return 'text-vanna-magenta';
		if (x === 'R') return 'text-vanna-teal';
		if (x === '?' && y === '?') return 'text-slate-500';
		return 'text-vanna-navy';
	}

	function getFileStatusBgColor(file) {
		const x = file.x || '';
		const y = file.y || '';

		if (x === 'A' || y === 'A') return 'bg-vanna-teal/20';
		if (x === 'D' || y === 'D') return 'bg-vanna-orange';
		if (x === 'M' || y === 'M') return 'bg-vanna-magenta/20';
		if (x === 'R') return 'bg-vanna-teal/20';
		if (x === '?' && y === '?') return 'bg-slate-200';
		return 'bg-vanna-cream/50';
	}

	function getFileStatusLabel(file) {
		const x = file.x || '';
		const y = file.y || '';

		if (x === 'A' || y === 'A') return 'Added';
		if (x === 'D' || y === 'D') return 'Deleted';
		if (x === 'M' || y === 'M') return 'Modified';
		if (x === 'R') return 'Renamed';
		if (x === '?' && y === '?') return 'Untracked';
		return 'Changed';
	}

	$: hasUnstagedChanges = gitStatus && ((gitStatus.unstagedFiles?.length || 0) + (gitStatus.untrackedFiles?.length || 0)) > 0;
	$: hasStagedChanges = gitStatus && (gitStatus.stagedFiles?.length || 0) > 0;
	$: canPush = gitStatus && gitStatus.ahead > 0;
</script>

<svelte:head>
	<title>Git Review - {directory?.path || 'Loading...'}</title>
</svelte:head>

<div class="space-y-6">
	<Breadcrumb segments={breadcrumbSegments} />

	{#if loading}
		<div class="flex items-center justify-center min-h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div>
		</div>
	{:else if error}
		<Card class="border-vanna-orange/30 bg-vanna-orange/5">
			<div class="flex">
				<svg class="w-5 h-5 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-vanna-orange">Error</h3>
					<div class="mt-2 text-sm text-vanna-orange/80">
						<p>{error}</p>
					</div>
				</div>
			</div>
		</Card>
		{:else if directory && gitStatus}
			<!-- Header -->
			<div class="border-b border-slate-200 pb-6">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 bg-vanna-teal rounded-xl flex items-center justify-center">
							<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
							</svg>
						</div>
						<div>
							<h1 class="text-2xl font-bold text-vanna-navy font-serif">Git Review</h1>
							<p class="text-sm text-slate-500 font-mono">{directory.path}</p>
						</div>
					</div>

					<div class="flex items-center gap-3">
						<!-- Branch info -->
						<div class="flex items-center gap-2 px-3 py-2 bg-vanna-cream/50 rounded-lg">
							<svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"/>
							</svg>
							<span class="font-medium text-vanna-navy">{gitStatus.currentBranch || 'main'}</span>
							{#if gitStatus.ahead > 0}
								<span class="text-xs bg-vanna-teal/10 text-vanna-teal px-2 py-0.5 rounded-full">
									{gitStatus.ahead} ahead
								</span>
							{/if}
							{#if gitStatus.behind > 0}
								<span class="text-xs bg-vanna-orange/10 text-vanna-orange px-2 py-0.5 rounded-full">
									{gitStatus.behind} behind
								</span>
							{/if}
						</div>

						<Button
							variant="primary"
							onclick={push}
							disabled={pushing || !canPush}
							loading={pushing}
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
							</svg>
							Push
						</Button>
					</div>
				</div>
			</div>

			{#if !gitStatus.isRepo}
				<Card>
					<div class="text-center py-8">
						<svg class="w-12 h-12 mx-auto text-vanna-orange mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
						</svg>
						<h3 class="text-lg font-medium text-vanna-navy mb-2">Not a Git Repository</h3>
						<p class="text-slate-500">This directory is not initialized as a git repository.</p>
					</div>
				</Card>
			{:else}
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<!-- Left Column: File Changes -->
					<div class="space-y-6">
						<!-- Staged Changes -->
						<Card>
							<div class="flex items-center justify-between mb-4">
								<div class="flex items-center gap-2">
									<div class="w-3 h-3 rounded-full bg-vanna-teal"></div>
									<h3 class="font-semibold text-vanna-navy">Staged Changes</h3>
									<span class="text-sm text-slate-500">({gitStatus.stagedFiles?.length || 0})</span>
								</div>
								{#if hasStagedChanges}
									<Button
										variant="ghost"
										size="xs"
										onclick={unstageAll}
										disabled={unstagingFile === '__all__'}
									>
										Unstage All
									</Button>
								{/if}
							</div>

							{#if (gitStatus.stagedFiles?.length || 0) > 0}
								<div class="space-y-1 max-h-48 overflow-y-auto">
									{#each gitStatus.stagedFiles as file}
										<div
											class="flex items-center justify-between p-2 rounded-lg hover:bg-vanna-cream/30 cursor-pointer group {selectedFile?.path === file.path && selectedFile?.staged ? 'bg-vanna-teal/10' : ''}"
											on:click={() => viewDiff(file, true)}
										>
											<div class="flex items-center gap-2 min-w-0">
												<span class="text-xs font-mono px-1.5 py-0.5 rounded {getFileStatusColor(file)} {getFileStatusBgColor(file)}">
													{getFileStatusLabel(file)}
												</span>
												<span class="text-sm text-vanna-navy truncate">{file.path}</span>
											</div>
											<Button
												variant="ghost"
												size="xs"
												onclick={(e) => { e.stopPropagation(); unstageFile(file); }}
												disabled={unstagingFile === file.path}
												class="opacity-0 group-hover:opacity-100"
											>
												{unstagingFile === file.path ? '...' : 'Unstage'}
											</Button>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-slate-500 text-center py-4">No staged changes</p>
							{/if}
						</Card>

						<!-- Unstaged Changes -->
						<Card>
							<div class="flex items-center justify-between mb-4">
								<div class="flex items-center gap-2">
									<div class="w-3 h-3 rounded-full bg-vanna-orange"></div>
									<h3 class="font-semibold text-vanna-navy">Unstaged Changes</h3>
									<span class="text-sm text-slate-500">({gitStatus.unstagedFiles?.length || 0})</span>
								</div>
								{#if (gitStatus.unstagedFiles?.length || 0) > 0}
									<Button
										variant="ghost"
										size="xs"
										onclick={stageAll}
										disabled={stagingFile === '__all__'}
									>
										Stage All
									</Button>
								{/if}
							</div>

							{#if (gitStatus.unstagedFiles?.length || 0) > 0}
								<div class="space-y-1 max-h-48 overflow-y-auto">
									{#each gitStatus.unstagedFiles as file}
										<div
											class="flex items-center justify-between p-2 rounded-lg hover:bg-vanna-cream/30 cursor-pointer group {selectedFile?.path === file.path && !selectedFile?.staged ? 'bg-vanna-orange/10' : ''}"
											on:click={() => viewDiff(file, false)}
										>
											<div class="flex items-center gap-2 min-w-0">
												<span class="text-xs font-mono px-1.5 py-0.5 rounded {getFileStatusColor(file)} {getFileStatusBgColor(file)}">
													{getFileStatusLabel(file)}
												</span>
												<span class="text-sm text-vanna-navy truncate">{file.path}</span>
											</div>
											<Button
												variant="ghost"
												size="xs"
												onclick={(e) => { e.stopPropagation(); stageFile(file); }}
												disabled={stagingFile === file.path}
												class="opacity-0 group-hover:opacity-100"
											>
												{stagingFile === file.path ? '...' : 'Stage'}
											</Button>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-slate-500 text-center py-4">No unstaged changes</p>
							{/if}
						</Card>

						<!-- Untracked Files -->
						<Card>
							<div class="flex items-center justify-between mb-4">
								<div class="flex items-center gap-2">
									<div class="w-3 h-3 rounded-full bg-slate-400"></div>
									<h3 class="font-semibold text-vanna-navy">Untracked Files</h3>
									<span class="text-sm text-slate-500">({gitStatus.untrackedFiles?.length || 0})</span>
								</div>
							</div>

							{#if (gitStatus.untrackedFiles?.length || 0) > 0}
								<div class="space-y-1 max-h-48 overflow-y-auto">
									{#each gitStatus.untrackedFiles as file}
										<div
											class="flex items-center justify-between p-2 rounded-lg hover:bg-vanna-cream/30 cursor-pointer group {selectedFile?.path === file.path && selectedFile?.untracked ? 'bg-slate-200/50' : ''}"
											on:click={() => viewDiff(file, false, true)}
										>
											<div class="flex items-center gap-2 min-w-0">
												<span class="text-xs font-mono px-1.5 py-0.5 rounded text-slate-500 bg-vanna-cream/50">
													New
												</span>
												<span class="text-sm text-vanna-navy truncate">{file.path}</span>
											</div>
											<Button
												variant="ghost"
												size="xs"
												onclick={(e) => { e.stopPropagation(); stageFile(file); }}
												disabled={stagingFile === file.path}
												class="opacity-0 group-hover:opacity-100"
											>
												{stagingFile === file.path ? '...' : 'Stage'}
											</Button>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-slate-500 text-center py-4">No untracked files</p>
							{/if}
						</Card>

						<!-- Merge Conflicts -->
						{#if (gitStatus.mergeConflicts?.length || 0) > 0}
							<Card class="border-vanna-magenta/30">
								<div class="flex items-center gap-2 mb-4">
									<div class="w-3 h-3 rounded-full bg-vanna-magenta"></div>
									<h3 class="font-semibold text-vanna-magenta">Merge Conflicts</h3>
									<span class="text-sm text-vanna-magenta/80">({gitStatus.mergeConflicts.length})</span>
								</div>

								<div class="space-y-1">
									{#each gitStatus.mergeConflicts as file}
										<div class="flex items-center gap-2 p-2 bg-vanna-magenta/10 rounded-lg">
											<svg class="w-4 h-4 text-vanna-magenta" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
											</svg>
											<span class="text-sm text-vanna-magenta">{file.path}</span>
										</div>
									{/each}
								</div>
							</Card>
						{/if}

						<!-- Commit Section -->
						<Card>
							<h3 class="font-semibold text-vanna-navy mb-4">Create Commit</h3>

							<div class="space-y-4">
								<Textarea
									bind:value={commitMessage}
									placeholder="Enter commit message..."
									rows={3}
								/>

								<Button
									variant="success"
									onclick={commit}
									disabled={committing || !hasStagedChanges || !commitMessage.trim()}
									loading={committing}
									class="w-full"
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
									</svg>
									{committing ? 'Committing...' : 'Commit'}
								</Button>
							</div>
						</Card>
					</div>

					<!-- Right Column: Diff Viewer -->
					<div class="lg:sticky lg:top-6 lg:self-start">
						<Card class="h-[calc(100vh-12rem)] flex flex-col">
							<div class="flex items-center justify-between mb-4">
								<h3 class="font-semibold text-vanna-navy">
									{#if selectedFile}
										{selectedFile.untracked ? 'New File' : 'Diff'}: {selectedFile.path}
										<span class="text-xs font-normal text-slate-500 ml-2">
											({selectedFile.untracked ? 'untracked' : selectedFile.staged ? 'staged' : 'unstaged'})
										</span>
									{:else}
										Diff Viewer
									{/if}
								</h3>
								{#if selectedFile}
									<Button
										variant="ghost"
										size="xs"
										onclick={() => { selectedFile = null; diffContent = ''; }}
									>
										Clear
									</Button>
								{/if}
							</div>

							<div class="flex-1 overflow-auto bg-vanna-navy rounded-lg p-4 font-mono text-sm">
								{#if diffLoading}
									<div class="flex items-center justify-center h-full">
										<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-vanna-teal"></div>
									</div>
								{:else if diffContent}
									<pre class="whitespace-pre-wrap">{#each diffContent.split('\n') as line}{#if line.startsWith('+') && !line.startsWith('+++')}<span class="text-green-400">{line}</span>
{:else if line.startsWith('-') && !line.startsWith('---')}<span class="text-red-400">{line}</span>
{:else if line.startsWith('@@')}<span class="text-vanna-teal">{line}</span>
{:else if line.startsWith('diff') || line.startsWith('index') || line.startsWith('---') || line.startsWith('+++')}<span class="text-slate-500">{line}</span>
{:else}<span class="text-slate-300">{line}</span>
{/if}{/each}</pre>
								{:else}
									<div class="flex flex-col items-center justify-center h-full text-slate-400">
										<svg class="w-12 h-12 mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
										</svg>
										<p>Select a file to view diff</p>
									</div>
								{/if}
							</div>
						</Card>
					</div>
				</div>
			{/if}
		{/if}
</div>
