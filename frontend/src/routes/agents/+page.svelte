<script>
	import { onMount } from 'svelte';
	
	let agents = [];
	let loading = true;

	onMount(async () => {
		await loadAgents();
	});

	async function loadAgents() {
		try {
			// Mock data for now
			agents = [
				{
					id: 1,
					name: "Claude Code Assistant",
					command: "claude-code",
					params: "--model claude-3-sonnet",
					status: "active"
				},
				{
					id: 2,
					name: "GitHub Copilot",
					command: "copilot",
					params: "",
					status: "inactive"
				}
			];
			loading = false;
		} catch (error) {
			console.error('Failed to load agents:', error);
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Agents - Remote-Code</title>
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
					<div class="w-12 h-12 bg-orange-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
						</svg>
					</div>
					<div>
						<h1 class="text-3xl font-bold text-orange-400 mb-1">Agents</h1>
						<p class="text-gray-300">Configure and manage AI development agents</p>
					</div>
				</div>
				<button class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					Add Agent
				</button>
			</div>
		</div>

		<!-- Agents List -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-400"></div>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				{#each agents as agent}
					<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 hover:border-orange-400 transition-colors">
						<div class="flex items-center justify-between mb-4">
							<div class="flex items-center gap-3">
								<div class="w-10 h-10 bg-orange-500 rounded-lg flex items-center justify-center">
									<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
									</svg>
								</div>
								<h3 class="text-lg font-semibold text-white">{agent.name}</h3>
							</div>
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white {agent.status === 'active' ? 'bg-green-500' : 'bg-gray-500'}">
								{agent.status}
							</span>
						</div>
						
						<div class="space-y-2 mb-4">
							<div class="text-sm">
								<span class="text-gray-400">Command:</span>
								<span class="text-white font-mono ml-2">{agent.command}</span>
							</div>
							{#if agent.params}
								<div class="text-sm">
									<span class="text-gray-400">Parameters:</span>
									<span class="text-white font-mono ml-2">{agent.params}</span>
								</div>
							{/if}
						</div>

						<div class="flex gap-2">
							<button class="flex-1 bg-orange-500 hover:bg-orange-600 text-white px-3 py-2 rounded text-sm transition-colors">
								Configure
							</button>
							<button class="bg-gray-600 hover:bg-gray-700 text-white px-3 py-2 rounded text-sm transition-colors">
								{agent.status === 'active' ? 'Disable' : 'Enable'}
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>