<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	
	let project = null;
	let loading = true;
	let error = null;
	let viewMode = 'kanban'; // 'kanban' or 'list'
	let showCreateTaskForm = false;
	let showCreateDirectoryForm = false;
	let newTask = { title: '', description: '', status: 'todo', baseDirectoryId: '' };
	let newDirectory = { 
		path: '', 
		gitInitialized: false,
		worktreeSetupCommands: '',
		worktreeTeardownCommands: '',
		devServerSetupCommands: '',
		devServerTeardownCommands: ''
	};
	let selectedTask = null;
	let showTaskModal = false;
	
	// Kanban columns
	const columns = [
		{ id: 'todo', title: 'To Do', color: 'bg-gray-500' },
		{ id: 'in_progress', title: 'In Progress', color: 'bg-blue-500' },
		{ id: 'done', title: 'Done', color: 'bg-green-500' }
	];
	
	// Function to set default base directory when opening form
	function setDefaultBaseDirectory() {
		if (project && project.baseDirectories && project.baseDirectories.length === 1) {
			newTask.baseDirectoryId = project.baseDirectories[0].base_directory_id;
		}
	}
	
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
		if (!newTask.baseDirectoryId) {
			alert('Please select a base directory for the task.');
			return;
		}
		
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
			
			// Reset form with default base directory if only one exists
			const defaultBaseDir = project.baseDirectories && project.baseDirectories.length === 1 
				? project.baseDirectories[0].base_directory_id 
				: '';
			newTask = { title: '', description: '', status: 'todo', baseDirectoryId: defaultBaseDir };
			showCreateTaskForm = false;
		} catch (error) {
			console.error('Failed to create task:', error);
			alert('Failed to create task. Please try again.');
		}
	}
	
	async function createDirectory() {
		console.log('createDirectory called', { newDirectory, projectId });
		
		if (!newDirectory.path.trim()) {
			console.log('Path is empty, returning');
			return;
		}
		
		try {
			const url = `/api/projects/${projectId}/base-directories`;
			const body = JSON.stringify(newDirectory);
			console.log('Making request to:', url, 'with body:', body);
			
			const response = await fetch(url, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: body
			});
			
			console.log('Response status:', response.status);
			
			if (!response.ok) {
				const errorText = await response.text();
				console.error('HTTP error response:', errorText);
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			const createdDirectory = await response.json();
			console.log('Created directory:', createdDirectory);
			
			project.baseDirectories = [...(project.baseDirectories || []), createdDirectory];
			newDirectory = { 
				path: '', 
				gitInitialized: false,
				worktreeSetupCommands: '',
				worktreeTeardownCommands: '',
				devServerSetupCommands: '',
				devServerTeardownCommands: ''
			};
			showCreateDirectoryForm = false;
			console.log('Directory creation completed successfully');
		} catch (error) {
			console.error('Failed to create directory:', error);
			alert('Failed to create directory. Please try again.');
		}
	}
	
	async function deleteDirectory(directoryId) {
		if (!confirm('Are you sure you want to delete this directory? This action cannot be undone.')) {
			return;
		}
		
		try {
			const url = `/api/projects/${projectId}/base-directories/${directoryId}`;
			console.log('Deleting directory:', url);
			
			const response = await fetch(url, {
				method: 'DELETE'
			});
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			// Remove the directory from the local state
			project.baseDirectories = project.baseDirectories.filter(dir => dir.base_directory_id !== directoryId);
			console.log('Directory deleted successfully');
		} catch (error) {
			console.error('Failed to delete directory:', error);
			alert('Failed to delete directory. Please try again.');
		}
	}
	
	function selectTask(task) {
		selectedTask = task;
		showTaskModal = true;
	}
	
	function closeTaskModal() {
		selectedTask = null;
		showTaskModal = false;
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
							on:click={() => showCreateDirectoryForm = true}
							class="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
							</svg>
							Add Directory
						</button>
						
						<button 
							on:click={() => {
								showCreateTaskForm = true;
								setDefaultBaseDirectory();
							}}
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
						<label for="task-base-directory" class="block text-sm font-medium text-gray-300 mb-2">
							Base Directory
						</label>
						<select 
							id="task-base-directory"
							bind:value={newTask.baseDirectoryId}
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
							required
						>
							<option value="">Select a base directory...</option>
							{#if project && project.baseDirectories}
								{#each project.baseDirectories as directory}
									<option value={directory.base_directory_id}>{directory.path}</option>
								{/each}
							{/if}
						</select>
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

		<!-- Create Directory Form -->
		{#if showCreateDirectoryForm}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-xl font-semibold text-white mb-4">Add Base Directory</h3>
				<form on:submit|preventDefault={createDirectory} class="space-y-4">
					<div>
						<label for="directory-path" class="block text-sm font-medium text-gray-300 mb-2">
							Directory Path
						</label>
						<input 
							id="directory-path"
							type="text" 
							bind:value={newDirectory.path}
							placeholder="/path/to/project/directory"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-green-400"
							required
						/>
					</div>
					
					<div class="flex items-center">
						<input 
							id="git-initialized"
							type="checkbox" 
							bind:checked={newDirectory.gitInitialized}
							class="w-4 h-4 text-green-500 bg-gray-700 border-gray-600 rounded focus:ring-green-400 focus:ring-2"
						/>
						<label for="git-initialized" class="ml-2 text-sm text-gray-300">
							Git initialized
						</label>
					</div>
					
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="worktree-setup" class="block text-sm font-medium text-gray-300 mb-2">
								Worktree Setup Commands
							</label>
							<textarea 
								id="worktree-setup"
								bind:value={newDirectory.worktreeSetupCommands}
								placeholder="npm install"
								rows="3"
								class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-green-400"
							></textarea>
						</div>
						
						<div>
							<label for="worktree-teardown" class="block text-sm font-medium text-gray-300 mb-2">
								Worktree Teardown Commands
							</label>
							<textarea 
								id="worktree-teardown"
								bind:value={newDirectory.worktreeTeardownCommands}
								placeholder="npm run clean"
								rows="3"
								class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-green-400"
							></textarea>
						</div>
						
						<div>
							<label for="dev-server-setup" class="block text-sm font-medium text-gray-300 mb-2">
								Dev Server Setup Commands
							</label>
							<textarea 
								id="dev-server-setup"
								bind:value={newDirectory.devServerSetupCommands}
								placeholder="npm run dev"
								rows="3"
								class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-green-400"
							></textarea>
						</div>
						
						<div>
							<label for="dev-server-teardown" class="block text-sm font-medium text-gray-300 mb-2">
								Dev Server Teardown Commands
							</label>
							<textarea 
								id="dev-server-teardown"
								bind:value={newDirectory.devServerTeardownCommands}
								placeholder="pkill -f 'npm run dev'"
								rows="3"
								class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-green-400"
							></textarea>
						</div>
					</div>
					
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Add Directory
						</button>
						<button 
							type="button"
							on:click={() => showCreateDirectoryForm = false}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Base Directories Section -->
		{#if project && (project.baseDirectories || []).length > 0}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
					</svg>
					Base Directories
				</h3>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each project.baseDirectories as directory}
						<div class="bg-gray-700 rounded-lg p-4 border border-gray-600">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center gap-2">
									<svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
									</svg>
									<span class="font-mono text-sm text-green-300">{directory.path}</span>
									{#if directory.gitInitialized}
										<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-900 text-blue-300">
											Git
										</span>
									{/if}
								</div>
								<button 
									on:click={() => deleteDirectory(directory.base_directory_id)}
									class="text-red-400 hover:text-red-300 hover:bg-red-900/20 rounded p-1 transition-colors"
									title="Delete directory"
									aria-label="Delete directory"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
									</svg>
								</button>
							</div>
							
							{#if directory.worktreeSetupCommands || directory.devServerSetupCommands}
								<div class="text-xs text-gray-400 space-y-1">
									{#if directory.worktreeSetupCommands}
										<div><strong>Setup:</strong> {directory.worktreeSetupCommands}</div>
									{/if}
									{#if directory.devServerSetupCommands}
										<div><strong>Dev Server:</strong> {directory.devServerSetupCommands}</div>
									{/if}
								</div>
							{/if}
						</div>
					{/each}
				</div>
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
									<div 
										class="bg-gray-700 rounded-lg p-3 border border-gray-600 hover:border-gray-500 hover:bg-gray-650 cursor-pointer transition-colors"
										on:click={() => selectTask(task)}
									>
										<h4 class="font-medium text-white mb-1">{task.title}</h4>
										{#if task.description}
											<p class="text-sm text-gray-300 mb-2">{task.description}</p>
										{/if}
										<div class="text-xs text-gray-400 mt-2">
											📁 {task.baseDirectory.path}
										</div>
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
								<div 
									class="p-4 hover:bg-gray-750 cursor-pointer transition-colors"
									on:click={() => selectTask(task)}
								>
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<h4 class="font-medium text-white mb-1">{task.title}</h4>
											{#if task.description}
												<p class="text-sm text-gray-300 mb-2">{task.description}</p>
											{/if}
											<div class="text-xs text-gray-400 mt-1">
												📁 {task.baseDirectory.path}
											</div>
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

<!-- Task Detail Modal -->
{#if showTaskModal && selectedTask}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={closeTaskModal}>
		<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 max-w-2xl w-full mx-4" on:click|stopPropagation>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-xl font-semibold text-white">Task Details</h2>
				<button 
					on:click={closeTaskModal}
					class="text-gray-400 hover:text-white transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>
			
			<div class="space-y-4">
				<div>
					<h3 class="text-lg font-medium text-white mb-2">{selectedTask.title}</h3>
					<div class="flex items-center gap-2 mb-3">
						{#each columns as column}
							{#if column.id === selectedTask.status}
								<span class="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-gray-700 text-gray-300">
									<div class="w-2 h-2 rounded-full {column.color}"></div>
									{column.title}
								</span>
							{/if}
						{/each}
					</div>
				</div>
				
				{#if selectedTask.description}
					<div>
						<h4 class="text-sm font-medium text-gray-300 mb-2">Description</h4>
						<p class="text-gray-400">{selectedTask.description}</p>
					</div>
				{/if}
				
				<div>
					<h4 class="text-sm font-medium text-gray-300 mb-2">Base Directory</h4>
					<div class="bg-gray-700 rounded-lg p-3 border border-gray-600">
						<div class="flex items-center gap-2 mb-2">
							<svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
							</svg>
							<span class="font-mono text-sm text-green-300">{selectedTask.baseDirectory.path}</span>
							{#if selectedTask.baseDirectory.git_initialized}
								<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-900 text-blue-300">
									Git
								</span>
							{/if}
						</div>
						
						{#if selectedTask.baseDirectory.worktree_setup_commands || selectedTask.baseDirectory.dev_server_setup_commands}
							<div class="text-xs text-gray-400 space-y-1">
								{#if selectedTask.baseDirectory.worktree_setup_commands}
									<div><strong>Setup:</strong> {selectedTask.baseDirectory.worktree_setup_commands}</div>
								{/if}
								{#if selectedTask.baseDirectory.dev_server_setup_commands}
									<div><strong>Dev Server:</strong> {selectedTask.baseDirectory.dev_server_setup_commands}</div>
								{/if}
							</div>
						{/if}
					</div>
				</div>
				
				<div class="pt-4 border-t border-gray-700">
					<h4 class="text-sm font-medium text-gray-300 mb-3">Execute Task</h4>
					<div class="bg-gray-700 rounded-lg p-4 text-center">
						<p class="text-gray-400 mb-3">Select agents to execute this task</p>
						<button 
							class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
							disabled
						>
							Start Execution (Coming Soon)
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}