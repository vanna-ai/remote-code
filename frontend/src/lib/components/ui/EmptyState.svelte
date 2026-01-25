<script lang="ts">
	interface Props {
		title: string;
		description?: string;
		icon?: 'tasks' | 'folder' | 'agents' | 'search' | 'generic';
		class?: string;
	}

	let {
		title,
		description,
		icon = 'generic',
		class: className = '',
		children
	}: Props = $props();

	const icons = {
		tasks: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>`,
		folder: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>`,
		agents: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>`,
		search: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>`,
		generic: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>`
	};
</script>

<div class="text-center py-12 {className}">
	<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-6">
		<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			{@html icons[icon]}
		</svg>
	</div>

	<h3 class="text-lg font-semibold text-vanna-navy mb-2">
		{title}
	</h3>

	{#if description}
		<p class="text-slate-500 mb-6 max-w-md mx-auto">
			{description}
		</p>
	{/if}

	{#if children}
		<div class="flex flex-col sm:flex-row gap-3 justify-center items-center">
			{@render children?.()}
		</div>
	{/if}
</div>
