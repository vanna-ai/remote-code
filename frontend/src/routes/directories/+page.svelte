<script>
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
</script>

<svelte:head>
	<title>Directories - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold text-vanna-navy font-serif">Directories</h1>
		<p class="mt-2 text-slate-500">Manage base directories and worktree configurations</p>
	</div>

	<!-- Coming Soon -->
	<Card class="text-center py-12">
		<div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-4">
			<svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
			</svg>
		</div>
		<h3 class="text-xl font-semibold text-vanna-navy mb-2">Coming Soon</h3>
		<p class="text-slate-500 mb-4">Directory management interface will be available in the next update</p>
		<Button href="/" variant="primary">
			Back to Dashboard
		</Button>
	</Card>
</div>
