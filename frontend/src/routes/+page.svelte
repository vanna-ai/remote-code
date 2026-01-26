<script>
	import { onMount } from 'svelte';
	import StatsCard from '$lib/components/ui/StatsCard.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	let stats = {
		active_sessions: 0,
		projects: 0,
		task_executions: 0,
		agents: 0,
		git_changes_awaiting_review: [],
		agents_waiting_for_input: [],
		remote_ports: []
	};
	let loading = true;
	let newPortNumber = '';
	let creatingTunnel = false;

	onMount(async () => {
		await loadDashboardStats();
		// Poll for updates every 3 seconds to catch URL updates
		const interval = setInterval(loadDashboardStats, 3000);
		return () => clearInterval(interval);
	});

	async function loadDashboardStats() {
		try {
			const response = await fetch('/api/dashboard/stats');
			if (response.ok) {
				stats = await response.json();
			}
		} catch (error) {
			console.error('Failed to load dashboard stats:', error);
		} finally {
			loading = false;
		}
	}

	async function createTunnel() {
		const port = parseInt(newPortNumber);
		if (!port || port <= 0 || port > 65535) {
			alert('Please enter a valid port number (1-65535)');
			return;
		}

		creatingTunnel = true;
		try {
			const response = await fetch('/api/remote-ports', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ port })
			});
			if (response.ok) {
				newPortNumber = '';
				await loadDashboardStats();
			} else {
				const error = await response.text();
				alert('Failed to create tunnel: ' + error);
			}
		} catch (error) {
			console.error('Failed to create tunnel:', error);
			alert('Failed to create tunnel');
		} finally {
			creatingTunnel = false;
		}
	}

	async function stopTunnel(id) {
		try {
			const response = await fetch(`/api/remote-ports/${id}`, {
				method: 'DELETE'
			});
			if (response.ok) {
				await loadDashboardStats();
			} else {
				alert('Failed to stop tunnel');
			}
		} catch (error) {
			console.error('Failed to stop tunnel:', error);
			alert('Failed to stop tunnel');
		}
	}

	async function copyToClipboard(url) {
		try {
			await navigator.clipboard.writeText(url);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}
</script>

<svelte:head>
	<title>Dashboard - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-vanna-navy font-serif">Dashboard</h1>
		<p class="mt-2 text-slate-500">Welcome to your development environment management platform</p>
	</div>

	<!-- Stats Cards -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
		<StatsCard
			title="Active Sessions"
			value={loading ? '...' : stats.active_sessions}
			icon="terminal"
			color="teal"
			href="/terminal"
			{loading}
		/>
		<StatsCard
			title="Projects"
			value={loading ? '...' : stats.projects}
			icon="projects"
			color="navy"
			href="/projects"
			{loading}
		/>
		<StatsCard
			title="Task Executions"
			value={loading ? '...' : stats.task_executions}
			icon="tasks"
			color="magenta"
			href="/task-executions"
			{loading}
		/>
		<StatsCard
			title="Agents"
			value={loading ? '...' : stats.agents}
			icon="agents"
			color="orange"
			href="/agents"
			{loading}
		/>
	</div>

	<!-- Remote Ports Card -->
	<Card>
		<div class="flex items-center justify-between mb-4">
			<div class="flex items-center">
				<svg class="w-5 h-5 mr-2 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
				</svg>
				<h2 class="text-lg font-semibold text-vanna-navy">Remote Ports</h2>
				{#if stats.remote_ports && stats.remote_ports.length > 0}
					<span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-teal/10 text-vanna-teal rounded-full">
						{stats.remote_ports.length}
					</span>
				{/if}
			</div>
		</div>

		<!-- New Tunnel Form -->
		<div class="flex items-center gap-3 mb-4">
			<div class="flex-1 max-w-xs">
				<input
					type="number"
					bind:value={newPortNumber}
					placeholder="Port number (e.g., 3000)"
					class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-vanna-teal/50 focus:border-vanna-teal"
					onkeydown={(e) => e.key === 'Enter' && createTunnel()}
				/>
			</div>
			<Button
				variant="primary"
				onclick={createTunnel}
				disabled={creatingTunnel}
			>
				{#if creatingTunnel}
					<svg class="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
					Starting...
				{:else}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
					</svg>
					Start Tunnel
				{/if}
			</Button>
		</div>

		<!-- Active Tunnels List -->
		{#if stats.remote_ports && stats.remote_ports.length > 0}
			<div class="space-y-3">
				{#each stats.remote_ports as tunnel}
					<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg">
						<div class="flex-1">
							<div class="font-medium text-vanna-navy font-mono text-sm">
								localhost:{tunnel.port}
							</div>
							{#if tunnel.external_url}
								<a
									href={tunnel.external_url}
									target="_blank"
									rel="noopener noreferrer"
									class="text-sm text-vanna-teal hover:underline break-all"
								>
									{tunnel.external_url}
								</a>
							{:else}
								<div class="text-sm text-slate-400 italic">Waiting for URL...</div>
							{/if}
						</div>
						<div class="flex items-center space-x-2">
							<!-- Status Badge -->
							{#if tunnel.status === 'connected'}
								<span class="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded">
									connected
								</span>
							{:else if tunnel.status === 'starting'}
								<span class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-700 rounded animate-pulse">
									starting
								</span>
							{:else}
								<span class="px-2 py-1 text-xs font-medium bg-red-100 text-red-700 rounded">
									{tunnel.status}
								</span>
							{/if}

							<!-- Copy URL Button -->
							{#if tunnel.external_url}
								<button
									onclick={() => copyToClipboard(tunnel.external_url)}
									class="p-2 text-slate-500 hover:text-vanna-teal hover:bg-vanna-teal/10 rounded-lg transition-colors"
									title="Copy URL"
									aria-label="Copy URL to clipboard"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
									</svg>
								</button>
							{/if}

							<!-- Stop Button -->
							<Button
								size="sm"
								variant="danger"
								onclick={() => stopTunnel(tunnel.id)}
							>
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
								</svg>
								Stop
							</Button>
						</div>
					</div>
				{/each}
			</div>
		{:else if !loading}
			<p class="text-sm text-slate-500 text-center py-4">
				No active tunnels. Enter a port number above to expose a local service remotely.
			</p>
		{/if}
	</Card>

	<!-- Git & Agent Cards Side by Side -->
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		{#if !loading && stats.git_changes_awaiting_review.length > 0}
			<Card>
				<div class="flex items-center mb-4">
					<svg class="w-5 h-5 mr-2 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
					</svg>
					<h2 class="text-lg font-semibold text-vanna-navy">Git Changes Awaiting Review</h2>
					<span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded-full">
						{stats.git_changes_awaiting_review.length}
					</span>
				</div>
				<div class="space-y-3">
					{#each stats.git_changes_awaiting_review as item}
						<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg">
							<div class="flex-1">
								<div class="font-medium text-vanna-navy font-mono text-sm">{item.task_name}</div>
								<div class="text-sm text-slate-500">
									<span class="inline-flex items-center gap-1">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"/>
										</svg>
										{item.agent}
									</span>
								</div>
							</div>
							<div class="flex items-center space-x-2">
								<span class="px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded">
									uncommitted
								</span>
								<Button
									size="sm"
									variant="primary"
									onclick={() => window.location.href = `/git/${item.id}`}
								>
									<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
									</svg>
									Review
								</Button>
							</div>
						</div>
					{/each}
				</div>
			</Card>
		{/if}

		{#if !loading && stats.agents_waiting_for_input.length > 0}
			<Card>
				<div class="flex items-center mb-4">
					<svg class="w-5 h-5 mr-2 text-vanna-orange animate-pulse" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
					</svg>
					<h2 class="text-lg font-semibold text-vanna-navy">Agents Waiting for Input</h2>
					<span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded-full">
						{stats.agents_waiting_for_input.length}
					</span>
				</div>
				<div class="space-y-3">
					{#each stats.agents_waiting_for_input as execution}
						<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg">
							<div class="flex-1">
								<div class="text-xs font-semibold text-vanna-teal uppercase tracking-wide mb-1">
									{execution.project_name}
								</div>
								<div class="font-medium text-vanna-navy">{execution.task_name}</div>
								<div class="text-sm text-slate-500">Agent: {execution.agent}</div>
							</div>
							<div class="flex items-center space-x-2">
								<span class="px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded animate-pulse">
									Waiting
								</span>
								<Button
									size="sm"
									variant="primary"
									onclick={() => window.location.href = `/task-executions/${execution.id}`}
								>
									Check Session
								</Button>
							</div>
						</div>
					{/each}
				</div>
			</Card>
		{/if}
	</div>


</div>
