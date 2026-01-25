<script lang="ts">
	import { page } from '$app/stores';

	interface NavItem {
		label: string;
		href: string;
		icon: string;
		badge?: string | number;
		children?: NavItem[];
	}

	interface Props {
		collapsed?: boolean;
		navItems: NavItem[];
		mobileOpen?: boolean;
		onCloseMobile?: () => void;
	}

	let { collapsed = false, navItems, mobileOpen = false, onCloseMobile }: Props = $props();

	function isActive(href: string): boolean {
		return $page.url.pathname === href || ($page.url.pathname.startsWith(href) && href !== '/');
	}

	function getIcon(iconName: string): string {
		const icons: Record<string, string> = {
			dashboard: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z',
			terminal: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
			projects: 'M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4',
			tasks: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
			agents: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
			settings: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
			directories: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z'
		};
		return icons[iconName] || icons.dashboard;
	}

	function handleNavClick() {
		// Close mobile menu when navigating
		if (mobileOpen && onCloseMobile) {
			onCloseMobile();
		}
	}
</script>

<!-- Mobile backdrop -->
{#if mobileOpen}
	<div
		class="fixed inset-0 z-30 bg-vanna-navy/50 backdrop-blur-sm lg:hidden"
		onclick={onCloseMobile}
		onkeydown={(e) => e.key === 'Escape' && onCloseMobile?.()}
		role="button"
		tabindex="0"
		aria-label="Close menu"
	></div>
{/if}

<!-- Sidebar -->
<aside
	class="fixed left-0 top-0 z-40 h-screen transition-all duration-300 ease-in-out
		{collapsed ? 'w-16' : 'w-64'}
		bg-white/95 backdrop-blur-sm border-r border-slate-200 shadow-vanna-card
		lg:translate-x-0
		{mobileOpen ? 'translate-x-0' : '-translate-x-full'}"
>
	<div class="h-full flex flex-col">
		<!-- Logo Section -->
		<div class="flex items-center justify-between px-4 py-4 border-b border-slate-200/60">
			{#if collapsed}
				<div class="w-8 h-8 bg-vanna-teal rounded-lg flex items-center justify-center mx-auto">
					<span class="text-white font-bold text-sm">R</span>
				</div>
			{:else}
				<img src="https://remote-code.com/static/images/banner.svg" class="h-8" alt="Remote-Code Logo" />
			{/if}

			<!-- Mobile close button -->
			<button
				onclick={onCloseMobile}
				class="lg:hidden p-1.5 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors"
				aria-label="Close sidebar"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>

		<!-- Navigation -->
		<nav class="flex-1 px-3 py-4 overflow-y-auto">
			<ul class="space-y-1">
				{#each navItems || [] as item}
					<li>
						<a
							href={item.href}
							onclick={handleNavClick}
							class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group
								{isActive(item.href)
									? 'bg-vanna-teal text-white shadow-vanna-subtle'
									: 'text-slate-600 hover:bg-vanna-cream/50 hover:text-vanna-navy'}"
						>
							<svg
								class="w-5 h-5 flex-shrink-0 transition-colors
									{isActive(item.href) ? 'text-white' : 'text-slate-400 group-hover:text-vanna-teal'}"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getIcon(item.icon)} />
							</svg>
							{#if !collapsed}
								<span class="flex-1 font-medium text-sm">{item.label}</span>
								{#if item.badge}
									<span class="inline-flex items-center justify-center min-w-[1.25rem] h-5 px-1.5 text-xs font-semibold rounded-full
										{isActive(item.href)
											? 'bg-white/20 text-white'
											: 'bg-vanna-teal/10 text-vanna-teal'}">
										{item.badge}
									</span>
								{/if}
							{/if}
						</a>
					</li>
				{/each}
			</ul>
		</nav>

		<!-- Footer -->
		<div class="px-3 py-4 border-t border-slate-200/60">
			{#if !collapsed}
				<div class="px-3 py-2 bg-vanna-cream/30 rounded-xl">
					<p class="text-xs text-slate-500">Remote-Code v1.0.0</p>
					<p class="text-xs text-vanna-teal">Connected</p>
				</div>
			{:else}
				<div class="flex justify-center">
					<div class="w-2 h-2 bg-vanna-teal rounded-full"></div>
				</div>
			{/if}
		</div>
	</div>
</aside>
