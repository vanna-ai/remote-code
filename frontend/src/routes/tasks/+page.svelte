<script>
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	
	const breadcrumbSegments = [
		{ label: "Remote-Code", href: "/", icon: "banner" },
		{ label: "Task Executions", href: "/tasks" }
	];
	
	let taskExecutions = [];
	let loading = true;
	let showExecuteForm = false;
	let tasks = [];
	let agents = [];
	let worktrees = [];
	let newExecution = { taskId: '', agentId: '', worktreeId: '' };
	let deletingExecutions = new Set();

	onMount(async () => {
		await loadTaskExecutions();
		await loadTasks();
		await loadAgents();
		await loadWorktrees();
	});

	async function loadTaskExecutions() {
		try {
			const response = await fetch('/api/task-executions');
			if (!response.ok) {
				throw new Error('Failed to fetch task executions');
			}
			taskExecutions = await response.json();
			loading = false;
		} catch (error) {
			console.error('Failed to load task executions:', error);
			// Fallback to empty array
			taskExecutions = [];
			loading = false;
		}
	}

	async function loadTasks() {
		try {
			const response = await fetch('/api/tasks');
			if (response.ok) {
				tasks = await response.json();
			}
		} catch (error) {
			console.error('Failed to load tasks:', error);
		}
	}

	async function loadAgents() {
		try {
			const response = await fetch('/api/agents');
			if (response.ok) {
				agents = await response.json();
			}
		} catch (error) {
			console.error('Failed to load agents:', error);
		}
	}

	async function loadWorktrees() {
		try {
			const response = await fetch('/api/worktrees');
			if (response.ok) {
				worktrees = await response.json();
			}
		} catch (error) {
			console.error('Failed to load worktrees:', error);
		}
	}

	async function executeTask() {
		if (!newExecution.taskId || !newExecution.agentId || !newExecution.worktreeId) return;
		
		try {
			const response = await fetch('/api/task-executions', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					task_id: parseInt(newExecution.taskId),
					agent_id: parseInt(newExecution.agentId),
					worktree_id: parseInt(newExecution.worktreeId),
					status: 'pending'
				})
			});
			
			if (!response.ok) {
				throw new Error('Failed to create task execution');
			}
			
			await loadTaskExecutions();
			newExecution = { taskId: '', agentId: '', worktreeId: '' };
			showExecuteForm = false;
		} catch (error) {
			console.error('Failed to execute task:', error);
			alert('Failed to execute task: ' + error.message);
		}
	}

	function getStatusColor(status) {
		switch (status) {
			case 'completed': return 'bg-green-500';
			case 'running': return 'bg-blue-500';
			case 'failed': return 'bg-red-500';
			case 'pending': return 'bg-gray-500';
			default: return 'bg-gray-500';
		}
	}

	function getStatusText(status) {
		switch (status) {
			case 'completed': return 'Completed';
			case 'running': return 'Running';
			case 'failed': return 'Failed';
			case 'pending': return 'Pending';
			default: return 'Unknown';
		}
	}

	function getTaskTitle(taskId) {
		const task = tasks.find(t => t.id === taskId);
		return task ? task.title : `Task ${taskId}`;
	}

	function getAgentName(agentId) {
		const agent = agents.find(a => a.id === agentId);
		return agent ? agent.name : `Agent ${agentId}`;
	}

	function getWorktreePath(worktreeId) {
		const worktree = worktrees.find(w => w.id === worktreeId);
		return worktree ? worktree.path : `Worktree ${worktreeId}`;
	}

	async function deleteTaskExecution(executionId) {
		if (deletingExecutions.has(executionId)) return;
		
		// Show confirmation dialog
		const confirmed = confirm(`Are you sure you want to delete this task execution? This will:\n\n• Kill all associated tmux sessions\n• Remove the worktree directory\n• Run teardown commands\n• Delete all related data\n\nThis action cannot be undone.`);
		
		if (!confirmed) return;
		
		try {
			// Add to deleting set to show loading state
			deletingExecutions = new Set([...deletingExecutions, executionId]);
			
			const response = await fetch(`/api/task-executions/${executionId}`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				// Reload task executions to reflect the deletion
				await loadTaskExecutions();
			} else {
				const errorData = await response.text();
				alert(`Failed to delete task execution: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to delete task execution:', err);
			alert('Failed to delete task execution');
		} finally {
			// Remove from deleting set
			deletingExecutions = new Set([...deletingExecutions].filter(id => id !== executionId));
		}
	}
</script>

<svelte:head>
	<title>Task Executions - Remote-Code</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />
		
		<!-- Header -->
		<div class="mb-6">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 bg-purple-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
						</svg>
					</div>
					<div>
						<h1 class="text-3xl font-bold text-purple-400 mb-1">Task Executions</h1>
						<p class="text-gray-300">Track and manage task executions and workflows</p>
					</div>
				</div>
			</div>
		</div>


		<!-- Task Executions List -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-400"></div>
			</div>
		{:else if taskExecutions.length === 0}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-gray-300 mb-2">No Task Executions Yet</h3>
				<p class="text-gray-400 mb-4">Execute tasks from the Projects page to see executions here</p>
				<a 
					href="/projects"
					class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors inline-block"
				>
					Go to Projects
				</a>
			</div>
		{:else}
			<div class="space-y-4">
				{#each taskExecutions as execution}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 hover:border-purple-400 transition-colors">
						<div class="flex items-start justify-between mb-4">
							<div class="flex-1">
								<div class="flex items-center gap-3 mb-2">
									<h3 class="text-lg font-semibold text-white">{execution.task_title || `Task ${execution.task_id}`}</h3>
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white {getStatusColor(execution.status)}">
										{getStatusText(execution.status)}
									</span>
								</div>
								<div class="space-y-2 text-sm text-gray-400">
									<div class="flex items-center">
										<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
										</svg>
										Agent: {execution.agent_name || `Agent ${execution.agent_id}`}
									</div>
									<div class="flex items-center">
										<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
										</svg>
										Worktree: {getWorktreePath(execution.worktree_id)}
									</div>
									{#if execution.created_at}
										<div class="flex items-center">
											<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
											</svg>
											Created: {new Date(execution.created_at).toLocaleString()}
										</div>
									{/if}
								</div>
							</div>
							<div class="flex gap-2 ml-4">
								<a 
									href="/tasks/{execution.id}"
									class="bg-blue-500 hover:bg-blue-600 text-white px-3 py-1 rounded text-sm transition-colors inline-block"
								>
									View Details
								</a>
								{#if execution.status === 'running'}
									<button class="bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded text-sm transition-colors">
										Stop
									</button>
								{/if}
								<button 
									on:click={() => deleteTaskExecution(execution.id)}
									disabled={deletingExecutions.has(execution.id)}
									class="bg-red-600 hover:bg-red-700 disabled:bg-red-800 disabled:cursor-not-allowed text-white px-3 py-1 rounded text-sm transition-colors flex items-center gap-1"
								>
									{#if deletingExecutions.has(execution.id)}
										<div class="animate-spin rounded-full h-3 w-3 border-b border-white"></div>
									{:else}
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
										</svg>
									{/if}
									Delete
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>