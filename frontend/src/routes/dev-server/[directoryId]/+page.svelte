<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	let terminalElement;
	let ws;
	let term;
	let fitAddon;
	let canvasAddon;
	let directory = null;
	let devServerStatus = null;
	let loading = true;
	let error = null;
	let isConnected = false;
	let terminalReady = false;
	let connectionError = null;
	let stoppingDevServer = false;

	$: directoryId = $page.params.directoryId;
	$: sessionId = devServerStatus?.tmux_session_id;
	$: breadcrumbSegments = [
		{ label: "", href: "/", icon: "banner" },
		{ label: "Directories", href: "/directories" },
		{ label: "Dev Server", href: `/dev-server/${directoryId}` }
	];

	onMount(() => {
		loadData();

		return () => {
			if (ws) {
				ws.close();
			}
			isConnected = false;
			terminalReady = false;
			connectionError = null;
			if (term) {
				try {
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

	async function loadData() {
		try {
			loading = true;
			error = null;

			// Load directory info
			const dirRes = await fetch(`/api/base-directories/${directoryId}`);
			if (!dirRes.ok) {
				error = 'Directory not found';
				return;
			}
			directory = await dirRes.json();

			// Load dev server status
			const devRes = await fetch(`/api/base-directories/${directoryId}/dev-server`);
			if (devRes.ok) {
				devServerStatus = await devRes.json();
			}

			if (!devServerStatus?.running) {
				error = 'Dev server is not running for this directory';
				return;
			}

			// Initialize terminal once we have session info
			initializeTerminal();
		} catch (err) {
			console.error('Failed to load:', err);
			error = 'Failed to load dev server information';
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

			const script4 = document.createElement('script');
			script4.src = 'https://cdn.jsdelivr.net/npm/@xterm/addon-unicode11@0.8.0/lib/addon-unicode11.js';
			document.head.appendChild(script4);

			script4.onload = () => createTerminal();
		} else {
			createTerminal();
		}
	}

	function createTerminal() {
		// Wait for the terminal element to be available
		if (!terminalElement) {
			console.error('Terminal element not available yet');
			setTimeout(createTerminal, 100);
			return;
		}

		// Reset connection state
		isConnected = false;
		terminalReady = false;
		connectionError = null;

		try {
			if (!window.Terminal) {
				throw new Error('Terminal library not loaded');
			}

			term = new window.Terminal({
				cursorBlink: true,
				fontSize: 14,
				fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
				lineHeight: 1.0,
				letterSpacing: 0,
				allowTransparency: false,
				allowProposedApi: true,
			});
		} catch (error) {
			console.error('Failed to create terminal:', error);
			connectionError = 'Failed to initialize terminal: ' + error.message;
			return;
		}

		try {
			if (!window.FitAddon || !window.CanvasAddon) {
				throw new Error('Terminal addons not loaded');
			}

			fitAddon = new window.FitAddon.FitAddon();
			canvasAddon = new window.CanvasAddon.CanvasAddon();
			term.loadAddon(fitAddon);
			term.loadAddon(canvasAddon);
			if (window.Unicode11Addon) {
				const unicode11 = new window.Unicode11Addon.Unicode11Addon();
				term.loadAddon(unicode11);
				term.unicode.activeVersion = '11';
			}

			term.open(terminalElement);
			fitAddon.fit();
			try {
				if (ws && ws.readyState === WebSocket.OPEN) {
					ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
				}
			} catch (e) {
				console.warn('Failed to send initial resize:', e);
			}
			terminalReady = true;
		} catch (error) {
			console.error('Failed to initialize terminal addons:', error);
			connectionError = 'Failed to initialize terminal addons: ' + error.message;
			return;
		}

		// Create WebSocket connection for dev server session
		const wsProtocol = $page.url.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${wsProtocol}//${$page.url.host}/ws?session=${sessionId}`;
		ws = new WebSocket(wsUrl);
		ws.binaryType = 'arraybuffer';

		ws.onopen = function() {
			console.log('WebSocket connected to dev server session:', sessionId);
			isConnected = true;
			try {
				if (term) {
					ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
				}
			} catch (e) {
				console.warn('Failed to send resize on open:', e);
			}
		};

		const decoder = new TextDecoder('utf-8');
		ws.onmessage = function(event) {
			if (typeof event.data === 'string') {
				term.write(event.data);
				return;
			}
			const text = decoder.decode(new Uint8Array(event.data), { stream: true });
			if (text) term.write(text);
		};

		ws.onerror = function(error) {
			console.error('WebSocket error:', error);
			isConnected = false;
			connectionError = 'WebSocket connection error';
		};

		ws.onclose = function() {
			console.log('WebSocket disconnected');
			isConnected = false;
		};

		term.onData(function(data) {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(data);
			}
		});

		const sendResize = () => {
			if (ws && ws.readyState === WebSocket.OPEN && term && fitAddon) {
				ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
			}
		};

		let resizeTimeout;
		const handleResize = () => {
			fitAddon.fit();
			clearTimeout(resizeTimeout);
			resizeTimeout = setTimeout(sendResize, 50);
		};

 		window.addEventListener('resize', handleResize);
 	}

	async function stopDevServer() {
		if (!confirm('Are you sure you want to stop the dev server?')) {
			return;
		}

		stoppingDevServer = true;
		try {
			const res = await fetch(`/api/base-directories/${directoryId}/dev-server`, {
				method: 'DELETE'
			});
			if (res.ok) {
				goto(`/files/${directoryId}`);
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
</script>

<svelte:head>
	<title>Dev Server: {directory?.path || 'Loading...'} - Remote-Code</title>
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
						<h1 class="text-2xl font-bold text-green-400">Loading Dev Server...</h1>
						<p class="text-gray-300">Connecting to terminal</p>
					</div>
				</div>
			{:else if error}
				<div class="flex items-center gap-4">
					<svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
					<div>
						<h1 class="text-2xl font-bold text-red-400">Dev Server Not Available</h1>
						<p class="text-gray-300">{error}</p>
					</div>
				</div>
			{:else if directory && devServerStatus}
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 bg-green-500 rounded-xl flex items-center justify-center">
							<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
							</svg>
						</div>
						<div>
							<div class="flex items-center gap-3 mb-1">
								<h1 class="text-2xl font-bold text-green-400">Dev Server Terminal</h1>
								<span class="px-2 py-1 rounded text-xs font-semibold bg-green-500 text-black flex items-center gap-1">
									<span class="w-1.5 h-1.5 bg-white rounded-full animate-pulse"></span>
									RUNNING
								</span>
							</div>
							<div class="flex items-center gap-4 text-sm text-gray-400">
								<span class="font-mono">{directory.path}</span>
								<span>Session: <span class="text-green-400 font-mono">{sessionId}</span></span>
							</div>
						</div>
					</div>

					<div class="flex gap-2">
						<button
							on:click={() => goto(`/files/${directoryId}`)}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Back to Files
						</button>
						<button
							on:click={stopDevServer}
							disabled={stoppingDevServer}
							class="bg-red-500 hover:bg-red-600 disabled:bg-red-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
						>
							{#if stoppingDevServer}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Stopping...
							{:else}
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
								</svg>
								Stop Server
							{/if}
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Terminal -->
		{#if !loading && !error && directory && devServerStatus?.running}
			<div class="bg-black rounded-lg border border-gray-700 p-4 shadow-xl">
				<div
					id="terminal"
					bind:this={terminalElement}
					class="w-full h-[80vh] focus:outline-none"
				></div>
			</div>

			<div class="mt-4 text-sm flex items-center gap-2">
				{#if connectionError}
					<div class="flex items-center gap-2 text-red-400">
						<div class="w-2 h-2 bg-red-400 rounded-full"></div>
						<span>Error: {connectionError}</span>
					</div>
				{:else if isConnected && terminalReady}
					<div class="flex items-center gap-2 text-green-400">
						<div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
						<span>Connected to dev server: <span class="font-mono">{sessionId}</span></span>
					</div>
				{:else if isConnected && !terminalReady}
					<div class="flex items-center gap-2 text-yellow-400">
						<div class="w-2 h-2 bg-yellow-400 rounded-full animate-pulse"></div>
						<span>Initializing terminal...</span>
					</div>
				{:else}
					<div class="flex items-center gap-2 text-red-400">
						<div class="w-2 h-2 bg-red-400 rounded-full"></div>
						<span>Disconnected from dev server</span>
					</div>
				{/if}
			</div>
		{:else if error}
			<div class="bg-gray-800 rounded-lg border border-red-600 p-8 text-center">
				<svg class="w-16 h-16 text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<h3 class="text-xl font-semibold text-red-400 mb-2">Dev Server Not Running</h3>
				<p class="text-gray-400 mb-6">{error}</p>
				<div class="flex gap-3 justify-center">
					<a
						href="/files/{directoryId}"
						class="bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg transition-colors"
					>
						Go to Files
					</a>
					<a
						href="/directories"
						class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg transition-colors"
					>
						View Directories
					</a>
				</div>
			</div>
		{/if}
	</div>
</div>
