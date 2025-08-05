<script>
	// Home page with card-based navigation
</script>

<svelte:head>
	<title>Remote-Code Dashboard</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white">
	<div class="container mx-auto p-6">
		<!-- Header -->
		<div class="mb-8">
			<div class="flex items-center gap-4 mb-6">
				<img 
					src="https://remote-code.com/static/images/banner.svg" 
					alt="Remote-Code Logo" 
					class="h-16 w-auto"
				/>
				<div>
					<h1 class="text-4xl font-bold text-green-400 mb-2">Remote-Code</h1>
					<p class="text-gray-300 text-lg">Development environment management platform</p>
				</div>
			</div>
		</div>

		<!-- Navigation Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
			<!-- Terminal Card -->
			<a 
				href="/terminal" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-green-400 transition-all duration-300 hover:shadow-xl hover:shadow-green-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-green-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-green-400 transition-colors">Terminal</h3>
				</div>
				<p class="text-gray-400 mb-4">Interactive terminal with tmux session management</p>
				<div class="flex items-center text-green-400 text-sm font-medium">
					<span>Launch Terminal</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>

			<!-- Projects Card -->
			<a 
				href="/projects" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-blue-400 transition-all duration-300 hover:shadow-xl hover:shadow-blue-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-blue-400 transition-colors">Projects</h3>
				</div>
				<p class="text-gray-400 mb-4">Manage your development projects and repositories</p>
				<div class="flex items-center text-blue-400 text-sm font-medium">
					<span>View Projects</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>

			<!-- Tasks Card -->
			<a 
				href="/tasks" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-purple-400 transition-all duration-300 hover:shadow-xl hover:shadow-purple-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-purple-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-purple-400 transition-colors">Tasks</h3>
				</div>
				<p class="text-gray-400 mb-4">Track and manage development tasks and workflows</p>
				<div class="flex items-center text-purple-400 text-sm font-medium">
					<span>View Tasks</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>

			<!-- Agents Card -->
			<a 
				href="/agents" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-orange-400 transition-all duration-300 hover:shadow-xl hover:shadow-orange-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-orange-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-orange-400 transition-colors">Agents</h3>
				</div>
				<p class="text-gray-400 mb-4">Configure and manage AI development agents</p>
				<div class="flex items-center text-orange-400 text-sm font-medium">
					<span>Manage Agents</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>

			<!-- Base Directories Card -->
			<a 
				href="/directories" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-yellow-400 transition-all duration-300 hover:shadow-xl hover:shadow-yellow-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-yellow-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-yellow-400 transition-colors">Directories</h3>
				</div>
				<p class="text-gray-400 mb-4">Manage base directories and worktree configurations</p>
				<div class="flex items-center text-yellow-400 text-sm font-medium">
					<span>Configure Directories</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>

			<!-- Settings Card -->
			<a 
				href="/settings" 
				class="group bg-gray-800 rounded-xl border border-gray-700 p-6 hover:border-gray-400 transition-all duration-300 hover:shadow-xl hover:shadow-gray-400/20"
			>
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 bg-gray-500 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white group-hover:text-gray-400 transition-colors">Settings</h3>
				</div>
				<p class="text-gray-400 mb-4">Configure application settings and preferences</p>
				<div class="flex items-center text-gray-400 text-sm font-medium">
					<span>Open Settings</span>
					<svg class="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
					</svg>
				</div>
			</a>
		</div>

		<!-- Quick Stats -->
		<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
			<div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
				<div class="text-2xl font-bold text-green-400">0</div>
				<div class="text-sm text-gray-400">Active Sessions</div>
			</div>
			<div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
				<div class="text-2xl font-bold text-blue-400">0</div>
				<div class="text-sm text-gray-400">Projects</div>
			</div>
			<div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
				<div class="text-2xl font-bold text-purple-400">0</div>
				<div class="text-sm text-gray-400">Tasks</div>
			</div>
			<div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
				<div class="text-2xl font-bold text-orange-400">0</div>
				<div class="text-sm text-gray-400">Agents</div>
			</div>
		</div>
	</div>
</div>