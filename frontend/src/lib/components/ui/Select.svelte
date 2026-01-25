<script lang="ts">
	interface Props {
		value?: string | number;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		class?: string;
		id?: string;
		name?: string;
		size?: 'sm' | 'md' | 'lg';
		error?: boolean;
		onChange?: (event: Event) => void;
		onFocus?: (event: FocusEvent) => void;
		onBlur?: (event: FocusEvent) => void;
	}

	let {
		value = $bindable(),
		placeholder,
		disabled = false,
		required = false,
		class: className = '',
		id,
		name,
		size = 'md',
		error = false,
		onChange,
		onFocus,
		onBlur,
		children
	}: Props = $props();

	const sizeClasses = {
		sm: 'px-3 py-1.5 text-sm',
		md: 'px-3 py-2 text-sm',
		lg: 'px-4 py-3 text-base'
	};

	const baseClasses = 'w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed appearance-none bg-no-repeat bg-right bg-[length:16px_16px] pr-10';
	const normalClasses = 'border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal';
	const errorClasses = 'border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange';

	// SVG chevron down icon as data URL for the background
	const chevronIcon = "data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='m6 8 4 4 4-4'/%3e%3c/svg%3e";

	let classes = $derived(`${baseClasses} ${sizeClasses[size]} ${error ? errorClasses : normalClasses} ${className}`);
	let backgroundImage = $derived(`url("${chevronIcon}")`);
</script>

<div class="relative">
	<select
		{id}
		{name}
		{disabled}
		{required}
		bind:value
		class={classes}
		style="background-image: {backgroundImage};"
		on:change={onChange}
		on:focus={onFocus}
		on:blur={onBlur}
	>
		{#if placeholder}
			<option value="" disabled hidden>{placeholder}</option>
		{/if}
		{@render children?.()}
	</select>
</div>
