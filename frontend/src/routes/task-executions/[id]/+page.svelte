<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import StatusBadge from '$lib/components/ui/StatusBadge.svelte';

	$: breadcrumbSegments = [
		{ label: "", href: "/", icon: "banner" },
		{ label: "Projects", href: "/projects" },
		{ label: execution?.project_name || "Project", href: execution?.project_id ? `/projects/${execution.project_id}` : "/projects" },
		{ label: execution?.task_title ? (execution.task_title.length > 20 ? execution.task_title.substring(0, 20) + "..." : execution.task_title) : "Task", href: execution?.task_id ? `/task-executions?task_id=${execution.task_id}` : "#" },
		{ label: execution?.agent_name || `Agent`, href: `/task-executions/${$page.params.id}` }
	];

	let terminalElement;
	let devTerminalElement;
	let ws;
	let devWs;
	let term;
	let devTerm;
	let fitAddon;
	let canvasAddon;
	let devFitAddon;
	let devCanvasAddon;
	let execution = null;
	let loading = true;
	let error = null;
	let devServerRunning = false;
	let showDevTerminal = false;
	let inputText = '';
	let isSendingInput = false;
	let isResendingTask = false;
	let isDeleting = false;
	let isAccepting = false;

	$: executionId = $page.params.id;

	onMount(() => {
		loadTaskExecutionDetails();

		// Periodically refresh task execution status to update waiting status (lightweight)
		const executionRefreshInterval = setInterval(async () => {
			// Only refresh if not currently loading to avoid conflicts
			if (!loading) {
				await refreshExecutionStatus();
			}
		}, 10000); // Refresh every 10 seconds

		return () => {
			if (executionRefreshInterval) {
				clearInterval(executionRefreshInterval);
			}
			if (ws) {
				ws.close();
			}
			if (devWs) {
				devWs.close();
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
			if (devTerm) {
				try {
					// Dispose dev terminal addons first
					if (devCanvasAddon) {
						devCanvasAddon.dispose();
						devCanvasAddon = null;
					}
					if (devFitAddon) {
						devFitAddon.dispose();
						devFitAddon = null;
					}
					devTerm.dispose();
					devTerm = null;
				} catch (error) {
					console.warn('Error disposing dev terminal:', error);
				}
			}
		};
	});

    async function loadTaskExecutionDetails() {
        try {
            loading = true;
            error = null;

            // Get task execution details with all related information
            const response = await fetch(`/api/task-executions/${executionId}`);
            if (response.ok) {
                execution = await response.json();

				// Check if dev server is running (based on dev_server_tmux_id)
				devServerRunning = execution.dev_server_tmux_id?.Valid || false;
				showDevTerminal = devServerRunning;

				// Initialize terminal once we have execution info (only if not already initialized)
				if (!term) {
					initializeTerminal();
				}

				// Initialize dev terminal if dev server is running (only if not already initialized)
				if (showDevTerminal && !devTerm) {
					initializeDevTerminal();
				}
			} else {
				error = `Task execution ${executionId} not found`;
			}
		} catch (err) {
			console.error('Failed to load task execution details:', err);
			error = 'Failed to load task execution details';
		} finally {
			loading = false;
		}
	}

	// Lightweight function to refresh only the execution status without causing flicker
	async function refreshExecutionStatus() {
		if (!execution) return;

		try {
			// Get task execution details without setting loading state
			const response = await fetch(`/api/task-executions/${executionId}`);
			if (response.ok) {
				const updatedExecution = await response.json();
				// Only update specific fields that might change status
				if (execution) {
					execution.status = updatedExecution.status;
					execution.updated_at = updatedExecution.updated_at;
					// Update the execution object reactively
					execution = { ...execution, status: updatedExecution.status, updated_at: updatedExecution.updated_at };
				}
			}
		} catch (err) {
			// Silently fail - this is just a status check
			console.warn('Failed to refresh execution status:', err);
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
			allowProposedApi: true,
		});

		fitAddon = new window.FitAddon.FitAddon();
		canvasAddon = new window.CanvasAddon.CanvasAddon();
		term.loadAddon(fitAddon);
		term.loadAddon(canvasAddon);

		term.open(terminalElement);
		fitAddon.fit();

		// Create WebSocket connection for the task execution session
		const sessionName = `task_${execution.task_id}_agent_${execution.agent_id}`;
		const wsProtocol = $page.url.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${wsProtocol}//${$page.url.host}/ws?session=${sessionName}`;
		ws = new WebSocket(wsUrl);
		ws.binaryType = 'arraybuffer';

		ws.onopen = function() {
			console.log('WebSocket connected to task execution session:', sessionName);
			// Send initial size
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
		};

		ws.onclose = function() {
			console.log('WebSocket disconnected');
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

	function initializeDevTerminal() {
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

			script4.onload = () => createDevTerminal();
		} else {
			createDevTerminal();
		}
	}

	function createDevTerminal() {
		// Wait for the dev terminal element to be available
		if (!devTerminalElement) {
			console.error('Dev terminal element not available yet');
			setTimeout(createDevTerminal, 100);
			return;
		}

		devTerm = new window.Terminal({
			cursorBlink: true,
			fontSize: 14,
			fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
			lineHeight: 1.0,
			letterSpacing: 0,
			allowTransparency: false,
			allowProposedApi: true,
		});

		devFitAddon = new window.FitAddon.FitAddon();
		devCanvasAddon = new window.CanvasAddon.CanvasAddon();
		devTerm.loadAddon(devFitAddon);
		devTerm.loadAddon(devCanvasAddon);
		if (window.Unicode11Addon) {
			const unicode11 = new window.Unicode11Addon.Unicode11Addon();
			devTerm.loadAddon(unicode11);
			devTerm.unicode.activeVersion = '11';
		}

		devTerm.open(devTerminalElement);
		devFitAddon.fit();

		// Create WebSocket connection for the dev server session (uses execution ID)
		const devSessionName = `dev_${executionId}`;
		const wsProtocol = $page.url.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${wsProtocol}//${$page.url.host}/ws?session=${devSessionName}`;
		devWs = new WebSocket(wsUrl);
		devWs.binaryType = 'arraybuffer';

		devWs.onopen = function() {
			console.log('WebSocket connected to dev server session:', devSessionName);
			try {
				if (devTerm) {
					devWs.send(JSON.stringify({ type: 'resize', cols: devTerm.cols, rows: devTerm.rows }));
					// Kick the shell to redraw prompt
					devWs.send('\r');
				}
			} catch (e) {
				console.warn('Failed to send dev resize on open:', e);
			}
		};

		const devDecoder = new TextDecoder('utf-8');
		devWs.onmessage = function(event) {
			if (typeof event.data === 'string') {
				devTerm.write(event.data);
				return;
			}
			const text = devDecoder.decode(new Uint8Array(event.data), { stream: true });
			if (text) devTerm.write(text);
		};

		devWs.onerror = function(error) {
			console.error('Dev WebSocket error:', error);
		};

		devWs.onclose = function() {
			console.log('Dev WebSocket disconnected');
		};

		devTerm.onData(function(data) {
			if (devWs && devWs.readyState === WebSocket.OPEN) {
				devWs.send(data);
			}
		});

		const devSendResize = () => {
			if (devWs && devWs.readyState === WebSocket.OPEN && devTerm && devFitAddon) {
				devWs.send(JSON.stringify({ type: 'resize', cols: devTerm.cols, rows: devTerm.rows }));
			}
		};
		let devResizeTimeout;
		const handleResize = () => {
			devFitAddon.fit();
			clearTimeout(devResizeTimeout);
			devResizeTimeout = setTimeout(devSendResize, 50);
		};

		window.addEventListener('resize', handleResize);
	}

	async function runDevServer() {
		try {
			const response = await fetch(`/api/task-executions/${executionId}/dev-server`, {
				method: 'POST'
			});

			if (response.ok) {
				devServerRunning = true;
				showDevTerminal = true;

				// Initialize dev terminal after a short delay to ensure the session is created
				setTimeout(() => {
					initializeDevTerminal();
				}, 1000);
			}
		} catch (err) {
			console.error('Failed to start dev server:', err);
			alert('Failed to start dev server');
		}
	}

	async function stopDevServer() {
		try {
			const response = await fetch(`/api/task-executions/${executionId}/dev-server`, {
				method: 'DELETE'
			});

			if (response.ok) {
				devServerRunning = false;
				showDevTerminal = false;

				// Close dev terminal connections
				if (devWs) {
					devWs.close();
				}
				if (devTerm) {
					try {
						// Dispose dev terminal addons first
						if (devCanvasAddon) {
							devCanvasAddon.dispose();
							devCanvasAddon = null;
						}
						if (devFitAddon) {
							devFitAddon.dispose();
							devFitAddon = null;
						}
						devTerm.dispose();
						devTerm = null;
					} catch (error) {
						console.warn('Error disposing dev terminal:', error);
					}
				}
			}
		} catch (err) {
			console.error('Failed to stop dev server:', err);
			alert('Failed to stop dev server');
		}
	}

	function getStatusColor(status) {
		switch (status?.toLowerCase()) {
			case 'completed': return 'text-green-400 bg-green-500/20 border-green-500';
			case 'running': return 'text-blue-400 bg-blue-500/20 border-blue-500';
			case 'waiting': return 'text-yellow-400 bg-yellow-500/20 border-yellow-500';
			case 'failed': return 'text-red-400 bg-red-500/20 border-red-500';
			case 'pending': return 'text-gray-400 bg-gray-500/20 border-gray-500';
			default: return 'text-gray-400 bg-gray-500/20 border-gray-500';
		}
	}

	function formatDate(dateString) {
		return new Date(dateString).toLocaleString();
	}

	async function sendInputToSession() {
		if (!inputText.trim() || isSendingInput) return;

		try {
			isSendingInput = true;
			const response = await fetch(`/api/task-executions/${executionId}/send-input`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					input: inputText
				})
			});

			if (response.ok) {
				inputText = ''; // Clear the input
			} else {
				const errorData = await response.text();
				alert(`Failed to send input: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to send input:', err);
			alert('Failed to send input to session');
		} finally {
			isSendingInput = false;
		}
	}

	async function resendTaskToSession() {
		if (isResendingTask) return;

		try {
			isResendingTask = true;
			const response = await fetch(`/api/task-executions/${executionId}/resend-task`, {
				method: 'POST'
			});

			if (response.ok) {
				const result = await response.json();
				// Optional: Show a success message
				console.log('Task re-sent successfully:', result);
			} else {
				const errorData = await response.text();
				alert(`Failed to re-send task: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to re-send task:', err);
			alert('Failed to re-send task to session');
		} finally {
			isResendingTask = false;
		}
	}

	function handleInputKeydown(event) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			sendInputToSession();
		}
	}

	async function deleteTaskExecution() {
		if (isDeleting) return;

		// Show confirmation dialog
		const confirmed = confirm(`Are you sure you want to delete this task execution? This will:\n\n- Kill all associated tmux sessions\n- Run teardown commands\n- Delete all related data\n\nThis action cannot be undone.`);

		if (!confirmed) return;

		try {
			isDeleting = true;
			const response = await fetch(`/api/task-executions/${executionId}`, {
				method: 'DELETE'
			});

			if (response.ok) {
				// Navigate back to tasks list
				goto('/task-executions');
			} else {
				const errorData = await response.text();
				alert(`Failed to delete task execution: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to delete task execution:', err);
			alert('Failed to delete task execution');
		} finally {
			isDeleting = false;
		}
	}

	async function acceptTaskExecution() {
		if (isAccepting) return;

		const confirmed = confirm(`Are you sure you want to accept this task execution? This will:\n\n- Kill all associated tmux sessions\n- Run teardown commands\n- Move the task to "To Verify" status`);

		if (!confirmed) return;

		try {
			isAccepting = true;
			const response = await fetch(`/api/task-executions/${executionId}/accept`, {
				method: 'POST'
			});

			if (response.ok) {
				const data = await response.json();
				// Redirect to project page
				goto(`/projects/${data.project_id}`);
			} else {
				const errorData = await response.text();
				alert(`Failed to accept task execution: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to accept task execution:', err);
			alert('Failed to accept task execution');
		} finally {
			isAccepting = false;
		}
	}
</script>

<svelte:head>
	<title>Task Execution {executionId} - Remote-Code</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
</svelte:head>

<div class="min-h-screen">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />

		<!-- Header -->
		<div class="mb-6">
			{#if loading}
				<div class="border-b border-slate-200 pb-6 mb-8">
					<div class="flex items-center gap-4">
						<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
						<div>
							<h1 class="text-2xl font-bold text-vanna-navy font-serif">Loading Task Execution...</h1>
							<p class="text-slate-500">Fetching execution details</p>
						</div>
					</div>
				</div>
			{:else if error}
				<div class="border-b border-slate-200 pb-6 mb-8">
					<div class="flex items-center gap-4">
						<svg class="w-8 h-8 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
						</svg>
						<div>
							<h1 class="text-2xl font-bold text-vanna-orange font-serif">Execution Not Found</h1>
							<p class="text-slate-500">{error}</p>
						</div>
					</div>
				</div>
			{:else if execution}
				<div class="border-b border-slate-200 pb-6 mb-8">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-4 mb-2">
								<div class="w-12 h-12 bg-vanna-teal rounded-xl flex items-center justify-center">
									<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
									</svg>
								</div>
								<div>
									<div class="flex items-center gap-3">
										<h1 class="text-2xl font-bold text-vanna-navy font-serif sm:text-3xl">
											{execution.task_title || `Task ${execution.task_id}`}
										</h1>
										<StatusBadge status={execution.status?.toLowerCase()} />
									</div>
									<p class="mt-1 text-sm text-slate-500 flex items-center gap-2">
										<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal border border-vanna-teal/30">
											{execution.agent_name || `Agent ${execution.agent_id}`}
										</span>
										<span>working in</span>
										<code class="px-2 py-1 text-xs font-mono bg-vanna-cream/50 text-vanna-navy rounded border border-slate-200">
											{execution.base_directory_path || 'unknown path'}
										</code>
									</p>
								</div>
							</div>
						</div>

						<div class="flex items-center gap-3 ml-6">
							<div class="flex gap-2">
								<Button
									variant={devServerRunning ? 'secondary' : 'success'}
									onclick={() => runDevServer()}
									disabled={devServerRunning}
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
									</svg>
									{devServerRunning ? 'Dev Server Running' : 'Run Dev Server'}
								</Button>

								{#if devServerRunning}
									<Button
										variant="danger"
										onclick={() => stopDevServer()}
									>
										<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 9l6 6m0-6l-6 6"/>
										</svg>
										Stop Dev Server
									</Button>
								{/if}

								<Button
									variant="success"
									onclick={acceptTaskExecution}
									disabled={isAccepting || execution.status === 'completed' || execution.status === 'rejected'}
									loading={isAccepting}
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
									</svg>
									{isAccepting ? 'Accepting...' : 'Accept'}
								</Button>

								<Button
									variant="danger"
									onclick={deleteTaskExecution}
									disabled={isDeleting}
									loading={isDeleting}
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
									</svg>
									{isDeleting ? 'Deleting...' : 'Delete'}
								</Button>
							</div>
						</div>
					</div>
				</div>

				{#if execution.status?.toLowerCase() === 'waiting'}
					<Card class="mb-6 border-vanna-orange/30 bg-vanna-orange/5">
						<div class="flex items-center text-vanna-orange">
							<svg class="w-5 h-5 mr-3 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
								<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
							</svg>
							<strong>Agent may be waiting for user input.</strong> Check the terminal below.
						</div>
					</Card>
				{/if}

				<!-- Re-send Task Button (always show if running) -->
				{#if execution.status === 'running'}
					<Card class="mb-6">
						<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
							<h3 class="text-lg font-semibold text-vanna-navy">Task Controls</h3>
							<Button
								variant="primary"
								onclick={resendTaskToSession}
								disabled={isResendingTask}
								loading={isResendingTask}
								size="sm"
							>
								<svg class="w-3 h-3 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
								</svg>
								{isResendingTask ? 'Sending...' : 'Re-Send Task'}
							</Button>
						</div>
					</Card>
				{/if}

				<!-- Task Description -->
				{#if execution.task_description}
					<Card class="mb-6">
						<h3 class="text-lg font-semibold text-vanna-navy mb-3">Task Description</h3>
						<p class="text-slate-600">{execution.task_description}</p>
					</Card>
				{/if}
			{/if}
		</div>

    <!-- Terminal -->
		{#if !loading && !error && execution}
			<Card padding="none" class="bg-black border-slate-600 shadow-xl">
				<div class="flex items-center justify-between p-4 border-b border-slate-600">
					<div class="flex items-center gap-3">
						<div class="flex gap-1">
							<div class="w-3 h-3 rounded-full bg-vanna-orange"></div>
							<div class="w-3 h-3 rounded-full bg-vanna-magenta"></div>
							<div class="w-3 h-3 rounded-full bg-vanna-teal"></div>
						</div>
						<span class="text-gray-400 text-sm font-mono">tmux attach -t task_{execution.task_id}_agent_{execution.agent_id}</span>
					</div>
					<div class="flex items-center gap-2 text-xs text-gray-400">
						<div class="w-2 h-2 rounded-full bg-vanna-teal"></div>
						<span>Connected</span>
					</div>
				</div>
				<div class="p-4">
					<div
						id="terminal"
						bind:this={terminalElement}
						class="w-full h-[50vh] sm:h-[60vh] focus:outline-none"
					></div>
				</div>
			</Card>

			<div class="mt-4 text-sm text-slate-500 text-center">
				<p>Terminal connected to task execution session</p>
			</div>

			<!-- Input Section for Sending Text to Session -->
			{#if execution.status === 'running'}
				<Card class="mt-6">
					<h4 class="text-sm font-semibold text-vanna-navy mb-3 flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
						</svg>
						Send Input to Session
					</h4>
					<div class="flex flex-col sm:flex-row gap-2">
						<input
							type="text"
							bind:value={inputText}
							on:keydown={handleInputKeydown}
							placeholder="Type a message and press Enter to send..."
							disabled={isSendingInput}
							class="flex-1 w-full rounded-lg border border-slate-300 bg-white text-vanna-navy px-3 py-2 text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:border-vanna-teal focus:ring-vanna-teal disabled:opacity-50 disabled:cursor-not-allowed"
						/>
						<Button
							variant="primary"
							onclick={sendInputToSession}
							disabled={!inputText.trim() || isSendingInput}
							loading={isSendingInput}
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
							</svg>
							{isSendingInput ? '' : 'Send'}
						</Button>
					</div>
					<p class="text-xs text-slate-500 mt-2">
						Send text input directly to the agent session. Press Enter or click Send.
					</p>
				</Card>
			{/if}

			<!-- Dev Server Terminal -->
			{#if showDevTerminal}
				<Card padding="none" class="mt-6 bg-black border-vanna-teal shadow-xl">
					<div class="flex items-center justify-between p-4 border-b border-vanna-teal">
						<div class="flex items-center gap-3">
							<div class="flex gap-1">
								<div class="w-3 h-3 rounded-full bg-vanna-orange"></div>
								<div class="w-3 h-3 rounded-full bg-vanna-magenta"></div>
								<div class="w-3 h-3 rounded-full bg-vanna-teal"></div>
							</div>
							<span class="text-vanna-teal text-sm font-mono">tmux attach -t dev_{executionId}</span>
							<span class="px-2 py-1 rounded text-xs font-semibold bg-vanna-teal/20 text-vanna-teal border border-vanna-teal">DEV SERVER</span>
						</div>
						<div class="flex items-center gap-2 text-xs text-vanna-teal">
							<div class="w-2 h-2 rounded-full bg-vanna-teal animate-pulse"></div>
							<span>Running</span>
						</div>
					</div>
					<div class="p-4">
						<div
							id="dev-terminal"
							bind:this={devTerminalElement}
							class="w-full h-[40vh] sm:h-[50vh] focus:outline-none"
						></div>
					</div>
				</Card>

				<div class="mt-4 text-sm text-vanna-teal text-center">
					<p>Dev server terminal - showing startup logs and output</p>
				</div>
			{/if}
		{:else if error}
			<Card class="border-vanna-orange/30 bg-vanna-orange/5 p-8 text-center">
				<svg class="w-16 h-16 text-vanna-orange mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<h3 class="text-xl font-semibold text-vanna-orange mb-2">Task Execution Not Available</h3>
				<p class="text-slate-600 mb-6">{error}</p>
				<div class="flex gap-3 justify-center">
					<Button href="/task-executions" variant="secondary">
						View All Executions
					</Button>
					<Button href="/" variant="primary">
						Go to Dashboard
					</Button>
				</div>
			</Card>
		{/if}
	</div>
</div>
