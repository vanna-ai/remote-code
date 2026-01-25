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
		<label for={id} class="block text-sm font-medium text-vanna-navy">
			{label}
			{#if required}
				<span class="text-vanna-orange ml-1">*</span>
			{/if}
		</label>
	{/if}

	<div class="relative">
		{@render children?.()}
	</div>

	{#if description}
		<p class="text-sm text-slate-500">
			{description}
		</p>
	{/if}

	{#if error}
		<p class="text-sm text-vanna-orange flex items-center gap-1">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
			</svg>
			{error}
		</p>
	{/if}
</div>
