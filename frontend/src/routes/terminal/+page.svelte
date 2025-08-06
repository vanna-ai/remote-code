<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	
	const breadcrumbSegments = [
		{ label: "Dashboard", href: "/" },
		{ label: "Terminal", href: "/terminal" }
	];

	let sessions = [];
	let selectedSession = null;
	let terminalElement;
	let ws;
	let term;
	let fitAddon;
	let canvasAddon;
	let loading = true;
	let showGlobalTerminal = false;

	onMount(() => {
		loadSessions();
		
		// Refresh sessions every 5 seconds
		const interval = setInterval(loadSessions, 5000);
		
		return () => {
			clearInterval(interval);
			if (ws) {
				ws.close();
			}
		};
	});

	async function loadSessions() {
		try {
			const response = await fetch('/api/tmux-sessions');
			if (response.ok) {
				sessions = await response.json();
			}
		} catch (error) {
			console.error('Failed to load tmux sessions:', error);
		} finally {
			loading = false;
		}
	}

	function attachToSession(session) {
		goto(`/terminal/${session.name}`);
	}

	function showGlobal() {
		selectedSession = null;
		showGlobalTerminal = true;
		initializeTerminal('global');
	}

	function initializeTerminal(sessionType) {
		// Close existing connection
		if (ws) {
			ws.close();
		}
		if (term) {
			try {
				// Dispose addons first to avoid cleanup race conditions
				if (canvasAddon) {
					canvasAddon.dispose();
					canvasAddon = null;
				}
				if (fitAddon) {
					fitAddon.dispose();
					fitAddon = null;
				}
				term.dispose();
				term = null;
			} catch (error) {
				console.warn('Error disposing terminal:', error);
			}
		}

		// Load xterm if not already loaded
		if (!window.Terminal) {
			const script1 = document.createElement('script');
			script1.src = 'https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js';
			document.head.appendChild(script1);

			const script2 = document.createElement('script');
			script2.src = 'https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js';
			document.head.appendChild(script2);

			const script3 = document.createElement('script');
			script3.src = 'https://cdn.jsdelivr.net/npm/xterm-addon-canvas@0.5.0/lib/xterm-addon-canvas.js';
			document.head.appendChild(script3);

			script3.onload = () => createTerminal(sessionType);
		} else {
			createTerminal(sessionType);
		}
	}

	function createTerminal(sessionType) {
		term = new window.Terminal({
			cursorBlink: true,
			fontSize: 14,
			fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
			lineHeight: 1.0,
			letterSpacing: 0,
			allowTransparency: false,
		});

		fitAddon = new window.FitAddon.FitAddon();
		canvasAddon = new window.CanvasAddon.CanvasAddon();
		term.loadAddon(fitAddon);
		term.loadAddon(canvasAddon);

		term.open(terminalElement);
		fitAddon.fit();

		// Create WebSocket connection
		const wsUrl = sessionType === 'global' 
			? 'ws://localhost:8080/ws'
			: `ws://localhost:8080/ws?session=${selectedSession.name}`;
		
		ws = new WebSocket(wsUrl);

		ws.onopen = function() {
			console.log('WebSocket connected');
			if (sessionType !== 'global') {
				// Attach to tmux session
				ws.send(`tmux attach -t ${selectedSession.name}\n`);
			}
		};

		ws.onmessage = function(event) {
			term.write(event.data);
		};

		ws.onerror = function(error) {
			console.error('WebSocket error:', error);
		};

		term.onData(function(data) {
			ws.send(data);
		});

		const handleResize = () => {
			fitAddon.fit();
		};

		window.addEventListener('resize', handleResize);
	}

	function closeTerminal() {
		if (ws) {
			ws.close();
		}
		if (term) {
			try {
				// Dispose addons first to avoid cleanup race conditions
				if (canvasAddon) {
					canvasAddon.dispose();
					canvasAddon = null;
				}
				if (fitAddon) {
					fitAddon.dispose();
					fitAddon = null;
				}
				term.dispose();
				term = null;
			} catch (error) {
				console.warn('Error disposing terminal:', error);
			}
		}
		selectedSession = null;
		showGlobalTerminal = false;
	}

	function formatTime(timestamp) {
		const date = new Date(parseInt(timestamp) * 1000);
		return date.toLocaleString();
	}

	function getSessionTypeIcon(session) {
		if (session.is_task) {
			return `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
			</svg>`;
		}
		return `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
		</svg>`;
	}

	$: taskSessions = sessions.filter(s => s.is_task);
	$: regularSessions = sessions.filter(s => !s.is_task);

	// Auto-scroll preview containers to bottom when sessions update
	$: if (sessions.length > 0) {
		// Use a small timeout to ensure DOM is updated
		setTimeout(() => {
			const previewContainers = document.querySelectorAll('.terminal-preview');
			previewContainers.forEach(container => {
				container.scrollTop = container.scrollHeight;
			});
		}, 10);
	}

</script>

<svelte:head>
	<title>Remote-Code Terminal</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
</svelte:head>

<style>
	/* Hide scrollbars while keeping scroll functionality */
	.scrollbar-hide {
		/* Firefox */
		scrollbar-width: none;
	}
	
	.scrollbar-hide::-webkit-scrollbar {
		/* Webkit browsers (Chrome, Safari, Edge) */
		display: none;
	}
