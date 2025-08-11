<script lang="ts">
	interface Props {
		type?: 'text' | 'email' | 'password' | 'number' | 'url' | 'tel' | 'search';
		value?: string | number;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		readonly?: boolean;
		class?: string;
		id?: string;
		name?: string;
		autocomplete?: string;
		min?: number;
		max?: number;
		step?: number;
		pattern?: string;
		size?: 'sm' | 'md' | 'lg';
		error?: boolean;
		onInput?: (event: Event) => void;
		onChange?: (event: Event) => void;
		onFocus?: (event: FocusEvent) => void;
		onBlur?: (event: FocusEvent) => void;
	}

	let {
		type = 'text',
		value = $bindable(),
		placeholder,
		disabled = false,
		required = false,
		readonly = false,
		class: className = '',
		id,
		name,
		autocomplete,
		min,
		max,
		step,
		pattern,
		size = 'md',
		error = false,
		onInput,
		onChange,
		onFocus,
		onBlur
	}: Props = $props();

	const sizeClasses = {
		sm: 'px-3 py-1.5 text-sm',
		md: 'px-3 py-2 text-sm',
		lg: 'px-4 py-3 text-base'
	};

	const baseClasses = 'w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';
	const normalClasses = 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-blue-500 dark:focus:border-blue-400 dark:focus:ring-blue-400';
	const errorClasses = 'border-red-300 dark:border-red-600 bg-red-50 dark:bg-red-900/20 text-red-900 dark:text-red-100 focus:border-red-500 focus:ring-red-500';

	let classes = $derived(`${baseClasses} ${sizeClasses[size]} ${error ? errorClasses : normalClasses} ${className}`);
</script>

<input
	{type}
	{id}
	{name}
	{placeholder}
	{disabled}
	{required}
	{readonly}
	{autocomplete}
	{min}
	{max}
	{step}
	{pattern}
	bind:value
	class={classes}
	on:input={onInput}
	on:change={onChange}
	on:focus={onFocus}
	on:blur={onBlur}
/>
