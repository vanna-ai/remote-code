<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/components/ui/Sidebar.svelte';
	import Header from '$lib/components/ui/Header.svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let { children } = $props();
	let sidebarCollapsed = $state(false);
	let stats = $state({
		active_sessions: 0,
		projects: 0,
		task_executions: 0,
		agents: 0
	});

	const navItems = [
		{ label: 'Dashboard', href: '/', icon: 'dashboard' },
		{ label: 'Terminal', href: '/terminal', icon: 'terminal', badge: stats.active_sessions || undefined },
		{ label: 'Projects', href: '/projects', icon: 'projects', badge: stats.projects || undefined },
		{ label: 'Task Executions', href: '/task-executions', icon: 'tasks', badge: stats.task_executions || undefined },
		{ label: 'Agents', href: '/agents', icon: 'agents', badge: stats.agents || undefined },
		{ label: 'Directories', href: '/directories', icon: 'directories' },
		{ label: 'Settings', href: '/settings', icon: 'settings' }
	];

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
		}
	}

	function toggleSidebar() {
		sidebarCollapsed = !sidebarCollapsed;
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="bg-gray-50 dark:bg-gray-900 min-h-screen">
	<Sidebar {navItems} collapsed={sidebarCollapsed} />
	<Header {sidebarCollapsed} onToggleSidebar={toggleSidebar} />
	
	<main class="transition-all duration-300" style="margin-left: {sidebarCollapsed ? '4rem' : '16rem'}; margin-top: 4rem;">
		<div class="p-4">
			{@render children?.()}
		</div>
	</main>
</div>
