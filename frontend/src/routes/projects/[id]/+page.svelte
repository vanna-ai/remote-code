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
		{ label: "", href: "/", icon: "banner" },
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
		setupCommands: '',
		teardownCommands: '',
		devServerSetupCommands: '',
		devServerTeardownCommands: ''
	};
	let editingDirectoryId = '';
	let editDirectory = {
		path: '',
		gitInitialized: false,
		setupCommands: '',
		teardownCommands: '',
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
	let directoryGitStatus = new Map(); // Map of directory id to git status
	
	// Kanban columns
	const columns = [
		{ id: 'todo', title: 'To Do', color: 'bg-slate-400' },
		{ id: 'in_progress', title: 'In Progress', color: 'bg-vanna-magenta' },
		{ id: 'done', title: 'Done', color: 'bg-vanna-teal' }
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
		await loadDirectoryGitStatus();

		// Refresh task executions and git status every 5 seconds
		const interval = setInterval(async () => {
			await loadTaskExecutions();
			await loadDirectoryGitStatus();
		}, 5000);
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

	async function loadDirectoryGitStatus() {
		if (!project || !project.baseDirectories) return;

		try {
			for (const dir of project.baseDirectories) {
				const response = await fetch(`/api/git/status?path=${encodeURIComponent(dir.path)}`);
				if (response.ok) {
					const status = await response.json();
					directoryGitStatus.set(dir.id, status);
				}
			}
			// Trigger reactivity
			directoryGitStatus = new Map(directoryGitStatus);
		} catch (error) {
			console.error('Failed to load git status:', error);
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
				setupCommands: '',
				teardownCommands: '',
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
			setupCommands: dir.setupCommands ?? dir.setup_commands ?? '',
			teardownCommands: dir.teardownCommands ?? dir.teardown_commands ?? '',
			devServerSetupCommands: dir.devServerSetupCommands ?? dir.dev_server_setup_commands ?? '',
			devServerTeardownCommands: dir.devServerTeardownCommands ?? dir.dev_server_teardown_commands ?? ''
		};
	}

	function cancelEditDirectory() {
		editingDirectoryId = '';
		editDirectory = {
			path: '',
			gitInitialized: false,
			setupCommands: '',
			teardownCommands: '',
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
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div>
			</div>
		{:else if error}
			<Card class="border-vanna-orange/30 bg-vanna-orange/5">
				<div class="flex">
					<svg class="w-5 h-5 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
					</svg>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-vanna-orange">Error loading project</h3>
						<div class="mt-2 text-sm text-vanna-orange/80">
							<p>{error}</p>
						</div>
					</div>
				</div>
			</Card>
		{:else if project}
			<div class="border-b border-slate-200 pb-6 mb-8">
				<div class="flex items-center justify-between">
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-4 mb-2">
							<div class="w-12 h-12 bg-vanna-teal rounded-xl flex items-center justify-center">
								<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
								</svg>
							</div>
							<div>
								<h1 class="text-2xl font-bold text-vanna-navy font-serif sm:text-3xl">
									{project.name}
								</h1>
								<p class="mt-1 text-sm text-slate-500">
									{(project.tasks || []).length} tasks • {(project.baseDirectories || []).length} directories
								</p>
							</div>
						</div>
					</div>
					
					<div class="flex items-center gap-3 ml-6">
						<!-- View Toggle -->
						<div class="flex bg-vanna-cream/50 rounded-lg p-1">
							<button
								class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'kanban' ? 'bg-vanna-teal text-white' : 'text-vanna-navy hover:text-vanna-teal'}"
								on:click={() => viewMode = 'kanban'}
							>
								Kanban
							</button>
							<button
								class="px-3 py-1 rounded text-sm transition-colors {viewMode === 'list' ? 'bg-vanna-teal text-white' : 'text-vanna-navy hover:text-vanna-teal'}"
								on:click={() => viewMode = 'list'}
							>
								List
							</button>
						</div>
						
						<Button 
							variant="success"
							onclick={() => showCreateDirectoryForm = true}
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
							</svg>
							{deletingProject ? 'Deleting...' : 'Delete Project'}
						</Button>
					</div>
				</div>
			</div>
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
		<Modal open={showCreateDirectoryForm} title="Add Base Directory" onClose={() => showCreateDirectoryForm = false}>
			<form on:submit|preventDefault={createDirectory} class="space-y-6">
				<FormField label="Directory Path" id="directory-path" required>
					<Input 
						id="directory-path"
						type="text" 
						bind:value={newDirectory.path}
						placeholder="/path/to/project/directory"
						required
					/>
				</FormField>
				
				<div class="flex items-center">
					<input
						id="git-initialized"
						type="checkbox"
						bind:checked={newDirectory.gitInitialized}
						class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"
					/>
					<label for="git-initialized" class="ml-2 text-sm text-vanna-navy">
						Git initialized
					</label>
				</div>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<FormField label="Setup Commands" id="setup-commands">
						<Textarea
							id="setup-commands"
							bind:value={newDirectory.setupCommands}
							placeholder="npm install"
							rows={3}
						/>
					</FormField>

					<FormField label="Teardown Commands" id="teardown-commands">
						<Textarea
							id="teardown-commands"
							bind:value={newDirectory.teardownCommands}
							placeholder="npm run clean"
							rows={3}
						/>
					</FormField>
					
					<FormField label="Dev Server Setup Commands" id="dev-server-setup">
						<Textarea 
							id="dev-server-setup"
							bind:value={newDirectory.devServerSetupCommands}
							placeholder="npm run dev"
							rows={3}
						/>
					</FormField>
					
					<FormField label="Dev Server Teardown Commands" id="dev-server-teardown">
						<Textarea 
							id="dev-server-teardown"
							bind:value={newDirectory.devServerTeardownCommands}
							placeholder="pkill -f 'npm run dev'"
							rows={3}
						/>
					</FormField>
				</div>
				
				<div class="flex gap-3 pt-4">
					<Button type="submit" variant="success">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
						</svg>
						Add Directory
					</Button>
					<Button 
						type="button"
						variant="secondary"
						onclick={() => showCreateDirectoryForm = false}
					>
						Cancel
					</Button>
				</div>
			</form>
		</Modal>

		<!-- Base Directories Section -->
		{#if project && (project.baseDirectories || []).length > 0}
			<Card>
				<h3 class="text-lg font-semibold text-vanna-navy mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
					</svg>
					Base Directories
				</h3>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each project.baseDirectories as directory}
						<div class="bg-vanna-cream/30 rounded-lg p-4 border border-slate-200">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center gap-2">
									<svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
									</svg>
									<span class="font-mono text-sm text-vanna-teal break-all">{directory.path}</span>
									{#if directory.gitInitialized}
										<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal">
											Git
										</span>
									{/if}
								</div>
								<div class="flex items-center gap-2">
									<IconButton
										onclick={() => startEditDirectory(directory)}
										variant="ghost"
										size="sm"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
										</svg>
									</IconButton>
									<IconButton
										onclick={() => deleteDirectory(directory.base_directory_id)}
										variant="ghost"
										size="sm"
										class="text-vanna-orange hover:text-vanna-orange/80"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
										</svg>
									</IconButton>
								</div>
							</div>

							{#if directory.setupCommands || directory.setup_commands || directory.devServerSetupCommands || directory.dev_server_setup_commands}
								<div class="text-xs text-slate-500 space-y-1">
									{#if directory.setupCommands || directory.setup_commands}
										<div><strong>Setup:</strong> {directory.setupCommands || directory.setup_commands}</div>
									{/if}
									{#if directory.devServerSetupCommands || directory.dev_server_setup_commands}
										<div><strong>Dev Server:</strong> {directory.devServerSetupCommands || directory.dev_server_setup_commands}</div>
									{/if}
								</div>
							{/if}

							<!-- Git Status -->
							{#if directoryGitStatus.has(directory.id)}
								{@const gitStatus = directoryGitStatus.get(directory.id)}
								{#if gitStatus.isRepo}
									<div class="mt-3 pt-3 border-t border-slate-200">
										<div class="flex items-center justify-between">
											<div class="flex items-center gap-2 text-sm">
												<svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"/>
												</svg>
												<span class="text-vanna-navy">{gitStatus.currentBranch || 'main'}</span>
												{#if gitStatus.ahead > 0}
													<span class="text-xs bg-vanna-teal/10 text-vanna-teal px-1.5 py-0.5 rounded">
														{gitStatus.ahead} ahead
													</span>
												{/if}
												{#if gitStatus.behind > 0}
													<span class="text-xs bg-vanna-orange/10 text-vanna-orange px-1.5 py-0.5 rounded">
														{gitStatus.behind} behind
													</span>
												{/if}
											</div>
											{#if gitStatus.isDirty}
												<a
													href="/git/{directory.id}"
													class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-vanna-orange bg-vanna-orange/10 rounded hover:bg-vanna-orange/20 transition-colors"
												>
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
													</svg>
													{(gitStatus.stagedFiles?.length || 0) + (gitStatus.unstagedFiles?.length || 0) + (gitStatus.untrackedFiles?.length || 0)} changes
												</a>
											{:else}
												<span class="text-xs text-vanna-teal">Clean</span>
											{/if}
										</div>
									</div>
								{/if}
							{/if}

							{#if editingDirectoryId === directory.base_directory_id}
								<div class="mt-3 space-y-3 border-t border-slate-200 pt-3">
									<FormField label="Path" id={`path-${directory.base_directory_id}`}>
										<Input
											id={`path-${directory.base_directory_id}`}
											bind:value={editDirectory.path}
										/>
									</FormField>
									<div class="flex items-center gap-2">
										<input
											id={`git-${directory.base_directory_id}`}
											type="checkbox"
											bind:checked={editDirectory.gitInitialized}
											class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"
										/>
										<label for={`git-${directory.base_directory_id}`} class="text-sm text-vanna-navy">Git initialized</label>
									</div>
									<FormField label="Setup Commands" id={`setup-${directory.base_directory_id}`}>
										<Textarea 
											id={`setup-${directory.base_directory_id}`}
											bind:value={editDirectory.setupCommands}
											rows={2}
										/>
									</FormField>
									<FormField label="Teardown Commands" id={`teardown-${directory.base_directory_id}`}>
										<Textarea 
											id={`teardown-${directory.base_directory_id}`}
											bind:value={editDirectory.teardownCommands}
											rows={2}
										/>
									</FormField>
									<FormField label="Dev Server Setup Commands" id={`dev-setup-${directory.base_directory_id}`}>
										<Textarea 
											id={`dev-setup-${directory.base_directory_id}`}
											bind:value={editDirectory.devServerSetupCommands}
											rows={2}
										/>
									</FormField>
									<FormField label="Dev Server Teardown Commands" id={`dev-teardown-${directory.base_directory_id}`}>
										<Textarea 
											id={`dev-teardown-${directory.base_directory_id}`}
											bind:value={editDirectory.devServerTeardownCommands}
											rows={2}
										/>
									</FormField>
									<div class="flex gap-2">
										<Button variant="primary" onclick={saveEditDirectory}>Save</Button>
										<Button variant="secondary" onclick={cancelEditDirectory}>Cancel</Button>
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</Card>
		{/if}

		<!-- Content -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
			</div>
		{:else if error}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-vanna-orange/10 rounded-xl flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-vanna-navy mb-2">Failed to Load Project</h3>
				<p class="text-slate-500 mb-4">Unable to load project details</p>
			</div>
		{:else if project}
			{#if viewMode === 'kanban'}
				<!-- Kanban View -->
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each columns as column}
						<Card>
							<div class="flex items-center gap-2 mb-4">
								<div class="w-3 h-3 rounded-full {column.color}"></div>
								<h3 class="font-semibold text-vanna-navy">{column.title}</h3>
								<span class="text-sm text-slate-500">({column.id === 'todo' ? todoTasks.length : column.id === 'in_progress' ? inProgressTasks.length : doneTasks.length})</span>
							</div>
							
							<div class="space-y-3">
								{#each column.id === 'todo' ? todoTasks : column.id === 'in_progress' ? inProgressTasks : doneTasks as task}
									<div
										class="bg-white rounded-lg p-3 border border-slate-200 hover:border-vanna-teal/30 hover:shadow-md transition-all group"
									>
										<div class="flex items-start justify-between mb-1">
											<h4
												class="font-medium text-vanna-navy flex-1 cursor-pointer hover:text-vanna-teal transition-colors"
												on:click={() => selectTask(task)}
											>
												{task.title}
											</h4>
											<div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">
												<!-- Edit Button -->
												<IconButton
													onclick={() => openEditTaskModal(task)}
													variant="ghost"
													size="xs"
													class="text-vanna-teal hover:text-vanna-teal/80"
												>
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
													</svg>
												</IconButton>

												<!-- Status Change Dropdown -->
												<div class="relative">
													<select
														value={task.status}
														on:change={(e) => updateTaskStatus(task, e.target.value)}
														disabled={updatingTasks.has(task.id)}
														class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50"
														on:click|stopPropagation
													>
														{#each columns as col}
															<option value={col.id}>{col.title}</option>
														{/each}
													</select>
												</div>

												<!-- Delete Button -->
												<IconButton
													onclick={() => deleteTask(task)}
													disabled={deletingTasks.has(task.id)}
													variant="ghost"
													size="xs"
													class="text-vanna-orange hover:text-vanna-orange/80"
												>
													{#if deletingTasks.has(task.id)}
														<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>
													{:else}
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
														</svg>
													{/if}
												</IconButton>
											</div>
										</div>

										<div class="cursor-pointer" on:click={() => selectTask(task)}>
											{#if task.description}
												<p class="text-sm text-slate-600 mb-2">{task.description}</p>
											{/if}

										<!-- Task Executions -->
										{#if taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0}
											<div class="mb-2">
												<div class="flex flex-wrap gap-1">
													{#each taskExecutions.get(task.id) as execution}
														<a
															href="/task-executions/{execution.id}"
															class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"
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

											<div class="text-xs text-slate-500 mt-2">
												📁 {task.baseDirectory?.path || 'No base directory'}
											</div>
										</div>
									</div>
								{/each}
								
								{#if (column.id === 'todo' ? todoTasks : column.id === 'in_progress' ? inProgressTasks : doneTasks).length === 0}
									<EmptyState>
										<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
										</svg>
										<p class="text-sm">No tasks</p>
									</EmptyState>
								{/if}
							</div>
						</Card>
					{/each}
				</div>
			{:else}
				<!-- List View -->
				<Card>
					<div class="p-4 border-b border-slate-200">
						<h3 class="text-lg font-semibold text-vanna-navy">All Tasks</h3>
					</div>

					{#if (project.tasks || []).length === 0}
						<EmptyState>
							<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
								<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
								</svg>
							</div>
							<h3 class="text-xl font-semibold text-vanna-navy mb-2">No Tasks Yet</h3>
							<p class="text-slate-500 mb-4">Create your first task to get started</p>
						</EmptyState>
					{:else}
						<div class="divide-y divide-slate-200">
							{#each project.tasks || [] as task}
								<div
									class="p-4 hover:bg-vanna-cream/30 transition-colors group"
								>
									<div class="flex items-start justify-between">
										<div class="flex-1 cursor-pointer" on:click={() => selectTask(task)}>
											<h4 class="font-medium text-vanna-navy mb-1">{task.title}</h4>
											{#if task.description}
												<p class="text-sm text-slate-600 mb-2">{task.description}</p>
											{/if}

											<!-- Task Executions -->
											{#if taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0}
												<div class="mb-2">
													<div class="flex flex-wrap gap-1">
														{#each taskExecutions.get(task.id) as execution}
															<a
																href="/task-executions/{execution.id}"
																class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"
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

											<div class="text-xs text-slate-500 mt-1">
												📁 {task.baseDirectory?.path || 'No base directory'}
											</div>
										</div>

										<div class="flex items-center gap-3 ml-4">
											<!-- Status Badge -->
											{#each columns as column}
												{#if column.id === task.status}
													<StatusBadge status={task.status || 'pending'} size="sm" />
												{/if}
											{/each}

											<!-- Task Management Buttons (hidden by default, shown on hover) -->
											<div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
												<!-- Edit Button -->
												<IconButton
													onclick={() => openEditTaskModal(task)}
													variant="ghost"
													size="sm"
													class="text-vanna-teal hover:text-vanna-teal/80"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
													</svg>
												</IconButton>

												<!-- Status Change Dropdown -->
												<select
													value={task.status}
													on:change={(e) => updateTaskStatus(task, e.target.value)}
													disabled={updatingTasks.has(task.id)}
													class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50"
													on:click|stopPropagation
												>
													{#each columns as col}
														<option value={col.id}>{col.title}</option>
													{/each}
												</select>

												<!-- Delete Button -->
												<IconButton
													onclick={() => deleteTask(task)}
													disabled={deletingTasks.has(task.id)}
													variant="ghost"
													size="sm"
													class="text-vanna-orange hover:text-vanna-orange/80"
												>
													{#if deletingTasks.has(task.id)}
														<div class="animate-spin rounded-full h-4 w-4 border-b border-current"></div>
													{:else}
														<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
														</svg>
													{/if}
												</IconButton>
											</div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</Card>
			{/if}
		{/if}
</div>

<!-- Task Detail Modal -->
<Modal open={showTaskModal && selectedTask} title="Task Details" size="xl" onClose={closeTaskModal}>
	{#if selectedTask}
		<div class="space-y-4">
				<div>
					<h3 class="text-lg font-medium text-vanna-navy mb-2">{selectedTask.title}</h3>
					<div class="flex items-center gap-2 mb-3">
						<StatusBadge status={selectedTask.status || 'pending'} />
					</div>
				</div>

				{#if selectedTask.description}
					<div>
						<h4 class="text-sm font-medium text-vanna-navy mb-2">Description</h4>
						<p class="text-slate-600">{selectedTask.description}</p>
					</div>
				{/if}

				<div>
					<h4 class="text-sm font-medium text-vanna-navy mb-2">Base Directory</h4>
					<div class="bg-vanna-cream/30 rounded-lg p-3 border border-slate-200">
						<div class="flex items-center gap-2 mb-2">
							<svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
							</svg>
							<span class="font-mono text-sm text-vanna-teal">{selectedTask.baseDirectory?.path || 'No path'}</span>
							{#if selectedTask.baseDirectory?.git_initialized}
								<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal">
									Git
								</span>
							{/if}
						</div>

						{#if selectedTask.baseDirectory?.setup_commands || selectedTask.baseDirectory?.dev_server_setup_commands}
							<div class="text-xs text-slate-500 space-y-1">
								{#if selectedTask.baseDirectory?.setup_commands}
									<div><strong>Setup:</strong> {selectedTask.baseDirectory.setup_commands}</div>
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
						<h4 class="text-sm font-medium text-vanna-navy mb-3">Active Executions</h4>
						<div class="space-y-2">
							{#each taskExecutions.get(selectedTask.id) as execution}
								<div class="bg-vanna-orange/5 rounded-lg p-3 border border-vanna-orange/30">
									<div class="flex items-center justify-between">
										<div class="flex items-center gap-2">
											<svg class="w-4 h-4 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
											</svg>
											<span class="text-vanna-navy font-medium">{execution.agent_name}</span>
											<StatusBadge status={execution.status || 'pending'} size="sm" />
										</div>
										<Button
											href="/task-executions/{execution.id}"
											variant="primary"
											size="xs"
										>
											View Execution
										</Button>
									</div>
									<div class="text-xs text-slate-500 mt-2">
										Started: {new Date(execution.created_at.Time).toLocaleString()}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<div class="pt-4 border-t border-slate-200">
					<h4 class="text-sm font-medium text-vanna-navy mb-3">Execute Task</h4>

					{#if availableAgents.length > 0}
						<div class="bg-vanna-cream/30 rounded-lg p-4">
							<p class="text-slate-600 mb-3">Select agents to execute this task:</p>

							<div class="space-y-2 mb-4 max-h-40 overflow-y-auto">
								{#each availableAgents as agent}
									<label class="flex items-center gap-3 p-2 rounded hover:bg-vanna-cream/50 cursor-pointer">
										<input
											type="checkbox"
											checked={selectedAgents.some(a => a.id === agent.id)}
											on:change={() => toggleAgentSelection(agent)}
											class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"
										/>
										<div class="flex-1">
											<div class="font-medium text-vanna-navy">{agent.name}</div>
											<div class="text-sm text-slate-500 font-mono">{agent.command} {agent.params}</div>
										</div>
									</label>
								{/each}
							</div>

							{#if selectedAgents.length > 0}
								<div class="text-sm text-slate-600 mb-3">
									Selected: {selectedAgents.map(a => a.name).join(', ')}
								</div>
							{/if}

							<Button
								onclick={startTaskExecution}
								disabled={selectedAgents.length === 0 || executingTask}
								loading={executingTask}
								variant="primary"
								class="w-full"
							>
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h1m4 0h1m-6-8h1m4 0h1m-2-4h.01M12 16h.01M12 8h.01M12 12h.01"/>
								</svg>
								{executingTask ? 'Starting Execution...' : `Start Execution (${selectedAgents.length})`}
							</Button>
						</div>
					{:else}
						<div class="bg-vanna-cream/30 rounded-lg p-4 text-center">
							<p class="text-slate-600 mb-3">No agents configured</p>
							<Button
								href="/agents"
								variant="warning"
							>
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
								</svg>
								Configure Agents
							</Button>
						</div>
					{/if}
				</div>
		</div>
	{/if}
</Modal>

<!-- Task Edit Modal -->
<Modal open={showEditTaskModal && editingTask} title="Edit Task" size="lg" onClose={closeEditTaskModal}>
	{#if editingTask}
		<form on:submit|preventDefault={saveTaskEdit} class="space-y-6">
			<FormField label="Task Title" id="edit-task-title" required>
				<Input 
					id="edit-task-title"
					type="text" 
					bind:value={editTaskForm.title}
					placeholder="Enter task title"
					required
				/>
			</FormField>
			
			<FormField label="Description" id="edit-task-description">
				<Textarea 
					id="edit-task-description"
					bind:value={editTaskForm.description}
					placeholder="Enter task description"
					rows={4}
				/>
			</FormField>
			
			<FormField label="Status" id="edit-task-status">
				<Select 
					id="edit-task-status"
					bind:value={editTaskForm.status}
				>
					{#each columns as column}
						<option value={column.id}>{column.title}</option>
					{/each}
				</Select>
			</FormField>
			
			<FormField label="Base Directory" id="edit-task-base-directory" required>
				<Select 
					id="edit-task-base-directory"
					bind:value={editTaskForm.baseDirectoryId}
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
			
			<div class="flex gap-3 pt-4">
				<Button 
					type="submit"
					disabled={updatingTasks.has(editingTask.id)}
					loading={updatingTasks.has(editingTask.id)}
					variant="primary"
				>
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
					</svg>
					{updatingTasks.has(editingTask.id) ? 'Saving...' : 'Save Changes'}
				</Button>
				<Button 
					type="button"
					onclick={closeEditTaskModal}
					variant="secondary"
				>
					Cancel
				</Button>
			</div>
		</form>
	{/if}
</Modal>
