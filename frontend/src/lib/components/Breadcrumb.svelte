<script>
	export let segments = [];
	
	// segments should be an array of objects like:
	// [
	//   { label: "Home", href: "/", icon: "banner" }, // supports icon: "banner" for Remote-Code logo
	//   { label: "Projects", href: "/projects" },
	//   { label: "My Project", href: "/projects/1" }
	// ]
	// The last segment is automatically treated as current page (no link)
</script>

<nav class="flex items-center space-x-2 text-sm text-gray-400 mb-6" aria-label="Breadcrumb">
	{#each segments as segment, index}
		{#if index < segments.length - 1}
			<!-- Clickable segment -->
			<a 
				href={segment.href} 
				class="hover:text-white transition-colors flex items-center gap-2"
			>
				{#if segment.icon === "banner"}
					<img 
						src="https://remote-code.com/static/images/banner.svg" 
						alt="Remote-Code Logo" 
						class="h-10 w-auto"
					/>
				{:else if index === 0}
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H9L7 3H3a2 2 0 00-2 2v2z"/>
					</svg>
				{/if}
				{#if segment.icon !== "banner"}{segment.label}{/if}
			</a>
			<!-- Separator -->
			<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
			</svg>
		{:else}
			<!-- Current page (not clickable) -->
			<span class="text-white font-medium flex items-center gap-2">
				{#if segment.icon === "banner"}
					<img 
						src="https://remote-code.com/static/images/banner.svg" 
						alt="Remote-Code Logo" 
						class="h-10 w-auto"
					/>
				{:else if segments.length === 1}
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H9L7 3H3a2 2 0 00-2 2v2z"/>
					</svg>
				{/if}
				{#if segment.icon !== "banner"}{segment.label}{/if}
			</span>
		{/if}
	{/each}
</nav>