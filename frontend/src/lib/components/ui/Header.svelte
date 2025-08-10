<script lang="ts">
	import DarkModeToggle from './DarkModeToggle.svelte';

	interface Props {
		sidebarCollapsed?: boolean;
		onToggleSidebar?: () => void;
	}

	let { sidebarCollapsed = false, onToggleSidebar }: Props = $props();

	let showNotifications = $state(false);
	let showUserMenu = $state(false);

	// Mock notifications data
	const notifications = [
		{
			id: 1,
			title: 'Task Execution Completed',
			message: 'Your task "Setup Development Environment" has completed successfully.',
			time: '5 min ago',
			read: false
		},
		{
			id: 2,
			title: 'New Project Created',
			message: 'Project "E-commerce Platform" has been created.',
			time: '1 hr ago',
			read: false
		},
		{
			id: 3,
			title: 'Agent Status Update',
			message: 'Agent "Development Assistant" is now online.',
			time: '2 hrs ago',
			read: true
		}
	];

	const unreadCount = notifications.filter(n => !n.read).length;
</script>

<header class="fixed top-0 z-30 w-full bg-white border-b border-gray-200 dark:bg-gray-800 dark:border-gray-700" style="left: {sidebarCollapsed ? '4rem' : '16rem'}; width: calc(100% - {sidebarCollapsed ? '4rem' : '16rem'});">
	<div class="px-3 py-3 lg:px-5 lg:pl-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center justify-start">
				<button
					onclick={onToggleSidebar}
					class="inline-flex items-center p-2 text-sm text-gray-500 rounded-lg hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:text-gray-400 dark:hover:bg-gray-700 dark:focus:ring-gray-600"
				>
					<svg class="w-6 h-6" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
						<path clip-rule="evenodd" fill-rule="evenodd" d="M2 4.75A.75.75 0 012.75 4h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 4.75zm0 10.5a.75.75 0 01.75-.75h7.5a.75.75 0 010 1.5h-7.5a.75.75 0 01-.75-.75zM2 10a.75.75 0 01.75-.75h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 10z"></path>
					</svg>
				</button>
			</div>
			<div class="flex items-center">
				<div class="flex items-center ml-3">
					<!-- Search -->
					<div class="relative hidden md:block">
						<div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
							<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
								<path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"></path>
							</svg>
						</div>
						<input type="text" class="block w-full p-2 pl-10 text-sm text-gray-900 border border-gray-300 rounded-lg bg-gray-50 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="Search...">
					</div>

					<!-- Dark Mode Toggle -->
					<DarkModeToggle />

					<!-- Notifications -->
					<div class="relative ml-3">
						<button
							onclick={() => showNotifications = !showNotifications}
							class="relative flex items-center p-2 text-sm text-gray-500 rounded-lg hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:text-gray-400 dark:hover:bg-gray-700 dark:focus:ring-gray-600"
						>
							<svg class="w-6 h-6" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
								<path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z"></path>
							</svg>
							{#if unreadCount > 0}
								<span class="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-red-100 transform translate-x-1/2 -translate-y-1/2 bg-red-600 rounded-full">{unreadCount}</span>
							{/if}
						</button>

						{#if showNotifications}
							<div class="absolute right-0 z-50 mt-2 w-80 bg-white rounded-lg shadow-lg dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
								<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
									<h3 class="text-sm font-medium text-gray-900 dark:text-white">Notifications</h3>
								</div>
								<div class="max-h-64 overflow-y-auto">
									{#each notifications as notification}
										<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 {!notification.read ? 'bg-blue-50 dark:bg-blue-900/20' : ''}">
											<div class="flex items-start">
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium text-gray-900 dark:text-white">{notification.title}</p>
													<p class="text-sm text-gray-500 dark:text-gray-400">{notification.message}</p>
													<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{notification.time}</p>
												</div>
												{#if !notification.read}
													<div class="w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
												{/if}
											</div>
										</div>
									{/each}
								</div>
								<div class="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
									<a href="#" class="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300">View all notifications</a>
								</div>
							</div>
						{/if}
					</div>

					<!-- User Menu -->
					<div class="relative ml-3">
						<button
							onclick={() => showUserMenu = !showUserMenu}
							class="flex items-center p-2 text-sm bg-gray-800 rounded-full focus:outline-none focus:ring-2 focus:ring-gray-300 dark:focus:ring-gray-600"
						>
							<img class="w-8 h-8 rounded-full" src="https://flowbite.com/docs/images/people/profile-picture-5.jpg" alt="user photo">
						</button>

						{#if showUserMenu}
							<div class="absolute right-0 z-50 mt-2 w-48 bg-white rounded-lg shadow-lg dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
								<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
									<p class="text-sm text-gray-900 dark:text-white">Remote-Code User</p>
									<p class="text-sm text-gray-500 dark:text-gray-400 truncate">user@remote-code.com</p>
								</div>
								<ul class="py-2">
									<li>
										<a href="/profile" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 dark:text-gray-200 dark:hover:text-white">Profile</a>
									</li>
									<li>
										<a href="/settings" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 dark:text-gray-200 dark:hover:text-white">Settings</a>
									</li>
									<li>
										<a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 dark:text-gray-200 dark:hover:text-white">Sign out</a>
									</li>
								</ul>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	</div>
</header>

<!-- Click outside to close dropdowns -->
<svelte:window onclick={(e) => {
	if (!e.target.closest('.relative')) {
		showNotifications = false;
		showUserMenu = false;
	}
}} />