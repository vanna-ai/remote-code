<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/components/ui/Sidebar.svelte';
	import Header from '$lib/components/ui/Header.svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let { children } = $props();
	let sidebarCollapsed = $state(false);
	let mobileMenuOpen = $state(false);
	let stats = $state({
		active_sessions: 0,
		projects: 0,
		task_executions: 0,
		agents: 0,
		agents_waiting_for_input: [] as Array<{
			id: number;
			task_name: string;
			project_name: string;
			agent: string;
		}>
	});

	// Use $derived so navItems updates when stats changes
	let navItems = $derived([
		{ label: 'Dashboard', href: '/', icon: 'dashboard' },
		{ label: 'Terminal', href: '/terminal', icon: 'terminal', badge: stats?.active_sessions || undefined },
		{ label: 'Projects', href: '/projects', icon: 'projects', badge: stats?.projects || undefined },
		{ label: 'Task Executions', href: '/task-executions', icon: 'tasks', badge: stats?.task_executions || undefined },
		{ label: 'Agents', href: '/agents', icon: 'agents', badge: stats?.agents || undefined },
		{ label: 'Directories', href: '/directories', icon: 'directories' },
		{ label: 'Settings', href: '/settings', icon: 'settings' }
	]);

	onMount(async () => {
		await loadDashboardStats();
		const interval = setInterval(loadDashboardStats, 10000);
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
		}
	}

	function toggleSidebar() {
		sidebarCollapsed = !sidebarCollapsed;
	}

	function toggleMobileMenu() {
		mobileMenuOpen = !mobileMenuOpen;
	}

	function closeMobileMenu() {
		mobileMenuOpen = false;
	}

	// Close mobile menu on route change
	$effect(() => {
		$page.url.pathname;
		mobileMenuOpen = false;
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<!-- Main layout with Vanna design system background -->
<div class="min-h-screen bg-gradient-to-b from-vanna-cream via-white to-vanna-cream relative overflow-hidden">
	<!-- Decorative background elements -->
	<div class="fixed inset-0 pointer-events-none">
		<!-- Teal radial glow top-left -->
		<div class="absolute -top-40 -left-40 w-[600px] h-[600px] rounded-full bg-vanna-teal/10 blur-[180px]"></div>
		<!-- Cream radial glow bottom-right -->
		<div class="absolute -bottom-40 -right-40 w-[500px] h-[500px] rounded-full bg-vanna-cream blur-[150px]"></div>
		<!-- Subtle magenta accent -->
		<div class="absolute top-1/2 right-0 w-[300px] h-[300px] rounded-full bg-vanna-magenta/5 blur-[120px]"></div>
		<!-- Dot pattern overlay -->
		<div class="absolute inset-0 dot-pattern opacity-30"></div>
	</div>

	<!-- Content -->
	<div class="relative z-10">
		<Sidebar
			{navItems}
			collapsed={sidebarCollapsed}
			mobileOpen={mobileMenuOpen}
			onCloseMobile={closeMobileMenu}
		/>
		<Header
			{sidebarCollapsed}
			onToggleSidebar={toggleSidebar}
			onToggleMobile={toggleMobileMenu}
			agentsWaitingForInput={stats.agents_waiting_for_input}
		/>

		<!-- Main content area -->
		<main
			class="transition-all duration-300 pt-16 lg:ml-64"
			class:lg:ml-16={sidebarCollapsed}
		>
			<div class="p-4 lg:p-6">
				{@render children?.()}
			</div>
		</main>
	</div>
</div>
