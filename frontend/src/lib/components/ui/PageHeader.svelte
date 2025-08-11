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

<div class="border-b border-gray-200 dark:border-gray-700 pb-6 mb-8 {className}">
	<div class="flex items-center justify-between">
		<div class="min-w-0 flex-1">
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
				{title}
			</h1>
			{#if subtitle}
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
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
