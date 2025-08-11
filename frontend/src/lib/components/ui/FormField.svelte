<script lang="ts">
	interface Props {
		label?: string;
		id?: string;
		required?: boolean;
		error?: string;
		description?: string;
		class?: string;
	}

	let {
		label,
		id,
		required = false,
		error,
		description,
		class: className = '',
		children
	}: Props = $props();
</script>

<div class="space-y-2 {className}">
	{#if label}
		<label for={id} class="block text-sm font-medium text-gray-700 dark:text-gray-300">
			{label}
			{#if required}
				<span class="text-red-500 ml-1">*</span>
			{/if}
		</label>
	{/if}
	
	<div class="relative">
		{@render children?.()}
	</div>
	
	{#if description}
		<p class="text-sm text-gray-500 dark:text-gray-400">
			{description}
		</p>
	{/if}
	
	{#if error}
		<p class="text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
			</svg>
			{error}
		</p>
	{/if}
</div>
