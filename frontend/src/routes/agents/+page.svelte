<script>
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	
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

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-vanna-navy font-serif">Agents</h1>
			<p class="mt-2 text-slate-500">Configure and manage AI development agents</p>
		</div>
		<Button onclick={() => showAddAgentForm = true} variant="primary">
			<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
			</svg>
			Add Agent
		</Button>
	</div>

	<!-- Add Agent Form -->
	{#if showAddAgentForm}
		<Card>
			<h3 class="text-xl font-semibold text-vanna-navy mb-4">Add New Agent</h3>
			<form on:submit|preventDefault={() => addAgent(newAgent)} class="space-y-4">
				<div>
					<label for="agent-name" class="block text-sm font-medium text-vanna-navy mb-2">
						Agent Name
					</label>
					<input 
						id="agent-name"
						type="text" 
						bind:value={newAgent.name}
						placeholder="e.g., Claude Assistant"
						class="input-field"
						required
					/>
				</div>
				<div>
					<label for="agent-command" class="block text-sm font-medium text-vanna-navy mb-2">
						Command
					</label>
					<input 
						id="agent-command"
						type="text" 
						bind:value={newAgent.command}
						placeholder="e.g., claude"
						class="input-field"
						required
					/>
				</div>
				<div>
					<label for="agent-params" class="block text-sm font-medium text-vanna-navy mb-2">
						Parameters (optional)
					</label>
					<input 
						id="agent-params"
						type="text" 
						bind:value={newAgent.params}
						placeholder="e.g., --model claude-3-sonnet"
						class="input-field"
					/>
				</div>
				<div class="flex gap-3">
					<Button type="submit" variant="primary">
						Add Agent
					</Button>
					<Button type="button" onclick={() => showAddAgentForm = false} variant="secondary">
						Cancel
					</Button>
				</div>
			</form>
		</Card>
	{/if}

	<!-- Configure Agent Form -->
	{#if showConfigureForm && configuringAgent}
		<Card>
			<h3 class="text-xl font-semibold text-vanna-navy mb-4">Configure Agent</h3>
			<form on:submit|preventDefault={updateAgent} class="space-y-4">
				<div>
					<label for="configure-name" class="block text-sm font-medium text-vanna-navy mb-2">
						Agent Name
					</label>
					<input 
						id="configure-name"
						type="text" 
						bind:value={configuringAgent.name}
						placeholder="e.g., Claude Assistant"
						class="input-field"
						required
					/>
				</div>
				<div>
					<label for="configure-command" class="block text-sm font-medium text-vanna-navy mb-2">
						Command
					</label>
					<input 
						id="configure-command"
						type="text" 
						bind:value={configuringAgent.command}
						placeholder="e.g., claude"
						class="input-field"
						required
					/>
				</div>
				<div>
					<label for="configure-params" class="block text-sm font-medium text-vanna-navy mb-2">
						Parameters (optional)
					</label>
					<input 
						id="configure-params"
						type="text" 
						bind:value={configuringAgent.params}
						placeholder="e.g., --model claude-3-sonnet"
						class="input-field"
					/>
				</div>
				<div class="flex gap-3">
					<Button type="submit" variant="primary">
						Update Agent
					</Button>
					<Button type="button" onclick={cancelConfigure} variant="secondary">
						Cancel
					</Button>
				</div>
			</form>
		</Card>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div>
			<span class="ml-3 text-slate-500">Loading agents...</span>
		</div>
	{:else}
		<!-- Configured Agents Section -->
		{#if agents.length > 0}
			<div class="mb-6">
				<h3 class="text-lg font-semibold text-vanna-navy mb-4">Configured Agents</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					{#each agents as agent}
						<Card class="card-hover">
							<div class="flex items-center justify-between mb-4">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 bg-vanna-orange/10 rounded-xl flex items-center justify-center">
										<svg class="w-5 h-5 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
										</svg>
									</div>
									<h4 class="text-lg font-semibold text-vanna-navy">{agent.name}</h4>
								</div>
								<Badge variant="success" size="sm">
									Configured
								</Badge>
							</div>

							<div class="space-y-2 mb-4">
								<div class="text-sm">
									<span class="text-slate-500">Command:</span>
									<span class="text-vanna-navy font-mono ml-2">{agent.command}</span>
								</div>
								{#if agent.params}
									<div class="text-sm">
										<span class="text-slate-500">Parameters:</span>
										<span class="text-vanna-navy font-mono ml-2">{agent.params}</span>
									</div>
								{/if}
							</div>

							<div class="flex gap-2">
								<Button onclick={() => startConfiguring(agent)} variant="primary" class="flex-1">
									Configure
								</Button>
								<Button onclick={() => removeAgent(agent.id)} variant="danger">
									Remove
								</Button>
							</div>
						</Card>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Available Agents Section -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-lg font-semibold text-vanna-navy">Available Agents</h3>
				{#if detecting}
					<div class="flex items-center gap-2">
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-vanna-teal"></div>
						<span class="text-sm text-slate-500">Detecting...</span>
					</div>
				{:else}
					<Button onclick={detectAgents} variant="ghost" size="sm">
						Refresh Detection
					</Button>
				{/if}
			</div>
			
			{#if availableAgents.length > 0}
				<p class="text-slate-500 mb-4">
					AI development agents found on your system. Click "Add" to configure them.
				</p>

				{#if availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)).length > 0}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						{#each availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)) as agent}
							<div class="bg-vanna-cream/30 rounded-lg p-4 border border-slate-200">
								<div class="flex items-center justify-between">
									<div>
										<h4 class="font-medium text-vanna-navy">{agent.name}</h4>
										<p class="text-sm text-slate-500 font-mono">{agent.path}</p>
									</div>
									<Button onclick={() => addDetectedAgent(agent)} variant="primary" size="sm">
										Add
									</Button>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-slate-500 text-center py-4">
						All available agents are already configured.
					</p>
				{/if}
			{:else}
				<p class="text-slate-500 text-center py-4">
					{#if detecting}
						Scanning for available agents...
					{:else}
						No AI development agents were detected. You can manually add agents using the "Add Agent" button above.
					{/if}
				</p>
			{/if}
		</Card>

		<!-- Empty State (only when no configured agents AND no detection has run) -->
		{#if agents.length === 0 && availableAgents.length === 0 && !detecting}
			<Card class="text-center py-12">
				<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-vanna-navy mb-2">No Agents Configured</h3>
				<p class="text-slate-500 mb-4">Add AI development agents to execute tasks</p>
				<Button onclick={detectAgents} variant="primary">
					Detect Available Agents
				</Button>
			</Card>
		{/if}
	{/if}
</div>