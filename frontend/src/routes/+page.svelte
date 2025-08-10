<script>
	import { onMount } from 'svelte';
	import StatsCard from '$lib/components/ui/StatsCard.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	
	let stats = {
		active_sessions: 0,
		projects: 0,
		task_executions: 0,
		agents: 0
	};
	let loading = true;
	let recentActivity = [];
	
	onMount(async () => {
		await loadDashboardStats();
		await loadRecentActivity();
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

	async function loadRecentActivity() {
		try {
			const response = await fetch('/api/task-executions?limit=5');
			if (response.ok) {
				recentActivity = await response.json();
			}
		} catch (error) {
			console.error('Failed to load recent activity:', error);
		}
	}

	function getStatusColor(status) {
		switch (status?.toLowerCase()) {
			case 'completed': return 'success';
			case 'running': return 'primary';
			case 'waiting': return 'warning';
			case 'failed': return 'danger';
			case 'pending': return 'secondary';
			default: return 'secondary';
		}
	}

	function formatTimeAgo(dateString) {
		if (!dateString) return '';
		const date = new Date(dateString);
		const now = new Date();
		const diffInMinutes = Math.floor((now - date) / (1000 * 60));
		
		if (diffInMinutes < 1) return 'Just now';
		if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
		if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`;
		return `${Math.floor(diffInMinutes / 1440)}d ago`;
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
			href="/tasks"
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

	<!-- Main Content Grid -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Quick Actions -->
		<div class="lg:col-span-1">
			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Quick Actions</h3>
				<div class="space-y-3">
					<Button href="/terminal" variant="primary" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
						</svg>
						Launch Terminal
					</Button>
					<Button href="/projects" variant="secondary" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
						</svg>
						Create Project
					</Button>
					<Button href="/agents" variant="ghost" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
						</svg>
						Manage Agents
					</Button>
					<Button href="/settings" variant="ghost" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
						</svg>
						Settings
					</Button>
				</div>
			</Card>
		</div>

		<!-- Recent Activity -->
		<div class="lg:col-span-2">
			<Card>
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">Recent Activity</h3>
					<Button href="/tasks" variant="ghost" size="sm">View All</Button>
				</div>
				
				{#if recentActivity.length === 0}
					<div class="text-center py-8">
						<svg class="w-12 h-12 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
						</svg>
						<p class="text-gray-500 dark:text-gray-400">No recent activity</p>
						<p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Start by creating a project or running a task</p>
					</div>
				{:else}
					<div class="space-y-4">
						{#each recentActivity as activity}
							<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
								<div class="flex items-center space-x-3">
									<div class="flex-shrink-0">
										<div class="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center">
											<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
											</svg>
										</div>
									</div>
									<div>
										<p class="text-sm font-medium text-gray-900 dark:text-white">
											{activity.task_title || `Task ${activity.task_id}`}
										</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">
											Agent: {activity.agent_name || `Agent ${activity.agent_id}`}
										</p>
									</div>
								</div>
								<div class="flex items-center space-x-2">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-{getStatusColor(activity.status)}-100 text-{getStatusColor(activity.status)}-800 dark:bg-{getStatusColor(activity.status)}-900 dark:text-{getStatusColor(activity.status)}-300">
										{activity.status || 'Unknown'}
									</span>
									<span class="text-xs text-gray-400 dark:text-gray-500">
										{formatTimeAgo(activity.created_at)}
									</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</Card>
		</div>
	</div>

	<!-- System Status -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<Card>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">System Status</h3>
			<div class="space-y-3">
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">API Server</span>
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
						<div class="w-1.5 h-1.5 bg-green-400 rounded-full mr-1.5"></div>
						Online
					</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Database</span>
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
						<div class="w-1.5 h-1.5 bg-green-400 rounded-full mr-1.5"></div>
						Connected
					</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Task Queue</span>
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
						<div class="w-1.5 h-1.5 bg-green-400 rounded-full mr-1.5"></div>
						Active
					</span>
				</div>
			</div>
		</Card>

		<Card>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Quick Links</h3>
			<div class="space-y-2">
				<a href="/directories" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
					📁 Manage Directories
				</a>
				<a href="/terminal" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
					💻 Open Terminal
				</a>
				<a href="/projects" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
					🚀 View Projects
				</a>
				<a href="/agents" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
					🤖 Configure Agents
				</a>
			</div>
		</Card>
	</div>
</div>