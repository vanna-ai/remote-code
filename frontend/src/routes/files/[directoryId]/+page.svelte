<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import FileTree from '$lib/components/FileTree.svelte';
	import FileEditor from '$lib/components/FileEditor.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	$: directoryId = $page.params.directoryId;

	let directory: any = null;
	let loading = true;
	let error = '';
	let selectedFilePath = '';
	let treeWidth = 300;
	let resizing = false;
	let devServerStatus: any = null;
	let startingDevServer = false;
	let stoppingDevServer = false;

	$: breadcrumbSegments = [
		{ label: '', href: '/', icon: 'banner' },
		{ label: 'Directories', href: '/directories' },
		{ label: 'Files', href: `/files/${directoryId}` }
	];

	onMount(async () => {
		await loadDirectory();
		await loadDevServerStatus();
		// Poll for dev server status
		const interval = setInterval(loadDevServerStatus, 5000);
		return () => clearInterval(interval);
	});

	async function loadDirectory() {
		try {
			loading = true;
			error = '';

			const res = await fetch(`/api/base-directories/${directoryId}`);
			if (!res.ok) {
				throw new Error('Failed to load directory');
			}
			directory = await res.json();
		} catch (err) {
			console.error('Failed to load:', err);
			error = err instanceof Error ? err.message : 'Failed to load directory';
		} finally {
			loading = false;
		}
	}

	async function loadDevServerStatus() {
		try {
			const res = await fetch(`/api/base-directories/${directoryId}/dev-server`);
			if (res.ok) {
				devServerStatus = await res.json();
			}
		} catch (err) {
			console.error('Failed to load dev server status:', err);
		}
	}

	async function startDevServer() {
		startingDevServer = true;
		try {
			const res = await fetch(`/api/base-directories/${directoryId}/dev-server`, {
				method: 'POST'
			});
			if (res.ok) {
				devServerStatus = await res.json();
			} else {
				const error = await res.text();
				alert('Failed to start dev server: ' + error);
			}
		} catch (err) {
			console.error('Failed to start dev server:', err);
			alert('Failed to start dev server');
		} finally {
			startingDevServer = false;
		}
	}

	async function stopDevServer() {
		stoppingDevServer = true;
		try {
			const res = await fetch(`/api/base-directories/${directoryId}/dev-server`, {
				method: 'DELETE'
			});
			if (res.ok) {
				devServerStatus = { running: false };
			} else {
				alert('Failed to stop dev server');
			}
		} catch (err) {
			console.error('Failed to stop dev server:', err);
			alert('Failed to stop dev server');
		} finally {
			stoppingDevServer = false;
		}
	}

	function hasDevServerCommands() {
		return directory?.dev_server_setup_commands && directory.dev_server_setup_commands.trim() !== '';
	}

	function isDevServerRunning() {
		return devServerStatus?.running === true;
	}

	function selectFile(path: string) {
		selectedFilePath = path;
	}

	function startResize(e: MouseEvent) {
		resizing = true;
		document.addEventListener('mousemove', handleResize);
		document.addEventListener('mouseup', stopResize);
	}

	function handleResize(e: MouseEvent) {
		if (!resizing) return;
		const newWidth = e.clientX;
		treeWidth = Math.max(200, Math.min(600, newWidth));
	}

	function stopResize() {
		resizing = false;
		document.removeEventListener('mousemove', handleResize);
		document.removeEventListener('mouseup', stopResize);
	}
</script>

<svelte:head>
	<title>Files - {directory?.path || 'Loading...'}</title>
</svelte:head>

<div class="space-y-4">
	<Breadcrumb segments={breadcrumbSegments} />

	{#if loading}
		<div class="flex items-center justify-center min-h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div>
		</div>
	{:else if error}
		<div class="p-6 border border-vanna-orange/30 bg-vanna-orange/5 rounded-xl">
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
		</div>
	{:else if directory}
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-slate-200 pb-4">
			<div class="flex items-center gap-4">
				<div class="w-10 h-10 bg-vanna-teal rounded-xl flex items-center justify-center">
					<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
					</svg>
				</div>
				<div>
					<div class="flex items-center gap-2">
						<h1 class="text-xl font-bold text-vanna-navy font-serif">File Browser</h1>
						{#if isDevServerRunning()}
							<span class="text-xs bg-green-100 text-green-600 px-2 py-0.5 rounded-full flex items-center gap-1">
								<span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
								Dev Server Running
							</span>
						{/if}
					</div>
					<p class="text-sm text-slate-500 font-mono">{directory.path}</p>
				</div>
			</div>

			<div class="flex items-center gap-2">
				<!-- Dev Server Controls -->
				{#if hasDevServerCommands()}
					{#if isDevServerRunning()}
						<Button
							variant="ghost"
							size="sm"
							onclick={stopDevServer}
							disabled={stoppingDevServer}
						>
							{#if stoppingDevServer}
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
							Stop Dev
						</Button>
						<Button href="/dev-server/{directoryId}" variant="ghost" size="sm">
							<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
							</svg>
							Terminal
						</Button>
					{:else}
						<Button
							variant="ghost"
							size="sm"
							onclick={startDevServer}
							disabled={startingDevServer}
						>
							{#if startingDevServer}
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
				<Button href="/git/{directoryId}" variant="ghost" size="sm">
					<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
					</svg>
					Git Review
				</Button>
			</div>
		</div>

		<!-- Main Content -->
		<div class="flex h-[calc(100vh-14rem)] bg-white rounded-xl border border-slate-200 overflow-hidden shadow-vanna-card">
			<!-- File Tree -->
			<div style="width: {treeWidth}px" class="flex-shrink-0 border-r border-slate-200">
				<FileTree
					basePath={directory.path}
					selectedPath={selectedFilePath}
					onSelectFile={selectFile}
				/>
			</div>

			<!-- Resize Handle -->
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div
				class="w-1 bg-slate-200 hover:bg-vanna-teal cursor-col-resize flex-shrink-0 transition-colors"
				onmousedown={startResize}
				role="separator"
				aria-orientation="vertical"
				tabindex="0"
				aria-label="Resize panel"
			></div>

			<!-- Editor -->
			<div class="flex-1 min-w-0">
				{#if selectedFilePath}
					<FileEditor filePath={selectedFilePath} />
				{:else}
					<div class="h-full flex flex-col items-center justify-center text-slate-400 bg-slate-50">
						<svg class="w-16 h-16 mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
						</svg>
						<p class="text-lg font-medium mb-1">No file selected</p>
						<p class="text-sm">Select a file from the tree to start editing</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
