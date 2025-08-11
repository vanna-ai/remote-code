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
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
		<p class="mt-2 text-gray-600 dark:text-gray-400">Welcome to your development environment management platform</p>
	</div>

	<!-- Stats Cards -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
		<StatsCard
			title="Active Sessions"
			value={loading ? '...' : stats.active_sessions}
			icon="terminal"
			color="green"
			href="/terminal"
			{loading}
		/>
		<StatsCard
			title="Projects"
			value={loading ? '...' : stats.projects}
			icon="projects"
			color="blue"
			href="/projects"
			{loading}
		/>
		<StatsCard
			title="Task Executions"
			value={loading ? '...' : stats.task_executions}
			icon="tasks"
			color="purple"
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

	<!-- Git Changes Awaiting Review -->
	{#if !loading && stats.git_changes_awaiting_review.length > 0}
		<Card class="mt-6">
			<div class="flex items-center mb-4">
				<svg class="w-5 h-5 mr-2 text-yellow-600 dark:text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd" />
				</svg>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Git Changes Awaiting Review</h2>
				<span class="ml-2 px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded-full">
					{stats.git_changes_awaiting_review.length}
				</span>
			</div>
			<div class="space-y-3">
				{#each stats.git_changes_awaiting_review as execution}
					<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
						<div class="flex-1">
							<div class="font-medium text-gray-900 dark:text-white">{execution.task_name}</div>
							<div class="text-sm text-gray-500 dark:text-gray-400">Agent: {execution.agent}</div>
						</div>
						<div class="flex items-center space-x-2">
							<span class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded">
								{execution.status}
							</span>
							<Button 
								size="sm" 
								variant="outline"
								onclick={() => window.location.href = `/task-executions/${execution.id}`}
							>
								Review
							</Button>
						</div>
					</div>
				{/each}
			</div>
		</Card>
	{/if}

	<!-- Agents Waiting for User Input -->
	{#if !loading && stats.agents_waiting_for_input.length > 0}
		<Card class="mt-6">
			<div class="flex items-center mb-4">
				<svg class="w-5 h-5 mr-2 text-orange-600 dark:text-orange-400 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
				</svg>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Agents Waiting for Input</h2>
				<span class="ml-2 px-2 py-1 text-xs font-medium bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200 rounded-full">
					{stats.agents_waiting_for_input.length}
				</span>
			</div>
			<div class="space-y-3">
				{#each stats.agents_waiting_for_input as execution}
					<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
						<div class="flex-1">
							<div class="font-medium text-gray-900 dark:text-white">{execution.task_name}</div>
							<div class="text-sm text-gray-500 dark:text-gray-400">Agent: {execution.agent}</div>
						</div>
						<div class="flex items-center space-x-2">
							<span class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded animate-pulse">
								Waiting
							</span>
							<Button 
								size="sm" 
								variant="default"
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