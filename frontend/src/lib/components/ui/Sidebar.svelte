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
	}

	let { collapsed = false, navItems }: Props = $props();

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
</script>

<aside class="fixed left-0 top-0 z-40 h-screen transition-transform {collapsed ? 'w-16' : 'w-64'} bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
	<div class="h-full px-3 py-4 overflow-y-auto">
		<!-- Logo -->
		<div class="flex items-center {collapsed ? 'justify-center' : 'pl-2.5'} mb-5">
		<img src="https://remote-code.com/static/images/banner.svg" class="w-full" alt="Remote-Code Logo" />		</div>

		<!-- Navigation -->
		<ul class="space-y-2 font-medium">
			{#each navItems as item}
				<li>
					<a
						href={item.href}
						class="flex items-center p-2 text-gray-900 rounded-lg dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700 group {isActive(item.href) ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' : ''}"
					>
						<svg
							class="w-5 h-5 text-gray-500 transition duration-75 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-white {isActive(item.href) ? 'text-blue-700 dark:text-blue-300' : ''}"
							aria-hidden="true"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getIcon(item.icon)} />
						</svg>
						{#if !collapsed}
							<span class="ml-3">{item.label}</span>
							{#if item.badge}
								<span class="inline-flex items-center justify-center w-3 h-3 p-3 ml-3 text-sm font-medium text-blue-800 bg-blue-100 rounded-full dark:bg-blue-900 dark:text-blue-300">
									{item.badge}
								</span>
							{/if}
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	</div>
</aside>