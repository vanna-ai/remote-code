<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';

	interface Props {
		filePath: string;
		onSave?: (path: string, content: string) => Promise<void>;
	}

	let { filePath, onSave }: Props = $props();

	let container: HTMLDivElement;
	let editor: any = null;
	let monaco: any = null;
	let content = $state('');
	let originalContent = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let hasChanges = $derived(content !== originalContent);

	// Language detection by extension
	function getLanguage(path: string): string {
		const ext = path.split('.').pop()?.toLowerCase() || '';
		const langMap: Record<string, string> = {
			'js': 'javascript',
			'mjs': 'javascript',
			'cjs': 'javascript',
			'jsx': 'javascript',
			'ts': 'typescript',
			'tsx': 'typescript',
			'json': 'json',
			'html': 'html',
			'htm': 'html',
			'css': 'css',
			'scss': 'scss',
			'less': 'less',
			'md': 'markdown',
			'markdown': 'markdown',
			'py': 'python',
			'go': 'go',
			'rs': 'rust',
			'java': 'java',
			'c': 'c',
			'cpp': 'cpp',
			'h': 'c',
			'hpp': 'cpp',
			'sql': 'sql',
			'sh': 'shell',
			'bash': 'shell',
			'zsh': 'shell',
			'yaml': 'yaml',
			'yml': 'yaml',
			'xml': 'xml',
			'svg': 'xml',
			'svelte': 'html',
			'vue': 'html',
			'php': 'php',
			'rb': 'ruby',
			'swift': 'swift',
			'kt': 'kotlin',
			'graphql': 'graphql',
			'dockerfile': 'dockerfile',
		};
		return langMap[ext] || 'plaintext';
	}

	async function loadFile() {
		if (!filePath) return;

		loading = true;
		error = '';

		try {
			const res = await fetch(`/api/files/content?path=${encodeURIComponent(filePath)}`);
			const data = await res.json();

			if (!res.ok) {
				throw new Error(data.error || 'Failed to load file');
			}

			content = data.content;
			originalContent = data.content;

			if (editor) {
				editor.setValue(content);
				const model = editor.getModel();
				if (model) {
					monaco.editor.setModelLanguage(model, getLanguage(filePath));
				}
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load file';
		} finally {
			loading = false;
		}
	}

	async function saveFile() {
		if (!filePath || saving || !hasChanges) return;

		saving = true;
		error = '';

		try {
			if (onSave) {
				await onSave(filePath, content);
			} else {
				const res = await fetch('/api/files/content', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ path: filePath, content })
				});
				const data = await res.json();

				if (!res.ok) {
					throw new Error(data.error || 'Failed to save file');
				}
			}

			originalContent = content;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save file';
		} finally {
			saving = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 's') {
			e.preventDefault();
			saveFile();
		}
	}

	onMount(async () => {
		if (!browser) return;

		// Dynamic import of Monaco
		const monacoModule = await import('monaco-editor');
		monaco = monacoModule;

		// Configure Monaco environment - disable workers for simpler setup
		self.MonacoEnvironment = {
			getWorker: function () {
				return null as any;
			}
		};

		editor = monaco.editor.create(container, {
			value: content,
			language: getLanguage(filePath),
			theme: 'vs',
			automaticLayout: true,
			minimap: { enabled: false },
			fontSize: 14,
			lineNumbers: 'on',
			scrollBeyondLastLine: false,
			wordWrap: 'on',
			tabSize: 2,
			insertSpaces: false,
			renderWhitespace: 'selection',
		});

		editor.onDidChangeModelContent(() => {
			content = editor.getValue();
		});

		// Load file content
		await loadFile();

		// Add keydown listener
		window.addEventListener('keydown', handleKeydown);
	});

	onDestroy(() => {
		if (browser) {
			window.removeEventListener('keydown', handleKeydown);
		}
		if (editor) {
			editor.dispose();
		}
	});

	// Reload when file path changes
	$effect(() => {
		if (filePath && editor) {
			loadFile();
		}
	});
</script>

<div class="h-full flex flex-col bg-white">
	<!-- Header -->
	<div class="flex items-center justify-between px-4 py-2 border-b border-slate-200 bg-slate-50">
		<div class="flex items-center gap-2 min-w-0">
			<svg class="w-4 h-4 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
			</svg>
			<span class="text-sm font-mono text-slate-700 truncate">{filePath.split('/').pop()}</span>
			{#if hasChanges}
				<span class="w-2 h-2 rounded-full bg-vanna-orange flex-shrink-0" title="Unsaved changes"></span>
			{/if}
		</div>

		<div class="flex items-center gap-2">
			{#if error}
				<span class="text-xs text-red-600">{error}</span>
			{/if}
			<button
				class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed {hasChanges ? 'bg-vanna-teal text-white hover:bg-vanna-teal/90' : 'bg-slate-100 text-slate-400'}"
				disabled={!hasChanges || saving}
				onclick={saveFile}
			>
				{#if saving}
					<span class="flex items-center gap-1.5">
						<svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Saving...
					</span>
				{:else}
					Save
				{/if}
			</button>
		</div>
	</div>

	<!-- Editor -->
	<div class="flex-1 relative">
		{#if loading}
			<div class="absolute inset-0 flex items-center justify-center bg-white/80">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
			</div>
		{/if}
		<div bind:this={container} class="h-full w-full"></div>
	</div>
</div>
