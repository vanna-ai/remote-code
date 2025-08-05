<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	
	let project = null;
	let loading = true;
	let error = null;
	let viewMode = 'kanban'; // 'kanban' or 'list'
	let showCreateTaskForm = false;
	let newTask = { title: '', description: '', status: 'todo' };
	
	// Kanban columns
	const columns = [
		{ id: 'todo', title: 'To Do', color: 'bg-gray-500' },
		{ id: 'in_progress', title: 'In Progress', color: 'bg-blue-500' },
		{ id: 'done', title: 'Done', color: 'bg-green-500' }
	];
	
	$: projectId = $page.params.id;
	
	onMount(async () => {
		await loadProject();
	});
	
	async function loadProject() {
		try {
			loading = true;
			error = null;
			const response = await fetch(`/api/projects/${projectId}`);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			project = await response.json();
			
			// Ensure tasks have a status field for Kanban
			if (project.tasks) {
				project.tasks = project.tasks.map(task => ({
					...task,
					status: task.status || 'todo'
				}));
			}
			
			loading = false;
		} catch (err) {
			console.error('Failed to load project:', err);
			error = err.message;
			loading = false;
		}
	}
	
	function getTasksByStatus(status) {
		if (!project || !project.tasks) return [];
		return project.tasks.filter(task => task.status === status);
	}
	
	async function createTask() {
		if (!newTask.title.trim()) return;
		
		try {
			const response = await fetch(`/api/projects/${projectId}/tasks`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(newTask)
			});
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			const createdTask = await response.json();
			project.tasks = [...(project.tasks || []), createdTask];
			newTask = { title: '', description: '', status: 'todo' };
			showCreateTaskForm = false;
		} catch (error) {
			console.error('Failed to create task:', error);
			alert('Failed to create task. Please try again.');
		}
	}
</script>

<svelte:head>
	<title>{project?.name || 'Project'} - Remote-Code</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Header -->
		<div class="mb-6">
			<div class="flex items-center gap-4 mb-4">
				<a href="/projects" class="flex items-center gap-2 text-gray-400 hover:text-white transition-colors">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
					</svg>
					<span>Back to Projects</span>
				</a>
			</div>
			
			{#if loading}
				<div class="flex items-center gap-4">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-400"></div>
					<span class="text-gray-300">Loading project...</span>
				</div>
			{:else if error}
				<div class="bg-red-900/50 border border-red-500 rounded-lg p-4">
					<p class="text-red-300">Error: {error}</p>
				</div>
			{:else if project}
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
							<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
							</svg>
						</div>
						<div>
							<h1 class="text-3xl font-bold text-blue-400 mb-1">{project.name}</h1>
							<p class="text-gray-300">{(project.tasks || []).length} tasks • {(project.baseDirectories || []).length} directories</p>
						</div>
					</div>
					
					<div class="flex items-center gap-3">
						<!-- View Toggle -->
						<div class="flex bg-gray-800 rounded-lg p-1">
							<button 
								class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'kanban' ? 'bg-blue-500 text-white' : 'text-gray-300 hover:text-white'}"
								on:click={() => viewMode = 'kanban'}
							>
								Kanban
							</button>
							<button 
								class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'list' ? 'bg-blue-500 text-white' : 'text-gray-300 hover:text-white'}"
								on:click={() => viewMode = 'list'}
							>
								List
							</button>
						</div>
						
						<button 
							on:click={() => showCreateTaskForm = true}
							class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
							</svg>
							New Task
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Create Task Form -->
		{#if showCreateTaskForm}
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
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
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
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
						></textarea>
					</div>
					<div>
						<label for="task-status" class="block text-sm font-medium text-gray-300 mb-2">
							Initial Status
						</label>
						<select 
							id="task-status"
							bind:value={newTask.status}
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
						>
							{#each columns as column}
								<option value={column.id}>{column.title}</option>
							{/each}
						</select>
					</div>
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Create Task
						</button>
						<button 
							type="button"
							on:click={() => showCreateTaskForm = false}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Content -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-400"></div>
			</div>
		{:else if error}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-red-700 rounded-lg flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-red-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-gray-300 mb-2">Failed to Load Project</h3>
				<p class="text-gray-400 mb-4">Unable to load project details</p>
			</div>
		{:else if project}
			{#if viewMode === 'kanban'}
				<!-- Kanban View -->
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each columns as column}
						<div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
							<div class="flex items-center gap-2 mb-4">
								<div class="w-3 h-3 rounded-full {column.color}"></div>
								<h3 class="font-semibold text-white">{column.title}</h3>
								<span class="text-sm text-gray-400">({getTasksByStatus(column.id).length})</span>
							</div>
							
							<div class="space-y-3">
								{#each getTasksByStatus(column.id) as task}
									<div class="bg-gray-700 rounded-lg p-3 border border-gray-600 hover:border-gray-500 transition-colors">
										<h4 class="font-medium text-white mb-1">{task.title}</h4>
										{#if task.description}
											<p class="text-sm text-gray-300 mb-2">{task.description}</p>
										{/if}
									</div>
								{/each}
								
								{#if getTasksByStatus(column.id).length === 0}
									<div class="text-center py-6 text-gray-500">
										<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
										</svg>
										<p class="text-sm">No tasks</p>
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<!-- List View -->
				<div class="bg-gray-800 rounded-lg border border-gray-700">
					<div class="p-4 border-b border-gray-700">
						<h3 class="text-lg font-semibold text-white">All Tasks</h3>
					</div>
					
					{#if (project.tasks || []).length === 0}
						<div class="text-center py-12">
							<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
								<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
								</svg>
							</div>
							<h3 class="text-xl font-semibold text-gray-300 mb-2">No Tasks Yet</h3>
							<p class="text-gray-400 mb-4">Create your first task to get started</p>
						</div>
					{:else}
						<div class="divide-y divide-gray-700">
							{#each project.tasks || [] as task}
								<div class="p-4 hover:bg-gray-750 transition-colors">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<h4 class="font-medium text-white mb-1">{task.title}</h4>
											{#if task.description}
												<p class="text-sm text-gray-300 mb-2">{task.description}</p>
											{/if}
										</div>
										<div class="flex items-center gap-2 ml-4">
											{#each columns as column}
												{#if column.id === task.status}
													<span class="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-gray-700 text-gray-300">
														<div class="w-2 h-2 rounded-full {column.color}"></div>
														{column.title}
													</span>
												{/if}
											{/each}
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>