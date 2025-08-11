<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import FormField from '$lib/components/ui/FormField.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	
	$: breadcrumbSegments = [
		{ label: "Remote-Code", href: "/", icon: "banner" },
		{ label: "Projects", href: "/projects" },
		{ label: project?.name || "Loading...", href: `/projects/${$page.params.id}` }
	];
	
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
	let editingDirectoryId = '';
	let editDirectory = { 
		path: '', 
		gitInitialized: false,
		worktreeSetupCommands: '',
		worktreeTeardownCommands: '',
		devServerSetupCommands: '',
		devServerTeardownCommands: ''
	};
	let selectedTask = null;
	let showTaskModal = false;
	let availableAgents = [];
	let selectedAgents = [];
	let executingTask = false;
	let showEditTaskModal = false;
	let editingTask = null;
	let editTaskForm = { title: '', description: '', status: '', baseDirectoryId: '' };
	let taskExecutions = new Map(); // Map of taskId to array of executions
	let deletingTasks = new Set();
	let updatingTasks = new Set();
	let deletingProject = false;
	let refreshing = false;
	
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
	
	// Reactive statements for debugging
	$: if (project && project.tasks) {
		console.log('=== PROJECT TASKS DEBUG ===');
		console.log('Total tasks:', project.tasks.length);
		console.log('Task statuses:', project.tasks.map(t => ({ id: t.id, title: t.title, status: t.status })));
		console.log('Unique statuses in tasks:', [...new Set(project.tasks.map(t => t.status))]);
	}
	
	$: todoTasks = (project && !loading) ? getTasksByStatus('todo') : [];
	$: inProgressTasks = (project && !loading) ? getTasksByStatus('in_progress') : [];  
	$: doneTasks = (project && !loading) ? getTasksByStatus('done') : [];
	
	$: {
		console.log('=== KANBAN FILTER DEBUG ===');
		console.log('Todo tasks:', todoTasks.length, todoTasks);
		console.log('In Progress tasks:', inProgressTasks.length, inProgressTasks);
		console.log('Done tasks:', doneTasks.length, doneTasks);
		console.log('Columns config:', columns);
	}
	
	onMount(async () => {
		await loadProject();
		await loadTaskExecutions();
		
		// Refresh task executions every 5 seconds
		const interval = setInterval(loadTaskExecutions, 5000);
		return () => clearInterval(interval);
	});
	
	async function loadProject() {
		try {
			// Only show loading spinner on initial load, not on refreshes
			if (!project) {
				loading = true;
			} else {
				refreshing = true;
			}
			error = null;
			
			const response = await fetch(`/api/projects/${projectId}`);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			project = await response.json();
			
			console.log('=== RAW PROJECT DATA ===');
			console.log('Raw project from API:', project);
			console.log('Raw tasks:', project.tasks);
			
			// Ensure tasks have a status field for Kanban
			if (project.tasks) {
				project.tasks = project.tasks.map(task => ({
					...task,
					status: task.status || 'todo'
				}));
				
				console.log('=== PROCESSED TASKS ===');
				console.log('Tasks after processing:', project.tasks);
			}
			
			// Trigger reactivity
			project = { ...project };
			
			loading = false;
			refreshing = false;
		} catch (err) {
			console.error('Failed to load project:', err);
			error = err.message;
			loading = false;
			refreshing = false;
		}
	}
	
	function getTasksByStatus(status) {
		console.log('getTasksByStatus called with status:', status);
		console.log('Project exists:', !!project);
		console.log('Project.tasks exists:', !!(project && project.tasks));
		console.log('Project.tasks length:', project?.tasks?.length || 0);
		
		if (!project || !project.tasks) {
			console.log('Returning empty array - no project or tasks');
			return [];
		}
		
		console.log('All tasks with statuses:', project.tasks.map(t => ({ id: t.id, title: t.title, status: t.status })));
		
		const filteredTasks = project.tasks.filter(task => {
			console.log(`Task ${task.id} (${task.title}) has status '${task.status}', looking for '${status}', match: ${task.status === status}`);
			return task.status === status;
		});
		
		console.log(`Found ${filteredTasks.length} tasks with status '${status}':`, filteredTasks.map(t => ({ id: t.id, title: t.title })));
		return filteredTasks;
	}
	
	async function loadTaskExecutions() {
		if (!project || !project.tasks) return;
		
		try {
			// Load executions for all tasks in this project
			for (const task of project.tasks) {
				const response = await fetch(`/api/task-executions?task_id=${task.id}`);
				if (response.ok) {
					const executions = await response.json();
					taskExecutions.set(task.id, executions);
				}
			}
			// Trigger reactivity
			taskExecutions = new Map(taskExecutions);
		} catch (error) {
			console.error('Failed to load task executions:', error);
		}
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
			
			// Reload the entire project to ensure consistency
			await loadProject();
			
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
			
			// Reload the entire project to ensure consistency
			await loadProject();
			
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
			
			// Reload the entire project to ensure consistency
			await loadProject();
			console.log('Directory deleted successfully');
		} catch (error) {
			console.error('Failed to delete directory:', error);
			alert('Failed to delete directory. Please try again.');
		}
	}

	function startEditDirectory(dir) {
		editingDirectoryId = dir.base_directory_id;
		editDirectory = {
			path: dir.path,
			gitInitialized: (dir.gitInitialized ?? dir.git_initialized) || false,
			worktreeSetupCommands: dir.worktreeSetupCommands ?? dir.worktree_setup_commands ?? '',
			worktreeTeardownCommands: dir.worktreeTeardownCommands ?? dir.worktree_teardown_commands ?? '',
			devServerSetupCommands: dir.devServerSetupCommands ?? dir.dev_server_setup_commands ?? '',
			devServerTeardownCommands: dir.devServerTeardownCommands ?? dir.dev_server_teardown_commands ?? ''
		};
	}

	function cancelEditDirectory() {
		editingDirectoryId = '';
		editDirectory = { 
			path: '', 
			gitInitialized: false,
			worktreeSetupCommands: '',
			worktreeTeardownCommands: '',
			devServerSetupCommands: '',
			devServerTeardownCommands: ''
		};
	}

	async function saveEditDirectory() {
		if (!editingDirectoryId) return;
		try {
			const url = `/api/projects/${projectId}/base-directories/${editingDirectoryId}`;
			const response = await fetch(url, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(editDirectory)
			});
			if (!response.ok) {
				const text = await response.text();
				console.error('Update failed:', text);
				throw new Error(`HTTP ${response.status}`);
			}
			await loadProject();
			cancelEditDirectory();
		} catch (err) {
			console.error('Failed to update directory:', err);
			alert('Failed to update directory.');
		}
	}
	
	async function selectTask(task) {
		selectedTask = task;
		selectedAgents = [];
		showTaskModal = true;
		
		// Load available agents when task is selected
		await loadAvailableAgents();
	}
	
	function closeTaskModal() {
		selectedTask = null;
		showTaskModal = false;
		selectedAgents = [];
		availableAgents = [];
	}
	
	function openEditTaskModal(task) {
		editingTask = task;
		editTaskForm = {
			title: task.title,
			description: task.description,
			status: task.status,
			baseDirectoryId: task.baseDirectory?.base_directory_id || ''
		};
		showEditTaskModal = true;
	}
	
	function closeEditTaskModal() {
		editingTask = null;
		showEditTaskModal = false;
		editTaskForm = { title: '', description: '', status: '', baseDirectoryId: '' };
	}
	
	async function loadAvailableAgents() {
		try {
			const response = await fetch('/api/agents');
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			availableAgents = await response.json();
		} catch (error) {
			console.error('Failed to load agents:', error);
			availableAgents = [];
		}
	}
	
	function toggleAgentSelection(agent) {
		const isSelected = selectedAgents.some(a => a.id === agent.id);
		if (isSelected) {
			selectedAgents = selectedAgents.filter(a => a.id !== agent.id);
		} else {
			selectedAgents = [...selectedAgents, agent];
		}
	}
	
	async function startTaskExecution() {
		if (selectedAgents.length === 0) {
			alert('Please select at least one agent to execute the task.');
			return;
		}
		
		executingTask = true;
		
		try {
			for (const agent of selectedAgents) {
				const response = await fetch('/api/task-executions', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({
						task_id: selectedTask.id,
						agent_id: agent.id
					})
				});
				
				if (!response.ok) {
					throw new Error(`Failed to start execution with ${agent.name}: ${response.status}`);
				}
			}
			
			alert(`Task execution started with ${selectedAgents.length} agent(s)!`);
			// Refresh project data and task executions
			await loadProject();
			await loadTaskExecutions();
			closeTaskModal();
		} catch (error) {
			console.error('Failed to start task execution:', error);
			alert('Failed to start task execution. Please try again.');
		} finally {
			executingTask = false;
		}
	}

	async function deleteTask(task) {
		if (deletingTasks.has(task.id)) return;
		
		// Show confirmation dialog
		const confirmed = confirm(`Are you sure you want to delete "${task.title}"? This will:\n\n• Delete all task executions\n• Clean up all associated resources\n• Remove the task permanently\n\nThis action cannot be undone.`);
		
		if (!confirmed) return;
		
		try {
			// Add to deleting set to show loading state
			deletingTasks = new Set([...deletingTasks, task.id]);
			
			const response = await fetch(`/api/tasks/${task.id}`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				// Reload the entire project to ensure consistency
				await loadProject();
				// Reload task executions
				await loadTaskExecutions();
			} else {
				const errorData = await response.text();
				alert(`Failed to delete task: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to delete task:', err);
			alert('Failed to delete task');
		} finally {
			// Remove from deleting set
			deletingTasks = new Set([...deletingTasks].filter(id => id !== task.id));
		}
	}

	async function updateTaskStatus(task, newStatus) {
		if (updatingTasks.has(task.id)) return;
		
		try {
			// Add to updating set to show loading state
			updatingTasks = new Set([...updatingTasks, task.id]);
			
			const response = await fetch(`/api/tasks/${task.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					title: task.title,
					description: task.description,
					status: newStatus
				})
			});
			
			if (response.ok) {
				const updatedTask = await response.json();
				console.log('Task status updated:', updatedTask);
				
				// Immediately update the task in local state for instant UI feedback
				const taskIndex = project.tasks.findIndex(t => t.id === task.id);
				if (taskIndex >= 0) {
					project.tasks[taskIndex] = { ...project.tasks[taskIndex], status: updatedTask.status };
					// Force Svelte reactivity by reassigning the array and project
					project = { 
						...project, 
						tasks: [...project.tasks]
					};
				}
			} else {
				const errorData = await response.text();
				alert(`Failed to update task status: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to update task status:', err);
			alert('Failed to update task status');
		} finally {
			// Remove from updating set
			updatingTasks = new Set([...updatingTasks].filter(id => id !== task.id));
		}
	}

	async function saveTaskEdit() {
		if (!editingTask || !editTaskForm.title.trim()) return;
		
		try {
			// Add to updating set to show loading state
			updatingTasks = new Set([...updatingTasks, editingTask.id]);
			
			const response = await fetch(`/api/tasks/${editingTask.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					title: editTaskForm.title,
					description: editTaskForm.description,
					status: editTaskForm.status
				})
			});
			
			if (response.ok) {
				const updatedTask = await response.json();
				console.log('Task updated:', updatedTask);
				
				// Reload the entire project to ensure consistency
				await loadProject();
				
				closeEditTaskModal();
			} else {
				const errorData = await response.text();
				alert(`Failed to update task: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to update task:', err);
			alert('Failed to update task');
		} finally {
			// Remove from updating set
			updatingTasks = new Set([...updatingTasks].filter(id => id !== editingTask.id));
		}
	}

	async function deleteProject() {
		if (deletingProject) return;
		
		// Show confirmation dialog
		const confirmed = confirm(`Are you sure you want to delete the "${project.name}" project? This will:\n\n• Delete all tasks and their executions\n• Clean up all associated resources\n• Remove all base directories\n• Delete the project permanently\n\nThis action cannot be undone.`);
		
		if (!confirmed) return;
		
		try {
			deletingProject = true;
			
			const response = await fetch(`/api/projects/${projectId}`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				// Navigate back to projects list
				window.location.href = '/';
			} else {
				const errorData = await response.text();
				alert(`Failed to delete project: ${errorData}`);
			}
		} catch (err) {
			console.error('Failed to delete project:', err);
			alert('Failed to delete project');
		} finally {
			deletingProject = false;
		}
	}
</script>

<svelte:head>
	<title>{project?.name || 'Project'} - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Breadcrumb -->
	<Breadcrumb segments={breadcrumbSegments} />
	
	<!-- Header -->
	{#if loading}
		<div class="flex items-center justify-center min-h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
		</div>
	{:else if error}
		<Card class="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10">
			<div class="flex">
				<svg class="w-5 h-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error loading project</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">
						<p>{error}</p>
					</div>
				</div>
			</div>
		</Card>
	{:else if project}
		<Card>
			<PageHeader 
				title={project.name} 
				subtitle="{(project.tasks || []).length} tasks • {(project.baseDirectories || []).length} directories"
			>
				<!-- View Toggle -->
				<div class="flex bg-gray-200 dark:bg-gray-700 rounded-lg p-1">
					<button 
						class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'kanban' ? 'bg-blue-500 text-white' : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'}"
						on:click={() => viewMode = 'kanban'}
					>
						Kanban
					</button>
					<button 
						class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'list' ? 'bg-blue-500 text-white' : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'}"
						on:click={() => viewMode = 'list'}
					>
						List
					</button>
				</div>
				
				<Button 
					variant="success"
					onclick={() => showCreateDirectoryForm = true}
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
					</svg>
					Add Directory
				</Button>
				
				<Button 
					variant="primary"
					onclick={() => {
						showCreateTaskForm = true;
						setDefaultBaseDirectory();
					}}
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					New Task
				</Button>
				
				<Button 
					variant="danger"
					onclick={deleteProject}
					disabled={deletingProject}
					loading={deletingProject}
				>
					{#if !deletingProject}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
						</svg>
					{/if}
					{deletingProject ? 'Deleting...' : 'Delete Project'}
				</Button>
			</PageHeader>
		</Card>
	{/if}

		<!-- Create Task Modal -->
		<Modal open={showCreateTaskForm} title="Create New Task" onClose={() => showCreateTaskForm = false}>
		<form on:submit|preventDefault={createTask} class="space-y-6">
		<FormField label="Task Title" id="task-title" required>
		 <Input 
		 id="task-title"
		type="text" 
		bind:value={newTask.title}
		placeholder="Enter task title"
		required
		/>
		</FormField>
		
		<FormField label="Description" id="task-description">
		<Textarea 
		id="task-description"
		bind:value={newTask.description}
		 placeholder="Enter task description"
		 rows={3}
		/>
		</FormField>
		
		<FormField label="Base Directory" id="task-base-directory" required>
		<Select 
		id="task-base-directory"
		bind:value={newTask.baseDirectoryId}
		placeholder="Select a base directory..."
		required
		>
		 {#if project && project.baseDirectories}
		  {#each project.baseDirectories as directory}
		  <option value={directory.base_directory_id}>{directory.path}</option>
		{/each}
		{/if}
		</Select>
		</FormField>
		
		<FormField label="Initial Status" id="task-status">
		<Select 
		id="task-status"
		bind:value={newTask.status}
		>
		{#each columns as column}
		<option value={column.id}>{column.title}</option>
		{/each}
		</Select>
		</FormField>
		
		<div class="flex gap-3 pt-4">
		<Button type="submit" variant="primary">
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		 <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
		</svg>
		Create Task
		</Button>
		<Button 
		type="button"
		variant="secondary"
		onclick={() => showCreateTaskForm = false}
		>
		Cancel
		</Button>
		</div>
		</form>
		</Modal>

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
								<div class="flex items-center gap-2">
									<button 
										on:click={() => startEditDirectory(directory)}
										class="text-blue-400 hover:text-blue-300 hover:bg-blue-900/20 rounded p-1 transition-colors"
										title="Edit directory"
										aria-label="Edit directory"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
										</svg>
									</button>
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

							{#if editingDirectoryId === directory.base_directory_id}
								<div class="mt-3 space-y-3 border-t border-gray-600 pt-3">
									<div>
										<label class="block text-sm text-gray-300 mb-1">Path</label>
										<input class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white" bind:value={editDirectory.path} />
									</div>
									<div class="flex items-center gap-2">
										<input id={`git-${directory.base_directory_id}`} type="checkbox" bind:checked={editDirectory.gitInitialized} class="accent-blue-500" />
										<label for={`git-${directory.base_directory_id}`} class="text-sm text-gray-300">Git initialized</label>
									</div>
									<div>
										<label class="block text-sm text-gray-300 mb-1">Worktree Setup Commands</label>
										<textarea class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white" rows="2" bind:value={editDirectory.worktreeSetupCommands}></textarea>
									</div>
									<div>
										<label class="block text-sm text-gray-300 mb-1">Worktree Teardown Commands</label>
										<textarea class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white" rows="2" bind:value={editDirectory.worktreeTeardownCommands}></textarea>
									</div>
									<div>
										<label class="block text-sm text-gray-300 mb-1">Dev Server Setup Commands</label>
										<textarea class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white" rows="2" bind:value={editDirectory.devServerSetupCommands}></textarea>
									</div>
									<div>
										<label class="block text-sm text-gray-300 mb-1">Dev Server Teardown Commands</label>
										<textarea class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white" rows="2" bind:value={editDirectory.devServerTeardownCommands}></textarea>
									</div>
									<div class="flex gap-2">
										<button class="bg-blue-500 hover:bg-blue-600 text-white px-3 py-1.5 rounded" on:click={saveEditDirectory}>Save</button>
										<button class="bg-gray-600 hover:bg-gray-700 text-white px-3 py-1.5 rounded" on:click={cancelEditDirectory}>Cancel</button>
									</div>
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
								<span class="text-sm text-gray-400">({column.id === 'todo' ? todoTasks.length : column.id === 'in_progress' ? inProgressTasks.length : doneTasks.length})</span>
							</div>
							
							<div class="space-y-3">
								{#each column.id === 'todo' ? todoTasks : column.id === 'in_progress' ? inProgressTasks : doneTasks as task}
									<div 
										class="bg-gray-700 rounded-lg p-3 border border-gray-600 hover:border-gray-500 hover:bg-gray-650 transition-colors group"
									>
										<div class="flex items-start justify-between mb-1">
											<h4 
												class="font-medium text-white flex-1 cursor-pointer hover:text-blue-300 transition-colors" 
												on:click={() => selectTask(task)}
											>
												{task.title}
											</h4>
											<div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">
												<!-- Edit Button -->
												<button 
													on:click|stopPropagation={() => openEditTaskModal(task)}
													class="text-blue-400 hover:text-blue-300 hover:bg-blue-900/20 rounded p-1 transition-colors"
													title="Edit task"
													aria-label="Edit task"
												>
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
													</svg>
												</button>
												
												<!-- Status Change Dropdown -->
												<div class="relative">
													<select 
														value={task.status}
														on:change={(e) => updateTaskStatus(task, e.target.value)}
														disabled={updatingTasks.has(task.id)}
														class="bg-gray-600 border border-gray-500 rounded px-2 py-1 text-xs text-white hover:bg-gray-500 transition-colors disabled:opacity-50"
														on:click|stopPropagation
													>
														{#each columns as col}
															<option value={col.id}>{col.title}</option>
														{/each}
													</select>
												</div>
												
												<!-- Delete Button -->
												<button 
													on:click|stopPropagation={() => deleteTask(task)}
													disabled={deletingTasks.has(task.id)}
													class="text-red-400 hover:text-red-300 hover:bg-red-900/20 rounded p-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
													title="Delete task"
												>
													{#if deletingTasks.has(task.id)}
														<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>
													{:else}
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
														</svg>
													{/if}
												</button>
											</div>
										</div>
										
										<div class="cursor-pointer" on:click={() => selectTask(task)}>
											{#if task.description}
												<p class="text-sm text-gray-300 mb-2">{task.description}</p>
											{/if}
										
										<!-- Task Executions -->
										{#if taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0}
											<div class="mb-2">
												<div class="flex flex-wrap gap-1">
													{#each taskExecutions.get(task.id) as execution}
														<a 
															href="/tasks/{execution.id}"
															class="inline-flex items-center gap-1 bg-purple-600 hover:bg-purple-700 text-purple-100 px-2 py-1 rounded text-xs transition-colors"
															on:click|stopPropagation
														>
															<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
															</svg>
															{execution.agent_name}
														</a>
													{/each}
												</div>
											</div>
										{/if}
										
											<div class="text-xs text-gray-400 mt-2">
												📁 {task.baseDirectory?.path || 'No base directory'}
											</div>
										</div>
									</div>
								{/each}
								
								{#if (column.id === 'todo' ? todoTasks : column.id === 'in_progress' ? inProgressTasks : doneTasks).length === 0}
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
									class="p-4 hover:bg-gray-750 transition-colors group"
								>
									<div class="flex items-start justify-between">
										<div class="flex-1 cursor-pointer" on:click={() => selectTask(task)}>
											<h4 class="font-medium text-white mb-1">{task.title}</h4>
											{#if task.description}
												<p class="text-sm text-gray-300 mb-2">{task.description}</p>
											{/if}
											
											<!-- Task Executions -->
											{#if taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0}
												<div class="mb-2">
													<div class="flex flex-wrap gap-1">
														{#each taskExecutions.get(task.id) as execution}
															<a 
																href="/tasks/{execution.id}"
																class="inline-flex items-center gap-1 bg-purple-600 hover:bg-purple-700 text-purple-100 px-2 py-1 rounded text-xs transition-colors"
																on:click|stopPropagation
															>
																<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
																</svg>
																{execution.agent_name}
															</a>
														{/each}
													</div>
												</div>
											{/if}
											
											<div class="text-xs text-gray-400 mt-1">
												📁 {task.baseDirectory?.path || 'No base directory'}
											</div>
										</div>
										
										<div class="flex items-center gap-3 ml-4">
											<!-- Status Badge -->
											{#each columns as column}
												{#if column.id === task.status}
													<span class="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-gray-700 text-gray-300">
														<div class="w-2 h-2 rounded-full {column.color}"></div>
														{column.title}
													</span>
												{/if}
											{/each}
											
											<!-- Task Management Buttons (hidden by default, shown on hover) -->
											<div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
												<!-- Edit Button -->
												<button 
													on:click|stopPropagation={() => openEditTaskModal(task)}
													class="text-blue-400 hover:text-blue-300 hover:bg-blue-900/20 rounded p-2 transition-colors"
													title="Edit task"
													aria-label="Edit task"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
													</svg>
												</button>
												
												<!-- Status Change Dropdown -->
												<select 
													value={task.status}
													on:change={(e) => updateTaskStatus(task, e.target.value)}
													disabled={updatingTasks.has(task.id)}
													class="bg-gray-600 border border-gray-500 rounded px-2 py-1 text-xs text-white hover:bg-gray-500 transition-colors disabled:opacity-50"
													on:click|stopPropagation
												>
													{#each columns as col}
														<option value={col.id}>{col.title}</option>
													{/each}
												</select>
												
												<!-- Delete Button -->
												<button 
													on:click|stopPropagation={() => deleteTask(task)}
													disabled={deletingTasks.has(task.id)}
													class="text-red-400 hover:text-red-300 hover:bg-red-900/20 rounded p-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
													title="Delete task"
												>
													{#if deletingTasks.has(task.id)}
														<div class="animate-spin rounded-full h-4 w-4 border-b border-current"></div>
													{:else}
														<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
														</svg>
													{/if}
												</button>
											</div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		{/if}

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
							<span class="font-mono text-sm text-green-300">{selectedTask.baseDirectory?.path || 'No path'}</span>
							{#if selectedTask.baseDirectory?.git_initialized}
								<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-900 text-blue-300">
									Git
								</span>
							{/if}
						</div>
						
						{#if selectedTask.baseDirectory?.worktree_setup_commands || selectedTask.baseDirectory?.dev_server_setup_commands}
							<div class="text-xs text-gray-400 space-y-1">
								{#if selectedTask.baseDirectory?.worktree_setup_commands}
									<div><strong>Setup:</strong> {selectedTask.baseDirectory.worktree_setup_commands}</div>
								{/if}
								{#if selectedTask.baseDirectory?.dev_server_setup_commands}
									<div><strong>Dev Server:</strong> {selectedTask.baseDirectory.dev_server_setup_commands}</div>
								{/if}
							</div>
						{/if}
					</div>
				</div>
				
				<!-- Active Executions -->
				{#if taskExecutions.has(selectedTask.id) && taskExecutions.get(selectedTask.id).length > 0}
					<div>
						<h4 class="text-sm font-medium text-gray-300 mb-3">Active Executions</h4>
						<div class="space-y-2">
							{#each taskExecutions.get(selectedTask.id) as execution}
								<div class="bg-gray-700 rounded-lg p-3 border border-yellow-600">
									<div class="flex items-center justify-between">
										<div class="flex items-center gap-2">
											<svg class="w-4 h-4 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
											</svg>
											<span class="text-white font-medium">{execution.agent_name}</span>
											<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-600 text-yellow-100">
												{execution.status}
											</span>
										</div>
										<a 
											href="/tasks/{execution.id}"
											class="bg-purple-500 hover:bg-purple-600 text-white px-3 py-1 rounded text-xs transition-colors"
										>
											View Execution
										</a>
									</div>
									<div class="text-xs text-gray-400 mt-2">
										Started: {new Date(execution.created_at.Time).toLocaleString()}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
				
				<div class="pt-4 border-t border-gray-700">
					<h4 class="text-sm font-medium text-gray-300 mb-3">Execute Task</h4>
					
					{#if availableAgents.length > 0}
						<div class="bg-gray-700 rounded-lg p-4">
							<p class="text-gray-400 mb-3">Select agents to execute this task:</p>
							
							<div class="space-y-2 mb-4 max-h-40 overflow-y-auto">
								{#each availableAgents as agent}
									<label class="flex items-center gap-3 p-2 rounded hover:bg-gray-600 cursor-pointer">
										<input 
											type="checkbox"
											checked={selectedAgents.some(a => a.id === agent.id)}
											on:change={() => toggleAgentSelection(agent)}
											class="w-4 h-4 text-blue-500 bg-gray-600 border-gray-500 rounded focus:ring-blue-400 focus:ring-2"
										/>
										<div class="flex-1">
											<div class="font-medium text-white">{agent.name}</div>
											<div class="text-sm text-gray-400 font-mono">{agent.command} {agent.params}</div>
										</div>
									</label>
								{/each}
							</div>
							
							{#if selectedAgents.length > 0}
								<div class="text-sm text-gray-400 mb-3">
									Selected: {selectedAgents.map(a => a.name).join(', ')}
								</div>
							{/if}
							
							<button 
								on:click={startTaskExecution}
								disabled={selectedAgents.length === 0 || executingTask}
								class="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-600 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg transition-colors flex items-center justify-center gap-2"
							>
								{#if executingTask}
									<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
									Starting Execution...
								{:else}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h1m4 0h1m-6-8h1m4 0h1m-2-4h.01M12 16h.01M12 8h.01M12 12h.01"/>
									</svg>
									Start Execution ({selectedAgents.length})
								{/if}
							</button>
						</div>
					{:else}
						<div class="bg-gray-700 rounded-lg p-4 text-center">
							<p class="text-gray-400 mb-3">No agents configured</p>
							<a 
								href="/agents"
								class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors inline-flex items-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
								</svg>
								Configure Agents
							</a>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Task Edit Modal -->
{#if showEditTaskModal && editingTask}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={closeEditTaskModal}>
		<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 max-w-2xl w-full mx-4" on:click|stopPropagation>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-xl font-semibold text-white">Edit Task</h2>
				<button 
					on:click={closeEditTaskModal}
					class="text-gray-400 hover:text-white transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>
			
			<form on:submit|preventDefault={saveTaskEdit} class="space-y-4">
				<div>
					<label for="edit-task-title" class="block text-sm font-medium text-gray-300 mb-2">
						Task Title
					</label>
					<input 
						id="edit-task-title"
						type="text" 
						bind:value={editTaskForm.title}
						placeholder="Enter task title"
						class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
						required
					/>
				</div>
				
				<div>
					<label for="edit-task-description" class="block text-sm font-medium text-gray-300 mb-2">
						Description
					</label>
					<textarea 
						id="edit-task-description"
						bind:value={editTaskForm.description}
						placeholder="Enter task description"
						rows="4"
						class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
					></textarea>
				</div>
				
				<div>
					<label for="edit-task-status" class="block text-sm font-medium text-gray-300 mb-2">
						Status
					</label>
					<select 
						id="edit-task-status"
						bind:value={editTaskForm.status}
						class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
					>
						{#each columns as column}
							<option value={column.id}>{column.title}</option>
						{/each}
					</select>
				</div>
				
				<div>
					<label for="edit-task-base-directory" class="block text-sm font-medium text-gray-300 mb-2">
						Base Directory
					</label>
					<select 
						id="edit-task-base-directory"
						bind:value={editTaskForm.baseDirectoryId}
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
				
				<div class="flex gap-3 pt-4">
					<button 
						type="submit"
						disabled={updatingTasks.has(editingTask.id)}
						class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
					>
						{#if updatingTasks.has(editingTask.id)}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							Saving...
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
							</svg>
							Save Changes
						{/if}
					</button>
					<button 
						type="button"
						on:click={closeEditTaskModal}
						class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
					>
						Cancel
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

</div>
