<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	
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
					root_id: 1,
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

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-vanna-navy font-serif">Projects</h1>
			<p class="mt-2 text-slate-500">Manage your development projects and repositories</p>
		</div>
		<Button onclick={() => showCreateForm = true} variant="primary">
			<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
			</svg>
			New Project
		</Button>
	</div>

	<!-- Create Project Form -->
	{#if showCreateForm}
		<Card>
			<h3 class="text-xl font-semibold text-vanna-navy mb-4">Create New Project</h3>
			<form on:submit|preventDefault={createProject} class="space-y-4">
				<div>
					<label for="project-name" class="block text-sm font-medium text-vanna-navy mb-2">
						Project Name
					</label>
					<input 
						id="project-name"
						type="text" 
						bind:value={newProject.name}
						placeholder="Enter project name"
						class="input-field"
						required
					/>
				</div>
				<div class="flex gap-3">
					<Button type="submit" variant="primary">
						Create Project
					</Button>
					<Button type="button" onclick={() => showCreateForm = false} variant="secondary">
						Cancel
					</Button>
				</div>
			</form>
		</Card>
	{/if}

	<!-- Projects Content -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
		</div>
	{:else if projects.length === 0}
		<Card class="text-center py-12">
			<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
				<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
				</svg>
			</div>
			<h3 class="text-xl font-semibold text-vanna-navy mb-2">No Projects Yet</h3>
			<p class="text-slate-500 mb-4">Create your first project to get started</p>
			<Button onclick={() => showCreateForm = true} variant="primary">
				Create Project
			</Button>
		</Card>
	{:else}
		<!-- Projects Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each projects as project}
				<Card class="card-hover group cursor-pointer" onclick={() => goto(`/projects/${project.id}`)}>
					<div class="flex items-start justify-between mb-4">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 bg-vanna-teal/10 rounded-xl flex items-center justify-center">
								<svg class="w-5 h-5 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
								</svg>
							</div>
							<div>
								<h3 class="text-lg font-semibold text-vanna-navy group-hover:text-vanna-teal transition-colors">
									{project.name}
								</h3>
							</div>
						</div>
						<div class="opacity-0 group-hover:opacity-100 transition-opacity">
							<svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
							</svg>
						</div>
					</div>

					<div class="space-y-3 mb-4">
						<div class="flex items-center justify-between">
							<div class="flex items-center text-sm text-slate-500">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
								</svg>
								Directories
							</div>
							<Badge variant="secondary" size="sm">
								{(project.baseDirectories || []).length}
							</Badge>
						</div>
						<div class="flex items-center justify-between">
							<div class="flex items-center text-sm text-slate-500">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
								</svg>
								Tasks
							</div>
							<Badge variant="primary" size="sm">
								{(project.tasks || []).length}
							</Badge>
						</div>
					</div>

					<div class="pt-4 border-t border-slate-200">
						<div class="flex items-center justify-between">
							<span class="text-sm text-slate-500">
								{project.created_at ? new Date(project.created_at).toLocaleDateString() : 'Recently created'}
							</span>
							<div class="flex items-center text-sm text-vanna-teal font-medium">
								Open Project
								<svg class="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
								</svg>
							</div>
						</div>
					</div>
				</Card>
			{/each}
		</div>

		<!-- Projects Table View (Alternative) -->
		<Card class="hidden">
			<div class="px-4 py-5 sm:p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-lg font-medium text-vanna-navy">All Projects</h3>
					<div class="flex items-center space-x-2">
						<Button variant="ghost" size="sm">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/>
							</svg>
							Filter
						</Button>
						<Button variant="ghost" size="sm">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12"/>
							</svg>
							Sort
						</Button>
					</div>
				</div>
				<div class="overflow-hidden">
					<table class="min-w-full divide-y divide-slate-200">
						<thead class="bg-vanna-cream/30">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Directories</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Tasks</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Created</th>
								<th class="relative px-6 py-3"><span class="sr-only">Actions</span></th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-slate-200">
							{#each projects as project}
								<tr class="table-row">
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="flex items-center">
											<div class="w-8 h-8 bg-vanna-teal/10 rounded-lg flex items-center justify-center mr-3">
												<svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
												</svg>
											</div>
											<div class="text-sm font-medium text-vanna-navy">{project.name}</div>
										</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<Badge variant="secondary" size="sm">{(project.baseDirectories || []).length}</Badge>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<Badge variant="primary" size="sm">{(project.tasks || []).length}</Badge>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
										{project.created_at ? new Date(project.created_at).toLocaleDateString() : 'Recently'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
										<Button onclick={() => goto(`/projects/${project.id}`)} variant="ghost" size="sm">
											Open
										</Button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</Card>
	{/if}
</div>