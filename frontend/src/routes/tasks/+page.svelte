<script>
	import { onMount } from 'svelte';
	
	let tasks = [];
	let loading = true;
	let showCreateForm = false;
	let newTask = { title: '', description: '' };

	onMount(async () => {
		await loadTasks();
	});

	async function loadTasks() {
		try {
			// Mock data for now
			tasks = [
				{
					id: 1,
					title: "Setup development environment",
					description: "Configure tmux sessions and base directories",
					status: "in_progress",
					worktree: {
						path: "/home/user/projects/web-app",
						baseDirectoryId: "web-app"
					}
				},
				{
					id: 2,
					title: "Implement user authentication",
					description: "Add login and registration functionality",
					status: "pending",
					worktree: null
				}
			];
			loading = false;
		} catch (error) {
			console.error('Failed to load tasks:', error);
			loading = false;
		}
	}

	async function createTask() {
		if (!newTask.title.trim()) return;
		
		try {
			const task = {
				id: Date.now(),
				title: newTask.title,
				description: newTask.description,
				status: "pending",
				worktree: null
			};
			
			tasks = [...tasks, task];
			newTask = { title: '', description: '' };
			showCreateForm = false;
		} catch (error) {
			console.error('Failed to create task:', error);
		}
	}

	function getStatusColor(status) {
		switch (status) {
			case 'completed': return 'bg-green-500';
			case 'in_progress': return 'bg-yellow-500';
			case 'pending': return 'bg-gray-500';
			default: return 'bg-gray-500';
		}
	}

	function getStatusText(status) {
		switch (status) {
			case 'completed': return 'Completed';
			case 'in_progress': return 'In Progress';
			case 'pending': return 'Pending';
			default: return 'Unknown';
		}
	}
</script>

<svelte:head>
	<title>Tasks - Remote-Code</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Header -->
		<div class="mb-6">
			<div class="flex items-center gap-4 mb-4">
				<a href="/" class="flex items-center gap-2 text-gray-400 hover:text-white transition-colors">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
					</svg>
					<span>Back to Dashboard</span>
				</a>
			</div>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 bg-purple-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
						</svg>
					</div>
					<div>
						<h1 class="text-3xl font-bold text-purple-400 mb-1">Tasks</h1>
						<p class="text-gray-300">Track and manage development tasks and workflows</p>
					</div>
				</div>
				<button 
					on:click={() => showCreateForm = true}
					class="bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					New Task
				</button>
			</div>
		</div>

		<!-- Create Task Form -->
		{#if showCreateForm}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-xl font-semibold text-white mb-4">Create New Task</h3>
				<form on:submit|preventDefault={createTask} class="space-y-4">
					<div>
						<label for="task-title" class="block text-sm font-medium text-gray-300 mb-2">
							Task Title
						</label>
						<input 
							id="task-title"
							type="text" 
							bind:value={newTask.title}
							placeholder="Enter task title"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-400"
							required
						/>
					</div>
					<div>
						<label for="task-description" class="block text-sm font-medium text-gray-300 mb-2">
							Description
						</label>
						<textarea 
							id="task-description"
							bind:value={newTask.description}
							placeholder="Enter task description"
							rows="3"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-400"
						></textarea>
					</div>
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Create Task
						</button>
						<button 
							type="button"
							on:click={() => showCreateForm = false}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Tasks List -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-400"></div>
			</div>
		{:else if tasks.length === 0}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-gray-300 mb-2">No Tasks Yet</h3>
				<p class="text-gray-400 mb-4">Create your first task to get started</p>
				<button 
					on:click={() => showCreateForm = true}
					class="bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded-lg transition-colors"
				>
					Create Task
				</button>
			</div>
		{:else}
			<div class="space-y-4">
				{#each tasks as task}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 hover:border-purple-400 transition-colors">
						<div class="flex items-start justify-between mb-4">
							<div class="flex-1">
								<div class="flex items-center gap-3 mb-2">
									<h3 class="text-lg font-semibold text-white">{task.title}</h3>
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white {getStatusColor(task.status)}">
										{getStatusText(task.status)}
									</span>
								</div>
								{#if task.description}
									<p class="text-gray-400 mb-3">{task.description}</p>
								{/if}
								{#if task.worktree}
									<div class="flex items-center text-sm text-gray-400">
										<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
										</svg>
										{task.worktree.path}
									</div>
								{/if}
							</div>
							<div class="flex gap-2 ml-4">
								<button class="bg-purple-500 hover:bg-purple-600 text-white px-3 py-1 rounded text-sm transition-colors">
									Edit
								</button>
								<button class="bg-gray-600 hover:bg-gray-700 text-white px-3 py-1 rounded text-sm transition-colors">
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