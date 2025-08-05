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
	<title>Web Terminal</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
	<style>
		body {
			margin: 0;
			padding: 20px;
			background-color: #000;
			font-family: monospace;
		}
	</style>
</svelte:head>

<div id="terminal" bind:this={terminalElement}></div>

<style>
	#terminal {
		width: 100%;
		height: 80vh;
	}
</style>
