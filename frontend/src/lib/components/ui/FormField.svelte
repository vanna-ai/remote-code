<script lang="ts">
	import { setContext } from 'svelte';

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

	const errorId = id ? `${id}-error` : undefined;
	const descriptionId = id ? `${id}-description` : undefined;

	let describedBy = $derived(
		[error && errorId, description && descriptionId].filter(Boolean).join(' ') || undefined
	);

	setContext('formField', {
		get error() { return error; },
		get describedBy() { return describedBy; }
	});
</script>

<div class="space-y-2 {className}">
	{#if label}
		<label for={id} class="block text-sm font-medium text-vanna-navy">
			{label}
			{#if required}
				<span class="text-vanna-orange ml-1" aria-hidden="true">*</span>
				<span class="sr-only">(required)</span>
			{/if}
		</label>
	{/if}

	<div class="relative">
		{@render children?.()}
	</div>

	{#if description}
		<p id={descriptionId} class="text-sm text-slate-500">
			{description}
		</p>
	{/if}

	{#if error}
		<p id={errorId} class="text-sm text-vanna-orange flex items-center gap-1" role="alert">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
			</svg>
			{error}
		</p>
	{/if}
</div>
