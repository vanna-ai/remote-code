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
		agents_waiting_for_input: []
	};
	let loading = true;
	onMount(async () => {
		await loadDashboardStats();
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
