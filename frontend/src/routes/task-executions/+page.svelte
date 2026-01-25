<script>
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import KanbanCard from '$lib/components/ui/KanbanCard.svelte';
	
	let taskExecutions = [];
	let loading = true;
	let deletingExecutions = new Set();
	let viewMode = 'kanban'; // 'kanban' or 'list'

	onMount(async () => {
		await loadTaskExecutions();
		
		// Periodically refresh task executions to update waiting status
		const refreshInterval = setInterval(async () => {
			await loadTaskExecutions();
		}, 10000); // Refresh every 10 seconds
		
		// Cleanup interval on component destroy
		return () => {
			clearInterval(refreshInterval);
		};
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
			taskExecutions = [];
			loading = false;
		}
	}

	async function deleteTaskExecution(executionId) {
		if (deletingExecutions.has(executionId)) return;
		
		const confirmed = confirm(`Are you sure you want to delete this task execution? This will:\n\n• Kill all associated tmux sessions\n• Run teardown commands\n• Delete all related data\n\nThis action cannot be undone.`);
		
		if (!confirmed) return;
		
		try {
			deletingExecutions = new Set([...deletingExecutions, executionId]);
			
			const response = await fetch(`/api/task-executions/${executionId}`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				await loadTaskExecutions();
			} else {
				const errorData = await response.text();
				alert(`Failed to delete task execution: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to delete task execution:', err);
			alert('Failed to delete task execution');
		} finally {
			deletingExecutions = new Set([...deletingExecutions].filter(id => id !== executionId));
		}
	}

	async function updateTaskStatus(executionId, newStatus) {
		try {
			const response = await fetch(`/api/task-executions/${executionId}`, {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ status: newStatus })
			});
			
			if (response.ok) {
				await loadTaskExecutions();
			} else {
				alert('Failed to update task status');
			}
		} catch (error) {
			console.error('Failed to update task status:', error);
			alert('Failed to update task status');
		}
	}

	// Group tasks by status for Kanban view
	$: kanbanColumns = {
		pending: taskExecutions.filter(t => t.status?.toLowerCase() === 'pending'),
		running: taskExecutions.filter(t => t.status?.toLowerCase() === 'running'),
		waiting: taskExecutions.filter(t => t.status?.toLowerCase() === 'waiting'),
		completed: taskExecutions.filter(t => t.status?.toLowerCase() === 'completed'),
		failed: taskExecutions.filter(t => t.status?.toLowerCase() === 'failed')
	};

	$: totalTasks = taskExecutions.length;
	$: completedTasks = kanbanColumns.completed.length;
	$: runningTasks = kanbanColumns.running.length;
	$: waitingTasks = kanbanColumns.waiting.length;

	function getColumnTitle(status) {
		const counts = {
			pending: kanbanColumns.pending.length,
			running: kanbanColumns.running.length,
			waiting: kanbanColumns.waiting.length,
			completed: kanbanColumns.completed.length,
			failed: kanbanColumns.failed.length
		};

		const titles = {
			pending: 'Pending',
			running: 'Running',
			waiting: 'Waiting',
			completed: 'Completed',
			failed: 'Failed'
		};

		return `${titles[status]} ${counts[status]}`;
	}

	function getColumnColor(status) {
		const colors = {
			pending: 'border-slate-300',
			running: 'border-vanna-magenta',
			waiting: 'border-vanna-orange',
			completed: 'border-vanna-teal',
			failed: 'border-vanna-orange'
		};
		return colors[status] || 'border-slate-300';
	}
</script>

<svelte:head>
	<title>Task Executions - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-vanna-navy font-serif">Task Executions</h1>
			<p class="mt-2 text-slate-500">Track and manage task executions and workflows</p>
		</div>
		<div class="flex items-center space-x-3">
			<!-- View Toggle -->
			<div class="flex items-center bg-vanna-cream/50 rounded-lg p-1">
				<button
					onclick={() => viewMode = 'kanban'}
					class="px-3 py-1 text-sm font-medium rounded-md transition-colors {viewMode === 'kanban' ? 'bg-white text-vanna-navy shadow-sm' : 'text-slate-500 hover:text-vanna-navy'}"
				>
					<svg class="w-4 h-4 mr-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"/>
					</svg>
					Kanban
				</button>
				<button
					onclick={() => viewMode = 'list'}
					class="px-3 py-1 text-sm font-medium rounded-md transition-colors {viewMode === 'list' ? 'bg-white text-vanna-navy shadow-sm' : 'text-slate-500 hover:text-vanna-navy'}"
				>
					<svg class="w-4 h-4 mr-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/>
					</svg>
					List
				</button>
			</div>
			<Button href="/projects" variant="primary">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
				</svg>
				Execute Task
			</Button>
		</div>
	</div>

	<!-- Stats Cards -->
	<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
		<Card class="p-4">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-vanna-navy/10 rounded-lg flex items-center justify-center">
						<svg class="w-4 h-4 text-vanna-navy" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
						</svg>
					</div>
				</div>
				<div class="ml-3">
					<p class="text-sm font-medium text-slate-500">Total Tasks</p>
					<p class="text-lg font-semibold text-vanna-navy">{totalTasks}</p>
				</div>
			</div>
		</Card>
		<Card class="p-4">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-vanna-teal/10 rounded-lg flex items-center justify-center">
						<svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
						</svg>
					</div>
				</div>
				<div class="ml-3">
					<p class="text-sm font-medium text-slate-500">Completed</p>
					<p class="text-lg font-semibold text-vanna-navy">{completedTasks}</p>
				</div>
			</div>
		</Card>
		<Card class="p-4">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-vanna-magenta/10 rounded-lg flex items-center justify-center">
						<div class="w-2 h-2 bg-vanna-magenta rounded-full animate-pulse"></div>
					</div>
				</div>
				<div class="ml-3">
					<p class="text-sm font-medium text-slate-500">Running</p>
					<p class="text-lg font-semibold text-vanna-navy">{runningTasks}</p>
				</div>
			</div>
		</Card>
		<Card class="p-4">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-vanna-orange/10 rounded-lg flex items-center justify-center">
						<svg class="w-4 h-4 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
						</svg>
					</div>
				</div>
				<div class="ml-3">
					<p class="text-sm font-medium text-slate-500">Waiting</p>
					<p class="text-lg font-semibold text-vanna-navy">{waitingTasks}</p>
				</div>
			</div>
		</Card>
	</div>

	<!-- Task Content -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
		</div>
	{:else if taskExecutions.length === 0}
		<Card class="text-center py-12">
			<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
				<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
				</svg>
			</div>
			<h3 class="text-xl font-semibold text-vanna-navy mb-2">No Task Executions Yet</h3>
			<p class="text-slate-500 mb-4">Execute tasks from the Projects page to see executions here</p>
			<Button href="/projects" variant="primary">
				Go to Projects
			</Button>
		</Card>
	{:else if viewMode === 'kanban'}
		<!-- Kanban View -->
		<div class="grid grid-cols-1 lg:grid-cols-5 gap-6 min-h-96">
			{#each Object.entries(kanbanColumns) as [status, tasks]}
				<div class="flex flex-col">
					<div class="flex items-center justify-between mb-4 pb-2 border-b-2 {getColumnColor(status)}">
						<h3 class="font-semibold text-vanna-navy">{getColumnTitle(status)}</h3>
						{#if status === 'waiting' && tasks.length > 0}
							<Badge variant="warning" size="sm">Needs Attention</Badge>
						{/if}
					</div>
					<div class="flex-1 space-y-3 min-h-32">
						{#each tasks as task}
							<KanbanCard 
								{task} 
								onDelete={deleteTaskExecution}
								onStatusChange={updateTaskStatus}
							/>
						{/each}
						{#if tasks.length === 0}
							<div class="flex items-center justify-center h-32 border-2 border-dashed border-slate-300 rounded-lg">
								<p class="text-sm text-slate-500">No {status} tasks</p>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- List View -->
		<Card>
			<div class="overflow-hidden">
				<table class="min-w-full divide-y divide-slate-200">
					<thead class="bg-vanna-cream/30">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Task</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Agent</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Created</th>
							<th class="relative px-6 py-3"><span class="sr-only">Actions</span></th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-slate-200">
						{#each taskExecutions as execution}
							<tr class="table-row">
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm font-medium text-vanna-navy">
										{execution.task_title || `Task ${execution.task_id}`}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<Badge variant={execution.status?.toLowerCase() === 'completed' ? 'success' : execution.status?.toLowerCase() === 'running' ? 'primary' : execution.status?.toLowerCase() === 'waiting' ? 'warning' : execution.status?.toLowerCase() === 'failed' ? 'danger' : 'secondary'} size="sm">
										{execution.status || 'Unknown'}
									</Badge>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
									{execution.agent_name || `Agent ${execution.agent_id}`}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
									{execution.created_at ? new Date(execution.created_at).toLocaleDateString() : 'Recently'}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
									<Button href="/task-executions/{execution.id}" variant="ghost" size="sm">
										{execution.status?.toLowerCase() === 'waiting' ? 'Check Session' : 'View Details'}
									</Button>
									<Button onclick={() => deleteTaskExecution(execution.id)} variant="danger" size="sm" disabled={deletingExecutions.has(execution.id)}>
										{#if deletingExecutions.has(execution.id)}
											<div class="animate-spin rounded-full h-3 w-3 border-b border-white mr-1"></div>
										{/if}
										Delete
									</Button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	{/if}
</div>