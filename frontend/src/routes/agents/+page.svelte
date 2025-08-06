<script>
	import { onMount } from 'svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	
	const breadcrumbSegments = [
		{ label: "Dashboard", href: "/" },
		{ label: "Agents", href: "/agents" }
	];
	
	let agents = [];
	let availableAgents = [];
	let loading = true;
	let detecting = false;
	let showAddAgentForm = false;
	let showConfigureForm = false;
	let configuringAgent = null;
	let newAgent = { name: '', command: '', params: '' };

	onMount(async () => {
		await loadAgents();
	});

	async function loadAgents() {
		try {
			loading = true;
			const response = await fetch('/api/agents');
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			agents = await response.json();
			loading = false;
			
			// Always detect available agents when page loads
			await detectAgents();
			
		} catch (error) {
			console.error('Failed to load agents:', error);
			loading = false;
		}
	}
	
	async function detectAgents() {
		try {
			detecting = true;
			const response = await fetch('/api/agents/detect');
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			const data = await response.json();
			availableAgents = data.agents;
			detecting = false;
		} catch (error) {
			console.error('Failed to detect agents:', error);
			detecting = false;
		}
	}
	
	async function addAgent(agentData) {
		try {
			const response = await fetch('/api/agents', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					root_id: 1, // Default root for now
					name: agentData.name,
					command: agentData.command,
					params: agentData.params || ''
				})
			});
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			const createdAgent = await response.json();
			agents = [...agents, createdAgent];
			
			// Reset form
			newAgent = { name: '', command: '', params: '' };
			showAddAgentForm = false;
		} catch (error) {
			console.error('Failed to add agent:', error);
			alert('Failed to add agent. Please try again.');
		}
	}
	
	function addDetectedAgent(detectedAgent) {
		const agentName = detectedAgent.name.charAt(0).toUpperCase() + detectedAgent.name.slice(1);
		addAgent({
			name: `${agentName} Assistant`,
			command: detectedAgent.command,
			params: ''
		});
	}
	
	async function removeAgent(agentId) {
		if (!confirm('Are you sure you want to remove this agent? This action cannot be undone.')) {
			return;
		}
		
		try {
			const response = await fetch(`/api/agents/${agentId}`, {
				method: 'DELETE'
			});
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			// Remove the agent from local state
			agents = agents.filter(agent => agent.id !== agentId);
		} catch (error) {
			console.error('Failed to remove agent:', error);
			alert('Failed to remove agent. Please try again.');
		}
	}
	
	function startConfiguring(agent) {
		configuringAgent = { ...agent };
		showConfigureForm = true;
	}
	
	async function updateAgent() {
		try {
			const response = await fetch(`/api/agents/${configuringAgent.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					name: configuringAgent.name,
					command: configuringAgent.command,
					params: configuringAgent.params
				})
			});
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			const updatedAgent = await response.json();
			
			// Update the agent in local state
			agents = agents.map(agent => 
				agent.id === updatedAgent.id ? updatedAgent : agent
			);
			
			// Close form
			showConfigureForm = false;
			configuringAgent = null;
		} catch (error) {
			console.error('Failed to update agent:', error);
			alert('Failed to update agent. Please try again.');
		}
	}
	
	function cancelConfigure() {
		showConfigureForm = false;
		configuringAgent = null;
	}
</script>

<svelte:head>
	<title>Agents - Remote-Code</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />
		
		<!-- Header -->
		<div class="mb-6">
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
				<button 
					on:click={() => showAddAgentForm = true}
					class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
					</svg>
					Add Agent
				</button>
			</div>
		</div>

		<!-- Add Agent Form -->
		{#if showAddAgentForm}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-xl font-semibold text-white mb-4">Add New Agent</h3>
				<form on:submit|preventDefault={() => addAgent(newAgent)} class="space-y-4">
					<div>
						<label for="agent-name" class="block text-sm font-medium text-gray-300 mb-2">
							Agent Name
						</label>
						<input 
							id="agent-name"
							type="text" 
							bind:value={newAgent.name}
							placeholder="e.g., Claude Assistant"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
							required
						/>
					</div>
					<div>
						<label for="agent-command" class="block text-sm font-medium text-gray-300 mb-2">
							Command
						</label>
						<input 
							id="agent-command"
							type="text" 
							bind:value={newAgent.command}
							placeholder="e.g., claude"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
							required
						/>
					</div>
					<div>
						<label for="agent-params" class="block text-sm font-medium text-gray-300 mb-2">
							Parameters (optional)
						</label>
						<input 
							id="agent-params"
							type="text" 
							bind:value={newAgent.params}
							placeholder="e.g., --model claude-3-sonnet"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
						/>
					</div>
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Add Agent
						</button>
						<button 
							type="button"
							on:click={() => showAddAgentForm = false}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Configure Agent Form -->
		{#if showConfigureForm && configuringAgent}
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-6">
				<h3 class="text-xl font-semibold text-white mb-4">Configure Agent</h3>
				<form on:submit|preventDefault={updateAgent} class="space-y-4">
					<div>
						<label for="configure-name" class="block text-sm font-medium text-gray-300 mb-2">
							Agent Name
						</label>
						<input 
							id="configure-name"
							type="text" 
							bind:value={configuringAgent.name}
							placeholder="e.g., Claude Assistant"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
							required
						/>
					</div>
					<div>
						<label for="configure-command" class="block text-sm font-medium text-gray-300 mb-2">
							Command
						</label>
						<input 
							id="configure-command"
							type="text" 
							bind:value={configuringAgent.command}
							placeholder="e.g., claude"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
							required
						/>
					</div>
					<div>
						<label for="configure-params" class="block text-sm font-medium text-gray-300 mb-2">
							Parameters (optional)
						</label>
						<input 
							id="configure-params"
							type="text" 
							bind:value={configuringAgent.params}
							placeholder="e.g., --model claude-3-sonnet"
							class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-orange-400"
						/>
					</div>
					<div class="flex gap-3">
						<button 
							type="submit"
							class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Update Agent
						</button>
						<button 
							type="button"
							on:click={cancelConfigure}
							class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Loading State -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-400"></div>
				<span class="ml-3 text-gray-300">Loading agents...</span>
			</div>
		{:else}
			<!-- Configured Agents Section -->
			{#if agents.length > 0}
				<div class="mb-6">
					<h3 class="text-lg font-semibold text-white mb-4">Configured Agents</h3>
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
										<h4 class="text-lg font-semibold text-white">{agent.name}</h4>
									</div>
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium text-white bg-green-500">
										Configured
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
									<button 
										on:click={() => startConfiguring(agent)}
										class="flex-1 bg-orange-500 hover:bg-orange-600 text-white px-3 py-2 rounded text-sm transition-colors"
									>
										Configure
									</button>
									<button 
										on:click={() => removeAgent(agent.id)}
										class="bg-red-600 hover:bg-red-700 text-white px-3 py-2 rounded text-sm transition-colors"
									>
										Remove
									</button>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Available Agents Section -->
			<div class="bg-gray-800 rounded-lg border border-gray-700 p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-lg font-semibold text-white">Available Agents</h3>
					{#if detecting}
						<div class="flex items-center gap-2">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-orange-400"></div>
							<span class="text-sm text-gray-400">Detecting...</span>
						</div>
					{:else}
						<button 
							on:click={detectAgents}
							class="bg-gray-600 hover:bg-gray-700 text-white px-3 py-1 rounded text-sm transition-colors"
						>
							Refresh Detection
						</button>
					{/if}
				</div>
				
				{#if availableAgents.length > 0}
					<p class="text-gray-400 mb-4">
						AI development agents found on your system. Click "Add" to configure them.
					</p>
					
					{#if availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)).length > 0}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							{#each availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)) as agent}
								<div class="bg-gray-700 rounded-lg p-4 border border-gray-600">
									<div class="flex items-center justify-between">
										<div>
											<h4 class="font-medium text-white">{agent.name}</h4>
											<p class="text-sm text-gray-400 font-mono">{agent.path}</p>
										</div>
										<button 
											on:click={() => addDetectedAgent(agent)}
											class="bg-orange-500 hover:bg-orange-600 text-white px-3 py-1 rounded text-sm transition-colors"
										>
											Add
										</button>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-gray-500 text-center py-4">
							All available agents are already configured.
						</p>
					{/if}
				{:else}
					<p class="text-gray-500 text-center py-4">
						{#if detecting}
							Scanning for available agents...
						{:else}
							No AI development agents were detected. You can manually add agents using the "Add Agent" button above.
						{/if}
					</p>
				{/if}
			</div>
			
			<!-- Empty State (only when no configured agents AND no detection has run) -->
			{#if agents.length === 0 && availableAgents.length === 0 && !detecting}
				<div class="text-center py-12">
					<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-gray-300 mb-2">No Agents Configured</h3>
					<p class="text-gray-400 mb-4">Add AI development agents to execute tasks</p>
					<button 
						on:click={detectAgents}
						class="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors"
					>
						Detect Available Agents
					</button>
				</div>
			{/if}
		{/if}
	</div>
</div>