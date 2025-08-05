<script>
	import { onMount } from 'svelte';
	
	let projects = [];
	let loading = true;
	let showCreateForm = false;
	let newProject = { name: '' };

	onMount(async () => {
		await loadProjects();
	});

	async function loadProjects() {
		try {
			loading = true;
			const response = await fetch('/api/projects');
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			projects = await response.json();
			loading = false;
		} catch (error) {
			console.error('Failed to load projects:', error);
			// Fallback to empty array if API fails
			projects = [];
			loading = false;
		}
	}

	async function createProject() {
		if (!newProject.name.trim()) return;
		
		try {
			const response = await fetch('/api/projects', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					root_id: 1, // For now, use a default root_id
					name: newProject.name
				})
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const createdProject = await response.json();
			projects = [...projects, createdProject];
			newProject = { name: '' };
			showCreateForm = false;
		} catch (error) {
			console.error('Failed to create project:', error);
			alert('Failed to create project. Please try again.');
		}
	}
</script>

<svelte:head>
	<title>Projects - Remote-Code</title>
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
					<div class="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
						</svg>
					</div>
					<div>
						<h1 class="text-3xl font-bold text-blue-400 mb-1">Projects</h1>
						<p class="text-gray-300">Manage your development projects and repositories</p>
					</div>
				</div>
				<button 
					on:click={() => showCreateForm = true}
					class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					New Project
				</button>
			</div>
		</div>

		<!-- Create Project Form -->
		{#if showCreateForm}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-xl font-semibold text-white mb-4">Create New Project</h3>
				<form on:submit|preventDefault={createProject} class="space-y-4">
					<div>
						<label for="project-name" class="block text-sm font-medium text-gray-300 mb-2">
							Project Name
						</label>
						<input 
							id="project-name"
							type="text" 
							bind:value={newProject.name}
							placeholder="Enter project name"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-400"
							required
						/>
					</div>
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Create Project
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

		<!-- Projects Grid -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-400"></div>
			</div>
		{:else if projects.length === 0}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-gray-300 mb-2">No Projects Yet</h3>
				<p class="text-gray-400 mb-4">Create your first project to get started</p>
				<button 
					on:click={() => showCreateForm = true}
					class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
				>
					Create Project
				</button>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each projects as project}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 hover:border-blue-400 transition-colors">
						<div class="flex items-center gap-3 mb-4">
							<div class="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center">
								<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
								</svg>
							</div>
							<h3 class="text-lg font-semibold text-white">{project.name}</h3>
						</div>
						
						<div class="space-y-2 mb-4">
							<div class="flex items-center text-sm text-gray-400">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
								</svg>
								{(project.baseDirectories || []).length} directories
							</div>
							<div class="flex items-center text-sm text-gray-400">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
								</svg>
								{(project.tasks || []).length} tasks
							</div>
						</div>

						<div class="flex gap-2">
							<button class="flex-1 bg-blue-500 hover:bg-blue-600 text-white px-3 py-2 rounded text-sm transition-colors">
								Open
							</button>
							<button class="bg-gray-600 hover:bg-gray-700 text-white px-3 py-2 rounded text-sm transition-colors">
								Settings
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>