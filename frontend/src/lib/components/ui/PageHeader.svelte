<script lang="ts">
	interface Props {
		title: string;
		subtitle?: string;
		class?: string;
	}

	let {
		title,
		subtitle,
		class: className = '',
		children
	}: Props = $props();
</script>

<div class="border-b border-slate-200 pb-6 mb-8 {className}">
	<div class="flex items-center justify-between">
		<div class="min-w-0 flex-1">
			<h1 class="text-2xl font-bold text-vanna-navy font-serif sm:text-3xl">
				{title}
			</h1>
			{#if subtitle}
				<p class="mt-2 text-sm text-slate-500">
					{subtitle}
				</p>
			{/if}
		</div>

		{#if children}
			<div class="flex items-center gap-3 ml-6">
				{@render children?.()}
			</div>
		{/if}
	</div>
</div>
