<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';

	interface Props {
		sidebarCollapsed?: boolean;
		onToggleSidebar?: () => void;
		onToggleMobile?: () => void;
		agentsWaitingForInput?: Array<{
			id: number;
			task_name: string;
			project_name: string;
			agent: string;
		}>;
	}

	let { sidebarCollapsed = false, onToggleSidebar, onToggleMobile, agentsWaitingForInput = [] }: Props = $props();

	let showNotifications = $state(false);
	let showUserMenu = $state(false);

	let waitingCount = $derived(agentsWaitingForInput.length);

	async function handleLogout() {
		showUserMenu = false;
		await auth.logout();
		goto('/login');
	}
</script>

<header
	class="fixed top-0 right-0 z-30 h-16 bg-white/80 backdrop-blur-md border-b border-slate-200/60 transition-all duration-300
		lg:left-{sidebarCollapsed ? '16' : '64'}
		left-0"
	style="left: 0; width: 100%;"
>
	<!-- Desktop: Adjust for sidebar -->
	<div class="hidden lg:block absolute inset-0 transition-all duration-300" style="left: {sidebarCollapsed ? '4rem' : '16rem'};"></div>

	<div class="relative h-full px-4 lg:px-6">
		<div class="flex items-center justify-between h-full lg:ml-{sidebarCollapsed ? '16' : '64'}" style="margin-left: 0;">
			<!-- Left side -->
			<div class="flex items-center gap-3">
				<!-- Mobile menu button -->
				<button
					onclick={onToggleMobile}
					class="lg:hidden p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors"
					aria-label="Open menu"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				</button>

				<!-- Desktop collapse button -->
				<button
					onclick={onToggleSidebar}
					class="hidden lg:flex items-center p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors"
					style="margin-left: {sidebarCollapsed ? '4rem' : '16rem'};"
					aria-label="Toggle sidebar"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16" />
					</svg>
				</button>

				<!-- Mobile logo -->
				<div class="lg:hidden">
					<img src="https://remote-code.com/static/images/banner.svg" class="h-6" alt="Remote-Code" />
				</div>
			</div>

			<!-- Right side -->
			<div class="flex items-center gap-2">
				<!-- Search (desktop only) -->
				<div class="relative hidden md:block">
					<div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
						<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</div>
					<input
						type="text"
						class="w-64 pl-10 pr-4 py-2 text-sm text-vanna-navy bg-vanna-cream/30 border-0 rounded-xl placeholder-slate-400 focus:bg-white focus:ring-2 focus:ring-vanna-teal/30 transition-all"
						placeholder="Search..."
					/>
				</div>

				<!-- Notifications -->
				<div class="relative">
					<button
						onclick={() => showNotifications = !showNotifications}
						class="relative p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors"
						aria-label="Notifications"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						{#if waitingCount > 0}
							<span class="absolute -top-0.5 -right-0.5 w-4 h-4 text-xs font-bold text-white bg-vanna-orange rounded-full flex items-center justify-center animate-pulse">
								{waitingCount}
							</span>
						{/if}
					</button>

					{#if showNotifications}
						<div class="absolute right-0 z-50 mt-2 w-80 bg-white rounded-2xl shadow-vanna-card border border-slate-200/60 overflow-hidden">
							<div class="px-4 py-3 border-b border-slate-200 bg-vanna-cream/30">
								<h3 class="text-sm font-semibold text-vanna-navy">Agents Waiting for Input</h3>
							</div>
							<div class="max-h-64 overflow-y-auto">
								{#if agentsWaitingForInput.length === 0}
									<div class="px-4 py-6 text-center text-slate-500 text-sm">
										No agents waiting for input
									</div>
								{:else}
									{#each agentsWaitingForInput as agent}
										<a
											href="/task-executions/{agent.id}"
											class="block px-4 py-3 border-b border-slate-100 hover:bg-vanna-cream/30 transition-colors"
											onclick={() => showNotifications = false}
										>
											<div class="flex items-start gap-3">
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium text-vanna-navy">{agent.project_name}</p>
													<p class="text-sm text-slate-500 truncate">{agent.task_name}</p>
													<p class="text-xs text-slate-400 mt-1">{agent.agent}</p>
												</div>
												<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-vanna-orange/10 text-vanna-orange flex-shrink-0">
													Waiting
												</span>
											</div>
										</a>
									{/each}
								{/if}
							</div>
							<div class="px-4 py-3 border-t border-slate-200 bg-vanna-cream/20">
								<a href="/task-executions" class="text-sm font-medium text-vanna-teal hover:text-vanna-teal/80 transition-colors">
									View all task executions
								</a>
							</div>
						</div>
					{/if}
				</div>

				<!-- User Menu -->
				<div class="relative">
					<button
						onclick={() => showUserMenu = !showUserMenu}
						class="flex items-center gap-2 p-1.5 hover:bg-vanna-cream/50 rounded-xl transition-colors"
						aria-label="User menu"
					>
						<div class="w-8 h-8 bg-gradient-to-br from-vanna-teal to-vanna-navy rounded-lg flex items-center justify-center">
							<span class="text-white font-semibold text-sm">U</span>
						</div>
					</button>

					{#if showUserMenu}
						<div class="absolute right-0 z-50 mt-2 w-56 bg-white rounded-2xl shadow-vanna-card border border-slate-200/60 overflow-hidden">
							<div class="px-4 py-3 border-b border-slate-200 bg-vanna-cream/30">
								<p class="text-sm font-semibold text-vanna-navy">Remote-Code User</p>
								<p class="text-xs text-slate-500 truncate">user@remote-code.com</p>
							</div>
							<div class="py-1">
								<a href="/profile" class="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-600 hover:bg-vanna-cream/50 hover:text-vanna-navy transition-colors">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
									</svg>
									Profile
								</a>
								<a href="/settings" class="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-600 hover:bg-vanna-cream/50 hover:text-vanna-navy transition-colors">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									</svg>
									Settings
								</a>
								<div class="border-t border-slate-200 my-1"></div>
								<button
									onclick={handleLogout}
									class="flex items-center gap-3 px-4 py-2.5 text-sm text-vanna-orange hover:bg-vanna-orange/10 transition-colors w-full text-left"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
									</svg>
									Sign out
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</header>

<!-- Click outside to close dropdowns -->
<svelte:window onclick={(e) => {
	const target = e.target as HTMLElement;
	if (!target.closest('.relative')) {
		showNotifications = false;
		showUserMenu = false;
	}
}} />
