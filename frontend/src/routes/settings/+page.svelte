<script>
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	
	const breadcrumbSegments = [
		{ label: "Remote-Code", href: "/", icon: "banner" },
		{ label: "Settings", href: "/settings" }
	];
</script>

<svelte:head>
	<title>Settings - Remote-Code</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Breadcrumb -->
		<Breadcrumb segments={breadcrumbSegments} />
		
		<!-- Header -->
		<div class="mb-6">
			<div class="flex items-center gap-4">
				<div class="w-12 h-12 bg-gray-500 rounded-lg flex items-center justify-center">
					<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
					</svg>
				</div>
				<div>
					<h1 class="text-3xl font-bold text-gray-400 mb-1">Settings</h1>
					<p class="text-gray-300">Configure application settings and preferences</p>
				</div>
			</div>
		</div>

		<!-- Coming Soon -->
		<div class="text-center py-12">
			<div class="w-16 h-16 bg-gray-700 rounded-lg flex items-center justify-center mx-auto mb-4">
				<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
				</svg>
			</div>
			<h3 class="text-xl font-semibold text-gray-300 mb-2">Coming Soon</h3>
			<p class="text-gray-400">Settings interface will be available in the next update</p>
		</div>
	</div>
</div>