</style>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />
		
		<!-- Header -->
		<div class="mb-6">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-4">
					<img 
						src="https://remote-code.com/static/images/banner.svg" 
						alt="Remote-Code Logo" 
						class="h-12 w-auto"
					/>
					<div>
						<h1 class="text-3xl font-bold text-green-400 mb-1">Terminal Sessions</h1>
						<p class="text-gray-300">Manage tmux sessions and terminal access</p>
					</div>
				</div>
				<button 
					on:click={showGlobal}
					class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					Global Terminal
				</button>
			</div>
		</div>

		<!-- Terminal View (when session is selected) -->
		{#if selectedSession || showGlobalTerminal}
			<div class="mb-6">
				<div class="flex items-center justify-between mb-4">
					<div class="flex items-center gap-3">
						{#if selectedSession}
							<div class="flex items-center gap-2 text-yellow-400">
								{@html getSessionTypeIcon(selectedSession)}
								<span class="font-mono">{selectedSession.name}</span>
								{#if selectedSession.is_task}
									<span class="text-sm text-gray-400">
										({selectedSession.task_name || `Task ${selectedSession.task_id}`} • {selectedSession.agent_name || `Agent ${selectedSession.agent_id}`})
									</span>
								{/if}
							</div>
						{:else}
							<div class="flex items-center gap-2 text-blue-400">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
								</svg>
								<span class="font-mono">Global Terminal</span>
							</div>
						{/if}
					</div>
					<button 
						on:click={closeTerminal}
						class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
					>
						Close
					</button>
				</div>
				
				<div class="bg-black rounded-lg border border-gray-700 p-4 shadow-xl">
					<div 
						id="terminal" 
						bind:this={terminalElement}
						class="w-full h-[70vh] focus:outline-none"
					></div>
				</div>
			</div>
		{:else}
			<!-- Session Grid -->
			<div class="space-y-8">
				<!-- Task Sessions -->
				{#if taskSessions.length > 0}
					<div>
						<h2 class="text-xl font-semibold text-yellow-400 mb-4 flex items-center gap-2">
							<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
							</svg>
							Task Sessions ({taskSessions.length})
						</h2>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each taskSessions as session}
								<div 
									class="bg-gray-800 rounded-lg border border-yellow-600 p-4 hover:border-yellow-400 cursor-pointer transition-colors"
									on:click={() => attachToSession(session)}
									role="button"
									tabindex="0"
								>
									<div class="flex items-center justify-between mb-3">
										<div class="flex items-center gap-2 text-yellow-400">
											{@html getSessionTypeIcon(session)}
											<span class="font-mono text-sm">{session.name}</span>
										</div>
										<span class="text-xs text-gray-400">{formatTime(session.created)}</span>
									</div>
									<div class="text-sm text-gray-300 mb-3">
										<div>Task: <span class="text-white font-mono">{session.task_name || `ID: ${session.task_id}`}</span></div>
										<div>Agent: <span class="text-white font-mono">{session.agent_name || `ID: ${session.agent_id}`}</span></div>
									</div>
									<div class="terminal-preview bg-gray-900 rounded p-2 text-xs font-mono text-gray-300 h-32 overflow-y-auto overflow-x-auto scrollbar-hide" style="scroll-behavior: smooth;">
										<pre class="whitespace-pre font-mono" style="margin: 0;">{@html session.preview}</pre>
									</div>
									<div class="mt-3 flex justify-end">
										<span class="bg-yellow-500 text-black px-2 py-1 rounded text-xs font-semibold">
											TASK SESSION
										</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Regular Sessions -->
				{#if regularSessions.length > 0}
					<div>
						<h2 class="text-xl font-semibold text-green-400 mb-4 flex items-center gap-2">
							<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
							</svg>
							Regular Sessions ({regularSessions.length})
						</h2>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each regularSessions as session}
								<div 
									class="bg-gray-800 rounded-lg border border-gray-600 p-4 hover:border-gray-500 cursor-pointer transition-colors"
									on:click={() => attachToSession(session)}
									role="button"
									tabindex="0"
								>
									<div class="flex items-center justify-between mb-3">
										<div class="flex items-center gap-2 text-green-400">
											{@html getSessionTypeIcon(session)}
											<span class="font-mono text-sm">{session.name}</span>
										</div>
										<span class="text-xs text-gray-400">{formatTime(session.created)}</span>
									</div>
									<div class="terminal-preview bg-gray-900 rounded p-2 text-xs font-mono text-gray-300 h-32 overflow-y-auto overflow-x-auto scrollbar-hide" style="scroll-behavior: smooth;">
										<pre class="whitespace-pre font-mono" style="margin: 0;">{@html session.preview}</pre>
									</div>
									<div class="mt-3 flex justify-end">
										<span class="bg-green-500 text-black px-2 py-1 rounded text-xs font-semibold">
											TMUX SESSION
										</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				{#if loading}
					<div class="text-center py-12">
						<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-400 mx-auto mb-4"></div>
						<p class="text-gray-400">Loading tmux sessions...</p>
					</div>
				{:else if sessions.length === 0}
					<div class="text-center py-12">
						<svg class="w-16 h-16 text-gray-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
						</svg>
						<h3 class="text-xl font-semibold text-gray-400 mb-2">No Active Sessions</h3>
						<p class="text-gray-500 mb-6">No tmux sessions are currently running.</p>
						<button 
							on:click={showGlobal}
							class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg transition-colors"
						>
							Start Global Terminal
						</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
