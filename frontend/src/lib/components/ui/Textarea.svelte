<script lang="ts">
	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		readonly?: boolean;
		class?: string;
		id?: string;
		name?: string;
		rows?: number;
		cols?: number;
		resize?: 'none' | 'both' | 'horizontal' | 'vertical';
		error?: boolean;
		onInput?: (event: Event) => void;
		onChange?: (event: Event) => void;
		onFocus?: (event: FocusEvent) => void;
		onBlur?: (event: FocusEvent) => void;
	}

	let {
		value = $bindable(),
		placeholder,
		disabled = false,
		required = false,
		readonly = false,
		class: className = '',
		id,
		name,
		rows = 3,
		cols,
		resize = 'vertical',
		error = false,
		onInput,
		onChange,
		onFocus,
		onBlur
	}: Props = $props();

	const resizeClasses = {
		none: 'resize-none',
		both: 'resize',
		horizontal: 'resize-x',
		vertical: 'resize-y'
	};

	const baseClasses = 'w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed px-3 py-2 text-sm';
	const normalClasses = 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-blue-500 dark:focus:border-blue-400 dark:focus:ring-blue-400';
	const errorClasses = 'border-red-300 dark:border-red-600 bg-red-50 dark:bg-red-900/20 text-red-900 dark:text-red-100 focus:border-red-500 focus:ring-red-500';

	let classes = $derived(`${baseClasses} ${resizeClasses[resize]} ${error ? errorClasses : normalClasses} ${className}`);
</script>

<textarea
	{id}
	{name}
	{placeholder}
	{disabled}
	{required}
	{readonly}
	{rows}
	{cols}
	bind:value
	class={classes}
	on:input={onInput}
	on:change={onChange}
	on:focus={onFocus}
	on:blur={onBlur}
></textarea>
