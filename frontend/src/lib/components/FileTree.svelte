<script lang="ts">
	interface FileEntry {
		name: string;
		isDir: boolean;
		size: number;
		modTime: string;
	}

	interface Props {
		basePath: string;
		selectedPath?: string;
		onSelectFile?: (path: string) => void;
		onFileCreated?: (path: string) => void;
	}

	let { basePath, selectedPath = '', onSelectFile, onFileCreated }: Props = $props();

	let entries = $state<FileEntry[]>([]);
	let expandedDirs = $state<Set<string>>(new Set());
	let loadedDirs = $state<Map<string, FileEntry[]>>(new Map());
	let loading = $state(true);
	let error = $state('');
	let showNewFileInput = $state(false);
	let newFileName = $state('');
	let creatingFile = $state(false);
	let newFileError = $state('');

	$effect(() => {
		loadDirectory(basePath);
	});

	async function loadDirectory(path: string): Promise<FileEntry[]> {
		try {
			const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`);
			if (!res.ok) {
				const data = await res.json();
				throw new Error(data.error || 'Failed to load directory');
			}
			const data = await res.json();
			const sortedEntries = sortEntries(data.entries || []);

			if (path === basePath) {
				entries = sortedEntries;
				loading = false;
			}
			loadedDirs.set(path, sortedEntries);
			loadedDirs = new Map(loadedDirs);
			return sortedEntries;
		} catch (err) {
			if (path === basePath) {
				error = err instanceof Error ? err.message : 'Failed to load directory';
				loading = false;
			}
			return [];
		}
	}

	function sortEntries(entries: FileEntry[]): FileEntry[] {
		return [...entries].sort((a, b) => {
			// Directories first
			if (a.isDir && !b.isDir) return -1;
			if (!a.isDir && b.isDir) return 1;
			// Then alphabetically
			return a.name.localeCompare(b.name);
		});
	}

	async function toggleDir(path: string) {
		if (expandedDirs.has(path)) {
			expandedDirs.delete(path);
			expandedDirs = new Set(expandedDirs);
		} else {
			expandedDirs.add(path);
			expandedDirs = new Set(expandedDirs);
			// Load if not already loaded
			if (!loadedDirs.has(path)) {
				await loadDirectory(path);
			}
		}
	}

	function selectFile(path: string) {
		onSelectFile?.(path);
	}

	async function createFile() {
		if (!newFileName.trim()) return;

		creatingFile = true;
		newFileError = '';

		try {
			const filePath = `${basePath}/${newFileName.trim()}`;
			const res = await fetch('/api/files/content', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ path: filePath, content: '' })
			});

			if (!res.ok) {
				const data = await res.json();
				throw new Error(data.error || 'Failed to create file');
			}

			// Refresh the directory listing
			await loadDirectory(basePath);

			// Reset state and select the new file
			showNewFileInput = false;
			newFileName = '';
			onFileCreated?.(filePath);
			onSelectFile?.(filePath);
		} catch (err) {
			newFileError = err instanceof Error ? err.message : 'Failed to create file';
		} finally {
			creatingFile = false;
		}
	}

	function cancelNewFile() {
		showNewFileInput = false;
		newFileName = '';
		newFileError = '';
	}

	function handleNewFileKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			createFile();
		} else if (e.key === 'Escape') {
			cancelNewFile();
		}
	}

	function getFileIcon(name: string, isDir: boolean): string {
		if (isDir) return 'folder';

		const ext = name.split('.').pop()?.toLowerCase() || '';
		const iconMap: Record<string, string> = {
			'js': 'javascript',
			'ts': 'typescript',
			'jsx': 'react',
			'tsx': 'react',
			'svelte': 'svelte',
			'vue': 'vue',
			'py': 'python',
			'go': 'go',
			'rs': 'rust',
			'json': 'json',
			'md': 'markdown',
			'css': 'css',
			'scss': 'css',
			'html': 'html',
			'sql': 'database',
			'yml': 'yaml',
			'yaml': 'yaml',
			'sh': 'terminal',
			'bash': 'terminal',
		};
		return iconMap[ext] || 'file';
	}
</script>

{#snippet TreeNode(parentPath: string, items: FileEntry[], depth: number)}
	{#each items as entry}
		{@const fullPath = `${parentPath}/${entry.name}`}
		{@const isExpanded = expandedDirs.has(fullPath)}
		{@const isSelected = selectedPath === fullPath}
		{@const children = loadedDirs.get(fullPath) || []}

		<div class="select-none">
			<button
				class="w-full flex items-center gap-1.5 px-2 py-1 text-sm text-left hover:bg-slate-100 rounded transition-colors {isSelected ? 'bg-vanna-teal/10 text-vanna-teal' : 'text-slate-700'}"
				style="padding-left: {depth * 16 + 8}px"
				onclick={() => entry.isDir ? toggleDir(fullPath) : selectFile(fullPath)}
			>
				<!-- Chevron for directories -->
				{#if entry.isDir}
					<svg
						class="w-4 h-4 text-slate-400 transition-transform {isExpanded ? 'rotate-90' : ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				{:else}
					<span class="w-4"></span>
				{/if}

				<!-- Icon -->
				{#if entry.isDir}
					<svg class="w-4 h-4 {isExpanded ? 'text-vanna-teal' : 'text-amber-500'}" fill="currentColor" viewBox="0 0 24 24">
						{#if isExpanded}
							<path d="M2 6a2 2 0 012-2h5l2 2h9a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
						{:else}
							<path d="M2 6a2 2 0 012-2h5l2 2h9a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
						{/if}
					</svg>
				{:else}
					<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
					</svg>
				{/if}

				<!-- Name -->
				<span class="truncate">{entry.name}</span>
			</button>

			<!-- Children -->
			{#if entry.isDir && isExpanded}
				{@render TreeNode(fullPath, children, depth + 1)}
			{/if}
		</div>
	{/each}
{/snippet}

<div class="h-full flex flex-col bg-white border-r border-slate-200">
	<!-- Header with New File button -->
	<div class="flex items-center justify-between px-3 py-2 border-b border-slate-200 bg-slate-50">
		<span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Files</span>
		<button
			class="p-1 text-slate-500 hover:text-vanna-teal hover:bg-vanna-teal/10 rounded transition-colors"
			onclick={() => { showNewFileInput = true; }}
			title="New File"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
			</svg>
		</button>
	</div>

	<!-- New File Input -->
	{#if showNewFileInput}
		<div class="px-3 py-2 border-b border-slate-200 bg-vanna-cream/30">
			<div class="flex items-center gap-2">
				<input
					type="text"
					bind:value={newFileName}
					onkeydown={handleNewFileKeydown}
					placeholder="filename.txt"
					class="flex-1 px-2 py-1 text-sm border border-slate-300 rounded focus:outline-none focus:ring-1 focus:ring-vanna-teal focus:border-vanna-teal"
					autofocus
				/>
				<button
					onclick={createFile}
					disabled={creatingFile || !newFileName.trim()}
					class="p-1 text-vanna-teal hover:bg-vanna-teal/10 rounded disabled:opacity-50 disabled:cursor-not-allowed"
					title="Create"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
					</svg>
				</button>
				<button
					onclick={cancelNewFile}
					class="p-1 text-slate-500 hover:bg-slate-100 rounded"
					title="Cancel"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>
			{#if newFileError}
				<p class="mt-1 text-xs text-red-600">{newFileError}</p>
			{/if}
		</div>
	{/if}

	<!-- Tree Content -->
	<div class="flex-1 overflow-auto">
		{#if loading}
			<div class="flex items-center justify-center h-32">
				<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-vanna-teal"></div>
			</div>
		{:else if error}
			<div class="p-4 text-sm text-red-600">
				{error}
			</div>
		{:else}
			<div class="py-2">
				{@render TreeNode(basePath.replace(/\/$/, ''), entries, 0)}
			</div>
		{/if}
	</div>
</div>
