<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	
	$: breadcrumbSegments = [
		{ label: "Remote-Code", href: "/", icon: "banner" },
		{ label: "Projects", href: "/projects" },
		{ label: execution?.project_name || "Project", href: execution?.project_id ? `/projects/${execution.project_id}` : "/projects" },
		{ label: execution?.task_title ? (execution.task_title.length > 20 ? execution.task_title.substring(0, 20) + "..." : execution.task_title) : "Task", href: execution?.task_id ? `/tasks?task_id=${execution.task_id}` : "#" },
		{ label: execution?.agent_name || `Agent`, href: `/tasks/${$page.params.id}` }
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
	let gitStatus = null;
	let loading = true;
	let error = null;
	let devServerRunning = false;
	let showDevTerminal = false;
	let inputText = '';
	let isSendingInput = false;
	let isResendingTask = false;
	let isDeleting = false;
    let gitStatusPoll = null;

	$: executionId = $page.params.id;

	onMount(() => {
		loadTaskExecutionDetails();
		// Poll git status every second
		gitStatusPoll = setInterval(() => {
			loadGitStatus();
		}, 1000);
		
		// Periodically refresh task execution details to update waiting status
		const executionRefreshInterval = setInterval(async () => {
			// Only refresh if not currently loading to avoid conflicts
			if (!loading) {
				await loadTaskExecutionDetails();
			}
		}, 10000); // Refresh every 10 seconds
		
		return () => {
			if (gitStatusPoll) {
				clearInterval(gitStatusPoll);
				gitStatusPoll = null;
			}
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
				
				// Check if dev server is running
				if (execution.worktree_id) {
					const worktreeResponse = await fetch(`/api/worktrees/${execution.worktree_id}`);
					if (worktreeResponse.ok) {
						const worktree = await worktreeResponse.json();
						devServerRunning = worktree.dev_server_tmux_id && worktree.dev_server_tmux_id.Valid;
						showDevTerminal = devServerRunning;
					}
				}
				
            // Load git status (used for merge and push controls)
            await loadGitStatus();
				
				// Initialize terminal once we have execution info
				initializeTerminal();
				
				// Initialize dev terminal if dev server is running
				if (showDevTerminal) {
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

    async function loadGitStatus() {
        if (!execution?.worktree_path) return;
        try {
            const res = await fetch(`/api/git/status?path=${encodeURIComponent(execution.worktree_path)}`);
            if (res.ok) {
                gitStatus = await res.json();
            } else {
                gitStatus = null;
            }
            // Also compute merge readiness
            try {
                const mr = await fetch(`/api/git/merge-ready?path=${encodeURIComponent(execution.worktree_path)}`);
                if (mr.ok) {
                    const data = await mr.json();
                    mergeReady = !!data.ready;
                    mergeReadyReason = data.ready ? '' : (data.reason || 'Not ready');
                } else {
                    mergeReady = false;
                    mergeReadyReason = 'Merge readiness check failed';
                }
            } catch (e) {
                mergeReady = false;
                mergeReadyReason = 'Merge readiness check failed';
            }
        } catch (err) {
            console.error('Failed to load git status:', err);
        }
    }

    async function stageFile(file) {
        if (!execution?.worktree_path) return;
        await fetch(`/api/git/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: execution.worktree_path, file })
        });
        await loadGitStatus();
    }

    async function unstageFile(file) {
        if (!execution?.worktree_path) return;
        await fetch(`/api/git/unstage`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: execution.worktree_path, file })
        });
        await loadGitStatus();
    }

    let diffOpen = false;
    let diffText = '';
    let diffTitle = '';

    async function viewDiff(file, staged=false) {
        if (!execution?.worktree_path) return;
        const url = `/api/git/diff?path=${encodeURIComponent(execution.worktree_path)}&file=${encodeURIComponent(file)}&staged=${staged}`;
        const res = await fetch(url);
        if (res.ok) {
            const data = await res.json();
            diffText = data.diff || '';
            diffTitle = `${staged ? 'Index vs HEAD' : 'Working Tree vs Index'} — ${file}`;
            diffOpen = true;
        }
    }

    let commitMsg = '';
    let committing = false;
    async function commitChanges() {
        if (!execution?.worktree_path || !commitMsg.trim()) return;
        committing = true;
        try {
            await fetch(`/api/git/commit`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: execution.worktree_path, message: commitMsg })
            });
            commitMsg = '';
            await loadGitStatus();
        } finally {
            committing = false;
        }
    }

	let merging = false;
    let pushing = false;
    let mergeReady = false;
    let mergeReadyReason = '';
    let updatingFromMain = false;
    async function mergeBranch() {
        if (!execution?.worktree_path || !gitStatus?.currentBranch) return;
        merging = true;
        try {
            const res = await fetch(`/api/git/merge`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: execution.worktree_path, branch: gitStatus.currentBranch, taskId: execution.task_id })
            });
            let data = null;
            try {
                data = await res.json();
            } catch (e) {
                // ignore
            }
            if (!res.ok || (data && data.ok === false)) {
                const details = data?.error || data?.output || 'Unknown error';
                const step = data?.step ? ` (step: ${data.step})` : '';
                alert(`Merge failed${step}: ${details}`);
                return;
            }
            await loadGitStatus();
        } finally {
            merging = false;
        }
    }

    async function pushChanges() {
        if (!execution?.worktree_path) return;
        pushing = true;
        try {
            const res = await fetch(`/api/git/push`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: execution.worktree_path })
            });
            let data = null;
            try { data = await res.json(); } catch (e) {}
            if (!res.ok || (data && data.ok === false)) {
                const details = data?.error || data?.output || 'Unknown error';
                const step = data?.step ? ` (step: ${data.step})` : '';
                alert(`Push failed${step}: ${details}`);
                return;
            }
            await loadGitStatus();
        } finally {
            pushing = false;
        }
    }

    async function updateFromMain(strategy = 'merge') {
        if (!execution?.worktree_path) return;
        updatingFromMain = true;
        try {
            const res = await fetch(`/api/git/update-from-main`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: execution.worktree_path, strategy })
            });
            let data = null;
            try { data = await res.json(); } catch (e) {}
            if (!res.ok || (data && data.ok === false)) {
                const details = data?.error || data?.output || 'Unknown error';
                const step = data?.step ? ` (step: ${data.step})` : '';
                alert(`Update from main failed${step}: ${details}`);
                return;
            }
            await loadGitStatus();
        } finally {
            updatingFromMain = false;
        }
    }

    // No branch dropdown; merges current worktree branch into base main

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

		ws.onopen = function() {
			console.log('WebSocket connected to task execution session:', sessionName);
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

			script3.onload = () => createDevTerminal();
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
		});

		devFitAddon = new window.FitAddon.FitAddon();
		devCanvasAddon = new window.CanvasAddon.CanvasAddon();
		devTerm.loadAddon(devFitAddon);
		devTerm.loadAddon(devCanvasAddon);

		devTerm.open(devTerminalElement);
		devFitAddon.fit();

		// Create WebSocket connection for the dev server session
		const devSessionName = `dev_${execution.worktree_id}`;
		const wsProtocol = $page.url.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${wsProtocol}//${$page.url.host}/ws?session=${devSessionName}`;
		devWs = new WebSocket(wsUrl);

		devWs.onopen = function() {
			console.log('WebSocket connected to dev server session:', devSessionName);
		};

		devWs.onmessage = function(event) {
			devTerm.write(event.data);
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

		const handleResize = () => {
			devFitAddon.fit();
		};

		window.addEventListener('resize', handleResize);
	}

	async function runDevServer() {
		if (!execution?.worktree_id) return;
		
		try {
			const response = await fetch(`/api/worktrees/${execution.worktree_id}/dev-server`, {
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
		if (!execution?.worktree_id) return;
		
		try {
			const response = await fetch(`/api/worktrees/${execution.worktree_id}/dev-server`, {
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
		const confirmed = confirm(`Are you sure you want to delete this task execution? This will:\n\n• Kill all associated tmux sessions\n• Remove the worktree directory\n• Run teardown commands\n• Delete all related data\n\nThis action cannot be undone.`);
		
		if (!confirmed) return;
		
		try {
			isDeleting = true;
			const response = await fetch(`/api/task-executions/${executionId}`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				// Navigate back to tasks list
				goto('/tasks');
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
</script>

<svelte:head>
	<title>Task Execution {executionId} - Remote-Code</title>
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
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-400"></div>
					<div>
						<h1 class="text-2xl font-bold text-purple-400">Loading Task Execution...</h1>
						<p class="text-gray-300">Fetching execution details</p>
					</div>
				</div>
			{:else if error}
				<div class="flex items-center gap-4">
					<svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
					<div>
						<h1 class="text-2xl font-bold text-red-400">Execution Not Found</h1>
						<p class="text-gray-300">{error}</p>
					</div>
				</div>
			{:else if execution}
				<div class="flex items-center justify-between mb-6">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 bg-purple-500 rounded-lg flex items-center justify-center">
							<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
							</svg>
						</div>
						<div>
							<div class="flex items-center gap-3 mb-1">
								<h1 class="text-2xl font-bold text-white">{execution.task_title || `Task ${execution.task_id}`}</h1>
								<span class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-semibold border {getStatusColor(execution.status)}">
									{#if execution.status?.toLowerCase() === 'waiting'}
										<svg class="w-4 h-4 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
											<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
										</svg>
									{:else if execution.status?.toLowerCase() === 'running'}
										<div class="w-2 h-2 bg-current rounded-full animate-pulse"></div>
									{:else if execution.status?.toLowerCase() === 'completed'}
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
											<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
										</svg>
									{:else if execution.status?.toLowerCase() === 'failed'}
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
											<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
										</svg>
									{/if}
									{execution.status.toUpperCase()}
								</span>
							</div>
							<p class="text-gray-300">Task Execution #{execution.id}</p>
							{#if execution.status?.toLowerCase() === 'waiting'}
								<div class="flex items-center text-yellow-400 text-sm mt-2 bg-yellow-500/10 border border-yellow-500/20 rounded px-3 py-2">
									<svg class="w-4 h-4 mr-2 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
										<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
									</svg>
									<strong>Agent may be waiting for user input.</strong> Check the terminal below.
								</div>
							{/if}
						</div>
					</div>
					
					<div class="flex gap-2">
						<button 
							on:click={() => runDevServer()}
							disabled={devServerRunning}
							class="bg-green-500 hover:bg-green-600 disabled:bg-green-700 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
							</svg>
							{devServerRunning ? 'Dev Server Running' : 'Run Dev Server'}
						</button>
							
							{#if devServerRunning}
								<button 
									on:click={() => stopDevServer()}
									class="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 9l6 6m0-6l-6 6"/>
									</svg>
									Stop Dev Server
								</button>
							{/if}
						
						<!-- Delete Task Execution Button -->
						<button 
							on:click={deleteTaskExecution}
							disabled={isDeleting}
							class="bg-red-600 hover:bg-red-700 disabled:bg-red-800 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
						>
							{#if isDeleting}
								<div class="animate-spin rounded-full h-4 w-4 border-b border-white"></div>
							{:else}
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
								</svg>
							{/if}
							{isDeleting ? 'Deleting...' : 'Delete'}
						</button>
					</div>
				</div>

				<!-- Execution Info Grid -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
					<!-- Execution Details -->
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6">
						<h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
							</svg>
							Execution Details
						</h3>
						<div class="space-y-3 text-sm">
							<div class="flex justify-between">
								<span class="text-gray-400">Status:</span>
								<span class="font-mono {execution.status === 'completed' ? 'text-green-400' : execution.status === 'running' ? 'text-blue-400' : execution.status === 'failed' ? 'text-red-400' : 'text-gray-400'}">
									{execution.status}
								</span>
							</div>
							<div class="flex justify-between">
								<span class="text-gray-400">Agent:</span>
								<span class="font-mono text-orange-400">{execution.agent_name || `Agent ${execution.agent_id}`}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-gray-400">Created:</span>
								<span class="text-gray-300">{formatDate(execution.created_at)}</span>
							</div>
							{#if execution.updated_at}
								<div class="flex justify-between">
									<span class="text-gray-400">Updated:</span>
									<span class="text-gray-300">{formatDate(execution.updated_at)}</span>
								</div>
							{/if}
						</div>
					</div>

					<!-- Worktree Info -->
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6">
						<h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
							</svg>
							Worktree
						</h3>
						<div class="space-y-3 text-sm">
							<div>
								<span class="text-gray-400 block">Path:</span>
								<span class="font-mono text-yellow-400 text-xs">{execution.worktree_path || 'N/A'}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-gray-400">Base Directory:</span>
								<span class="font-mono text-blue-400">{execution.base_directory_id || 'N/A'}</span>
							</div>
						</div>
					</div>

					<!-- Git Status -->
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6">
						<h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
							</svg>
							Git Status
						</h3>
						{#if gitStatus}
							<div class="space-y-3 text-sm">
								<div class="flex justify-between">
									<span class="text-gray-400">Branch:</span>
									<span class="font-mono text-green-400">{gitStatus.currentBranch}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-400">Status:</span>
									<span class="font-mono {gitStatus.isDirty ? 'text-yellow-400' : 'text-green-400'}">
										{gitStatus.isDirty ? 'Modified' : 'Clean'}
									</span>
								</div>
								{#if gitStatus.ahead || gitStatus.behind}
									<div class="flex justify-between">
										<span class="text-gray-400">Sync:</span>
										<span class="font-mono text-blue-400">
											{#if gitStatus.ahead}+{gitStatus.ahead}{/if}
											{#if gitStatus.behind} -{gitStatus.behind}{/if}
										</span>
									</div>
								{/if}
								<div class="flex justify-between">
									<span class="text-gray-400">Changes:</span>
									<span class="font-mono text-gray-300">
										{gitStatus.staged}S {gitStatus.unstaged}M {gitStatus.untracked}?
									</span>
								</div>
								{#if gitStatus.mergeConflicts}
									<div class="flex items-center gap-2 text-red-400">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
										</svg>
										<span class="text-sm">Merge conflicts</span>
									</div>
								{/if}
							</div>
						{:else}
							<div class="text-gray-400 text-sm">
								<p>Git status not available</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- Re-send Task Button (always show if running) -->
				{#if execution.status === 'running'}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-6">
						<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
							<h3 class="text-lg font-semibold text-white">Task Controls</h3>
							<button 
								on:click={resendTaskToSession}
								disabled={isResendingTask}
								class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-700 disabled:cursor-not-allowed text-white px-3 py-2 rounded text-sm transition-colors flex items-center justify-center gap-2 w-full sm:w-auto"
							>
								{#if isResendingTask}
									<div class="animate-spin rounded-full h-3 w-3 border-b border-white"></div>
								{:else}
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
									</svg>
								{/if}
								{isResendingTask ? 'Sending...' : 'Re-Send Task'}
							</button>
						</div>
					</div>
				{/if}

				<!-- Task Description -->
				{#if execution.task_description}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
						<h3 class="text-lg font-semibold text-white mb-3">Task Description</h3>
						<p class="text-gray-300">{execution.task_description}</p>
					</div>
				{/if}
			{/if}
		</div>

    <!-- Git Panel -->
    {#if !loading && !error && execution && gitStatus}
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-6">
            <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                    <span class="text-sm text-gray-300">Branch</span>
                    <span class="px-2 py-1 rounded text-xs font-semibold bg-blue-500/20 text-blue-300 border border-blue-500">{gitStatus.currentBranch}</span>
                    {#if gitStatus.ahead || gitStatus.behind}
                        <span class="text-xs text-gray-400">{gitStatus.ahead ? `↑ ${gitStatus.ahead}` : ''} {gitStatus.behind ? `↓ ${gitStatus.behind}` : ''}</span>
                    {/if}
                </div>
                <div class="flex items-center gap-2">
                    <button class="text-xs px-3 py-1 rounded bg-gray-700 hover:bg-gray-600" on:click={loadGitStatus}>Refresh</button>
                    <button
                        class="text-xs px-3 py-1 rounded bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed"
                        disabled={pushing || !gitStatus.upstream || gitStatus.ahead === 0}
                        on:click={pushChanges}
                        title={(!gitStatus.upstream ? 'No upstream configured' : (gitStatus.ahead === 0 ? 'Nothing to push' : 'Push to upstream'))}
                    >
                        {pushing ? 'Pushing…' : 'Push'}
                    </button>
                </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                    <h4 class="text-sm text-gray-300 mb-2 flex items-center gap-2">
                        <span class="inline-flex w-2.5 h-2.5 rounded-full bg-green-400"></span>
                        <span>Staged</span>
                    </h4>
                    {#if gitStatus.stagedFiles?.length}
                        {#each gitStatus.stagedFiles as f}
                            <div class="flex items-center justify-between text-sm bg-gray-900 border border-gray-700 rounded px-2 py-1 mb-1">
                                <span class="truncate">{f.path}</span>
                                <div class="flex items-center gap-2">
                                    <button class="text-xs px-2 py-0.5 bg-gray-700 hover:bg-gray-600 rounded" on:click={() => viewDiff(f.path, true)}>Diff</button>
                                    <button class="text-xs px-2 py-0.5 bg-gray-700 hover:bg-gray-600 rounded" on:click={() => unstageFile(f.path)}>Unstage</button>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <div class="text-xs text-gray-500">No staged changes</div>
                    {/if}
                </div>
                <div>
                    <h4 class="text-sm text-gray-300 mb-2 flex items-center gap-2">
                        <span class="inline-flex w-2.5 h-2.5 rounded-full bg-orange-400"></span>
                        <span>Changes</span>
                    </h4>
                    {#if gitStatus.unstagedFiles?.length}
                        {#each gitStatus.unstagedFiles as f}
                            <div class="flex items-center justify-between text-sm bg-gray-900 border border-gray-700 rounded px-2 py-1 mb-1">
                                <span class="truncate">{f.path}</span>
                                <div class="flex items-center gap-2">
                                    <button class="text-xs px-2 py-0.5 bg-gray-700 hover:bg-gray-600 rounded" on:click={() => viewDiff(f.path, false)}>Diff</button>
                                    <button class="text-xs px-2 py-0.5 bg-gray-700 hover:bg-gray-600 rounded" on:click={() => stageFile(f.path)}>Stage</button>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <div class="text-xs text-gray-500">No unstaged changes</div>
                    {/if}
                </div>
                <div>
                    <h4 class="text-sm text-gray-300 mb-2 flex items-center gap-2">
                        <span class="inline-flex w-2.5 h-2.5 rounded-full bg-sky-400"></span>
                        <span>Untracked</span>
                    </h4>
                    {#if gitStatus.untrackedFiles?.length}
                        {#each gitStatus.untrackedFiles as f}
                            <div class="flex items-center justify-between text-sm bg-gray-900 border border-gray-700 rounded px-2 py-1 mb-1">
                                <span class="truncate">{f.path}</span>
                                <div class="flex items-center gap-2">
                                    <button class="text-xs px-2 py-0.5 bg-gray-700 hover:bg-gray-600 rounded" on:click={() => stageFile(f.path)}>Add</button>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <div class="text-xs text-gray-500">No untracked files</div>
                    {/if}
                </div>
            </div>
            <div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="flex gap-2">
                    <input class="flex-1 bg-gray-700 border border-gray-600 rounded px-3 py-2 text-sm" placeholder="Commit message" bind:value={commitMsg} />
                    <button class="px-3 py-2 bg-green-600 hover:bg-green-700 rounded text-sm disabled:bg-gray-600" disabled={committing || !commitMsg.trim()} on:click={commitChanges}>Commit</button>
                </div>
                <div class="flex gap-2 items-center">
                    <button class="px-3 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm disabled:bg-gray-600 disabled:cursor-not-allowed"
                            disabled={merging || !mergeReady}
                            title={mergeReady ? 'Merge the worktree branch into main' : (mergeReadyReason || 'Not ready to merge')}
                            on:click={mergeBranch}>
                        {merging ? 'Merging…' : 'Merge into main'}
                    </button>
                    {#if !mergeReady && mergeReadyReason}
                        <span class="text-xs text-gray-400">{mergeReadyReason}</span>
                        {#if mergeReadyReason.includes('Non fast-forward')}
                            <button class="ml-2 text-xs px-2 py-1 rounded bg-gray-700 hover:bg-gray-600 disabled:bg-gray-700 disabled:opacity-60"
                                    disabled={updatingFromMain}
                                    on:click={() => updateFromMain('merge')}>
                                {updatingFromMain ? 'Updating…' : 'Update from main'}
                            </button>
                        {/if}
                    {/if}
                </div>
            </div>
        </div>
    {/if}

    <!-- Diff Modal -->
    {#if diffOpen}
        <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div class="bg-gray-900 border border-gray-700 rounded-lg w-[90vw] max-w-4xl max-h-[80vh] overflow-hidden">
                <div class="flex items-center justify-between p-3 border-b border-gray-700">
                    <div class="text-sm text-gray-300 truncate">{diffTitle}</div>
                    <button class="text-gray-400 hover:text-white" on:click={() => diffOpen = false}>✕</button>
                </div>
                <div class="p-3 overflow-auto">
                    <pre class="text-xs leading-snug whitespace-pre-wrap">{diffText || 'No diff'}</pre>
                </div>
            </div>
        </div>
    {/if}

    <!-- Terminal -->
		{#if !loading && !error && execution}
			<div class="bg-black rounded-lg border border-gray-700 shadow-xl">
				<div class="flex items-center justify-between p-4 border-b border-gray-700">
					<div class="flex items-center gap-3">
						<div class="flex gap-1">
							<div class="w-3 h-3 rounded-full bg-red-500"></div>
							<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
							<div class="w-3 h-3 rounded-full bg-green-500"></div>
						</div>
						<span class="text-gray-400 text-sm font-mono">tmux attach -t task_{execution.task_id}_agent_{execution.agent_id}</span>
					</div>
					<div class="flex items-center gap-2 text-xs text-gray-500">
						<div class="w-2 h-2 rounded-full bg-green-500"></div>
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
			</div>
			
			<div class="mt-4 text-sm text-gray-400 text-center">
				<p>Terminal connected to task execution session</p>
			</div>

			<!-- Input Section for Sending Text to Session -->
			{#if execution.status === 'running'}
				<div class="mt-6 bg-gray-800 rounded-lg border border-gray-700 p-4">
					<h4 class="text-sm font-semibold text-white mb-3 flex items-center gap-2">
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
							class="flex-1 bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed text-sm"
						/>
						<button
							on:click={sendInputToSession}
							disabled={!inputText.trim() || isSendingInput}
							class="bg-purple-500 hover:bg-purple-600 disabled:bg-gray-600 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg transition-colors flex items-center justify-center gap-2 min-w-[80px] sm:min-w-[80px] w-full sm:w-auto"
						>
							{#if isSendingInput}
								<div class="animate-spin rounded-full h-4 w-4 border-b border-white"></div>
							{:else}
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
								</svg>
							{/if}
							{isSendingInput ? '' : 'Send'}
						</button>
					</div>
					<p class="text-xs text-gray-400 mt-2">
						Send text input directly to the agent session. Press Enter or click Send.
					</p>
				</div>
			{/if}

			<!-- Dev Server Terminal -->
			{#if showDevTerminal}
				<div class="mt-6 bg-black rounded-lg border border-green-700 shadow-xl">
					<div class="flex items-center justify-between p-4 border-b border-green-700">
						<div class="flex items-center gap-3">
							<div class="flex gap-1">
								<div class="w-3 h-3 rounded-full bg-red-500"></div>
								<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
								<div class="w-3 h-3 rounded-full bg-green-500"></div>
							</div>
							<span class="text-green-400 text-sm font-mono">tmux attach -t dev_{execution.worktree_id}</span>
							<span class="px-2 py-1 rounded text-xs font-semibold bg-green-500/20 text-green-400 border border-green-500">DEV SERVER</span>
						</div>
						<div class="flex items-center gap-2 text-xs text-green-500">
							<div class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
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
				</div>
				
				<div class="mt-4 text-sm text-green-400 text-center">
					<p>Dev server terminal - showing startup logs and output</p>
				</div>
			{/if}
		{:else if error}
			<div class="bg-gray-800 rounded-lg border border-red-600 p-8 text-center">
				<svg class="w-16 h-16 text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<h3 class="text-xl font-semibold text-red-400 mb-2">Task Execution Not Available</h3>
				<p class="text-gray-400 mb-6">{error}</p>
				<div class="flex gap-3 justify-center">
					<a 
						href="/tasks"
						class="bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg transition-colors"
					>
						View All Executions
					</a>
					<a 
						href="/"
						class="bg-purple-500 hover:bg-purple-600 text-white px-6 py-3 rounded-lg transition-colors"
					>
						Go to Dashboard
					</a>
				</div>
			</div>
		{/if}
	</div>
</div>
