<script>
	import { onMount } from 'svelte';

	let terminalElement;

	onMount(() => {
		// Load xterm from CDN
		const script1 = document.createElement('script');
		script1.src = 'https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js';
		document.head.appendChild(script1);

		const script2 = document.createElement('script');
		script2.src = 'https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js';
		document.head.appendChild(script2);

		script2.onload = () => {
			const term = new window.Terminal({
				cursorBlink: true,
				fontSize: 14,
				fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
			});

			const fitAddon = new window.FitAddon.FitAddon();
			term.loadAddon(fitAddon);

			term.open(terminalElement);
			fitAddon.fit();

			const ws = new WebSocket('ws://localhost:8080/ws');

			ws.onopen = function() {
				console.log('WebSocket connected');
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
		};
	});
</script>

<svelte:head>
	<title>Remote-Code Terminal</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Header with back button -->
		<div class="mb-6">
			<div class="flex items-center gap-4 mb-4">
				<a href="/" class="flex items-center gap-2 text-gray-400 hover:text-white transition-colors">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
					</svg>
					<span>Back to Dashboard</span>
				</a>
			</div>
			<div class="flex items-center gap-4">
				<img 
					src="https://remote-code.com/static/images/banner.svg" 
					alt="Remote-Code Logo" 
					class="h-12 w-auto"
				/>
				<div>
					<h1 class="text-3xl font-bold text-green-400 mb-1">Terminal</h1>
					<p class="text-gray-300">Interactive terminal with tmux session management</p>
				</div>
			</div>
		</div>
		
		<div class="bg-black rounded-lg border border-gray-700 p-4 shadow-xl">
			<div 
				id="terminal" 
				bind:this={terminalElement}
				class="w-full h-[80vh] focus:outline-none"
			></div>
		</div>
		
		<div class="mt-4 text-sm text-gray-400">
			<p>Connected to tmux session: <span class="text-green-400 font-mono">remote-code</span></p>
		</div>
	</div>
</div>