<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';

	let terminalElement;
	let ws;
	let term;
	let fitAddon;
	let canvasAddon;
	let session = null;
	let sessionInfo = null;
	let loading = true;
	let error = null;

	$: sessionId = $page.params.session;
	$: breadcrumbSegments = [
		{ label: "Dashboard", href: "/" },
		{ label: "Terminal", href: "/terminal" },
		{ label: sessionId || "Session", href: `/terminal/${sessionId}` }
	];

	onMount(() => {
		loadSessionInfo();
		
		return () => {
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
		};
	});

	async function loadSessionInfo() {
		try {
			loading = true;
			error = null;
			
			// Get session information
			const response = await fetch('/api/tmux-sessions');
			if (response.ok) {
				const sessions = await response.json();
				sessionInfo = sessions.find(s => s.name === sessionId);
				
				if (!sessionInfo) {
					error = `Session "${sessionId}" not found`;
					return;
				}
				
				// Initialize terminal once we have session info
				initializeTerminal();
			} else {
				error = 'Failed to load session information';
			}
		} catch (err) {
			console.error('Failed to load session info:', err);
			error = 'Failed to load session information';
		} finally {
			loading = false;
		}
	}

	function initializeTerminal() {
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

			script3.onload = () => createTerminal();
		} else {
			createTerminal();
		}
	}

	function createTerminal() {
		// Wait for the terminal element to be available
		if (!terminalElement) {
			console.error('Terminal element not available yet');
			// Try again after a short delay
			setTimeout(createTerminal, 100);
			return;
		}

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

		// Create WebSocket connection for specific session
		const wsUrl = `ws://localhost:8080/ws?session=${sessionId}`;
		ws = new WebSocket(wsUrl);

		ws.onopen = function() {
			console.log('WebSocket connected to session:', sessionId);
			// Backend handles session attachment automatically via query parameter
		};

		ws.onmessage = function(event) {
			term.write(event.data);
		};

		ws.onerror = function(error) {
			console.error('WebSocket error:', error);
		};

		ws.onclose = function() {
			console.log('WebSocket disconnected');
		};

		term.onData(function(data) {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(data);
			}
		});

		const handleResize = () => {
			fitAddon.fit();
		};

		window.addEventListener('resize', handleResize);
	}

	function formatTime(timestamp) {
		const date = new Date(parseInt(timestamp) * 1000);
		return date.toLocaleString();
	}

	function getSessionTypeInfo(session) {
		if (session.is_task) {
			return {
				icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
				</svg>`,
				color: 'text-yellow-400',
				badge: 'TASK SESSION',
				badgeColor: 'bg-yellow-500 text-black'
			};
		}
		return {
			icon: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
			</svg>`,
			color: 'text-green-400',
			badge: 'TMUX SESSION',
			badgeColor: 'bg-green-500 text-black'
		};
	}
</script>

<svelte:head>
	<title>Terminal: {sessionId} - Remote-Code</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />
		
		<!-- Header -->
		<div class="mb-6">
			
			{#if loading}
				<div class="flex items-center gap-4">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-green-400"></div>
					<div>
						<h1 class="text-2xl font-bold text-green-400">Loading Terminal Session...</h1>
						<p class="text-gray-300">Connecting to {sessionId}</p>
					</div>
				</div>
			{:else if error}
				<div class="flex items-center gap-4">
					<svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
					<div>
						<h1 class="text-2xl font-bold text-red-400">Session Not Found</h1>
						<p class="text-gray-300">{error}</p>
					</div>
				</div>
			{:else if sessionInfo}
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<img 
							src="https://remote-code.com/static/images/banner.svg" 
							alt="Remote-Code Logo" 
							class="h-10 w-auto"
						/>
						<div>
							<div class="flex items-center gap-3 mb-1">
								<div class="flex items-center gap-2 {getSessionTypeInfo(sessionInfo).color}">
									{@html getSessionTypeInfo(sessionInfo).icon}
									<h1 class="text-2xl font-bold font-mono">{sessionInfo.name}</h1>
								</div>
								<span class="px-2 py-1 rounded text-xs font-semibold {getSessionTypeInfo(sessionInfo).badgeColor}">
									{getSessionTypeInfo(sessionInfo).badge}
								</span>
							</div>
							<div class="flex items-center gap-4 text-sm text-gray-400">
								<span>Created: {formatTime(sessionInfo.created)}</span>
								{#if sessionInfo.is_task}
									<span>Task ID: <span class="text-yellow-400 font-mono">{sessionInfo.task_id}</span></span>
									<span>Agent: <span class="text-yellow-400 font-mono">{sessionInfo.agent_name || `ID: ${sessionInfo.agent_id}`}</span></span>
								{/if}
							</div>
						</div>
					</div>
					
					<div class="flex gap-2">
						<button 
							on:click={() => goto('/terminal')}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Session List
						</button>
						{#if sessionInfo.is_task}
							<button 
								on:click={() => goto('/projects')}
								class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
							>
								View Project
							</button>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Terminal -->
		{#if !loading && !error && sessionInfo}
			<div class="bg-black rounded-lg border border-gray-700 p-4 shadow-xl">
				<div 
					id="terminal" 
					bind:this={terminalElement}
					class="w-full h-[80vh] focus:outline-none"
				></div>
			</div>
			
			<div class="mt-4 text-sm text-gray-400">
				<p>Connected to tmux session: <span class="text-green-400 font-mono">{sessionInfo.name}</span></p>
			</div>
		{:else if error}
			<div class="bg-gray-800 rounded-lg border border-red-600 p-8 text-center">
				<svg class="w-16 h-16 text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<h3 class="text-xl font-semibold text-red-400 mb-2">Session Not Available</h3>
				<p class="text-gray-400 mb-6">{error}</p>
				<div class="flex gap-3 justify-center">
					<a 
						href="/terminal"
						class="bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg transition-colors"
					>
						View All Sessions
					</a>
					<a 
						href="/"
						class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg transition-colors"
					>
						Go to Dashboard
					</a>
				</div>
			</div>
		{/if}
	</div>
</div>
