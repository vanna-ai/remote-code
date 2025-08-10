<script>
	import { onMount } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	
	let agents = [];
	let availableAgents = [];
	let leaderboard = [];
	let loading = true;
	let detecting = false;
	let showAddAgentForm = false;
	let showConfigureForm = false;
	let configuringAgent = null;
	let newAgent = { name: '', command: '', params: '' };
	let showLeaderboard = true;

	onMount(async () => {
		await loadAgents();
		await loadLeaderboard();
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
	
	async function loadLeaderboard() {
		try {
			const response = await fetch('/api/elo/leaderboard');
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			leaderboard = await response.json();
		} catch (error) {
			console.error('Failed to load leaderboard:', error);
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
			
			// Reload leaderboard to include new agent
			await loadLeaderboard();
			
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

	function getAgentELO(agentId) {
		const agentELO = leaderboard.find(agent => agent.id === agentId);
		return agentELO ? {
			rating: agentELO.elo_rating?.Valid ? Math.round(agentELO.elo_rating.Float64) : 1500,
			games: agentELO.games_played?.Valid ? agentELO.games_played.Int64 : 0,
			wins: agentELO.wins?.Valid ? agentELO.wins.Int64 : 0,
			losses: agentELO.losses?.Valid ? agentELO.losses.Int64 : 0,
			draws: agentELO.draws?.Valid ? agentELO.draws.Int64 : 0,
			winPercentage: agentELO.win_percentage || 0,
			rank: agentELO.elo_rank || '-'
		} : {
			rating: 1500,
			games: 0,
			wins: 0,
			losses: 0,
			draws: 0,
			winPercentage: 0,
			rank: '-'
		};
	}

	function formatELORating(rating) {
		if (rating >= 2400) return { color: 'text-purple-400', label: 'Master' };
		if (rating >= 2200) return { color: 'text-blue-400', label: 'Expert' };
		if (rating >= 2000) return { color: 'text-green-400', label: 'Advanced' };
		if (rating >= 1800) return { color: 'text-yellow-400', label: 'Intermediate' };
		if (rating >= 1600) return { color: 'text-orange-400', label: 'Beginner' };
		return { color: 'text-gray-400', label: 'Novice' };
	}
</script>

<svelte:head>
	<title>Agents - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Agents</h1>
			<p class="mt-2 text-gray-600 dark:text-gray-400">Configure and manage AI development agents</p>
		</div>
		<div class="flex items-center space-x-3">
			<Button onclick={() => showLeaderboard = !showLeaderboard} variant="secondary">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
				</svg>
				{showLeaderboard ? 'Hide' : 'Show'} ELO Rankings
			</Button>
			<Button onclick={() => showAddAgentForm = true} variant="primary">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
				</svg>
				Add Agent
			</Button>
		</div>
	</div>

	<!-- Add Agent Form -->
	{#if showAddAgentForm}
		<Card>
			<h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Add New Agent</h3>
			<form on:submit|preventDefault={() => addAgent(newAgent)} class="space-y-4">
				<div>
					<label for="agent-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
					<label for="agent-command" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
					<label for="agent-params" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
			<h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Configure Agent</h3>
			<form on:submit|preventDefault={updateAgent} class="space-y-4">
				<div>
					<label for="configure-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
					<label for="configure-command" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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
					<label for="configure-params" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
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

	<!-- ELO Leaderboard -->
	{#if showLeaderboard && leaderboard.length > 0}
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-xl font-semibold text-gray-900 dark:text-white">ELO Leaderboard</h3>
				<Button onclick={loadLeaderboard} variant="ghost" size="sm">
					Refresh
				</Button>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead>
						<tr class="border-b border-gray-200 dark:border-gray-700">
							<th class="text-left py-3 text-gray-500 dark:text-gray-400 font-medium">Rank</th>
							<th class="text-left py-3 text-gray-500 dark:text-gray-400 font-medium">Agent</th>
							<th class="text-center py-3 text-gray-500 dark:text-gray-400 font-medium">ELO Rating</th>
							<th class="text-center py-3 text-gray-500 dark:text-gray-400 font-medium">Games</th>
							<th class="text-center py-3 text-gray-500 dark:text-gray-400 font-medium">W/L/D</th>
							<th class="text-center py-3 text-gray-500 dark:text-gray-400 font-medium">Win %</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
						{#each leaderboard as agent, index}
							{@const eloInfo = formatELORating(agent.elo_rating?.Valid ? Math.round(agent.elo_rating.Float64) : 1500)}
							<tr class="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
								<td class="py-4">
									<span class="inline-flex items-center justify-center w-8 h-8 rounded-full {index < 3 ? 'bg-yellow-500 text-black' : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300'} text-sm font-bold">
										{index + 1}
									</span>
								</td>
								<td class="py-4">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
											<svg class="w-4 h-4 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
											</svg>
										</div>
										<span class="text-gray-900 dark:text-white font-medium">{agent.name}</span>
									</div>
								</td>
								<td class="py-4 text-center">
									<div class="flex flex-col items-center">
										<span class="text-lg font-bold text-gray-900 dark:text-white">
											{agent.elo_rating?.Valid ? Math.round(agent.elo_rating.Float64) : 1500}
										</span>
										<Badge variant={eloInfo.color === 'text-purple-400' ? 'primary' : eloInfo.color === 'text-green-400' ? 'success' : 'secondary'} size="sm">
											{eloInfo.label}
										</Badge>
									</div>
								</td>
								<td class="py-4 text-center text-gray-900 dark:text-white">
									{agent.games_played?.Valid ? agent.games_played.Int64 : 0}
								</td>
								<td class="py-4 text-center">
									<div class="flex justify-center gap-1 text-sm">
										<span class="text-green-600 dark:text-green-400">{agent.wins?.Valid ? agent.wins.Int64 : 0}</span>
										<span class="text-gray-400">/</span>
										<span class="text-red-600 dark:text-red-400">{agent.losses?.Valid ? agent.losses.Int64 : 0}</span>
										<span class="text-gray-400">/</span>
										<span class="text-yellow-600 dark:text-yellow-400">{agent.draws?.Valid ? agent.draws.Int64 : 0}</span>
									</div>
								</td>
								<td class="py-4 text-center text-gray-900 dark:text-white">
									{agent.win_percentage || 0}%
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			<span class="ml-3 text-gray-600 dark:text-gray-400">Loading agents...</span>
		</div>
	{:else}
		<!-- Configured Agents Section -->
		{#if agents.length > 0}
			<div class="mb-6">
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Configured Agents</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					{#each agents as agent}
						{@const eloData = getAgentELO(agent.id)}
						{@const eloInfo = formatELORating(eloData.rating)}
						<Card class="card-hover">
							<div class="flex items-center justify-between mb-4">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
										<svg class="w-5 h-5 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
										</svg>
									</div>
									<div>
										<h4 class="text-lg font-semibold text-gray-900 dark:text-white">{agent.name}</h4>
										<div class="flex items-center gap-2">
											<span class="text-xs text-gray-600 dark:text-gray-400">ELO: {eloData.rating}</span>
											<span class="text-xs text-gray-500 dark:text-gray-500">({eloData.games} games)</span>
										</div>
									</div>
								</div>
								<div class="flex items-center gap-2">
									{#if eloData.games > 0}
										<Badge variant="secondary" size="sm">
											#{eloData.rank}
										</Badge>
									{/if}
									<Badge variant="success" size="sm">
										Configured
									</Badge>
								</div>
							</div>
							
							{#if eloData.games > 0}
								<div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-3 mb-4">
									<div class="grid grid-cols-4 gap-4 text-center text-sm">
										<div>
											<div class="text-green-600 dark:text-green-400 font-semibold">{eloData.wins}</div>
											<div class="text-gray-500 dark:text-gray-400 text-xs">Wins</div>
										</div>
										<div>
											<div class="text-red-600 dark:text-red-400 font-semibold">{eloData.losses}</div>
											<div class="text-gray-500 dark:text-gray-400 text-xs">Losses</div>
										</div>
										<div>
											<div class="text-yellow-600 dark:text-yellow-400 font-semibold">{eloData.draws}</div>
											<div class="text-gray-500 dark:text-gray-400 text-xs">Draws</div>
										</div>
										<div>
											<div class="text-gray-900 dark:text-white font-semibold">{eloData.winPercentage}%</div>
											<div class="text-gray-500 dark:text-gray-400 text-xs">Win Rate</div>
										</div>
									</div>
								</div>
							{/if}
							
							<div class="space-y-2 mb-4">
								<div class="text-sm">
									<span class="text-gray-600 dark:text-gray-400">Command:</span>
									<span class="text-gray-900 dark:text-white font-mono ml-2">{agent.command}</span>
								</div>
								{#if agent.params}
									<div class="text-sm">
										<span class="text-gray-600 dark:text-gray-400">Parameters:</span>
										<span class="text-gray-900 dark:text-white font-mono ml-2">{agent.params}</span>
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
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white">Available Agents</h3>
				{#if detecting}
					<div class="flex items-center gap-2">
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
						<span class="text-sm text-gray-600 dark:text-gray-400">Detecting...</span>
					</div>
				{:else}
					<Button onclick={detectAgents} variant="ghost" size="sm">
						Refresh Detection
					</Button>
				{/if}
			</div>
			
			{#if availableAgents.length > 0}
				<p class="text-gray-600 dark:text-gray-400 mb-4">
					AI development agents found on your system. Click "Add" to configure them.
				</p>
				
				{#if availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)).length > 0}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						{#each availableAgents.filter(agent => agent.available && !agents.some(configured => configured.command === agent.command)) as agent}
							<div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 border border-gray-200 dark:border-gray-600">
								<div class="flex items-center justify-between">
									<div>
										<h4 class="font-medium text-gray-900 dark:text-white">{agent.name}</h4>
										<p class="text-sm text-gray-600 dark:text-gray-400 font-mono">{agent.path}</p>
									</div>
									<Button onclick={() => addDetectedAgent(agent)} variant="primary" size="sm">
										Add
									</Button>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-gray-500 dark:text-gray-400 text-center py-4">
						All available agents are already configured.
					</p>
				{/if}
			{:else}
				<p class="text-gray-500 dark:text-gray-400 text-center py-4">
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
				<div class="w-16 h-16 bg-gray-100 dark:bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">No Agents Configured</h3>
				<p class="text-gray-600 dark:text-gray-400 mb-4">Add AI development agents to execute tasks</p>
				<Button onclick={detectAgents} variant="primary">
					Detect Available Agents
				</Button>
			</Card>
		{/if}
	{/if}
</div>