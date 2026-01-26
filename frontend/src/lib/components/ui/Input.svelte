<script lang="ts">
	import { getContext } from 'svelte';

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

	const formFieldContext = getContext<{ error: string; describedBy: string } | undefined>('formField');
	let hasError = $derived(error || !!formFieldContext?.error);
	let computedDescribedBy = $derived(formFieldContext?.describedBy);

	const sizeClasses = {
		sm: 'px-3 py-1.5 text-sm',
		md: 'px-3 py-2 text-sm',
		lg: 'px-4 py-3 text-base'
	};

	const baseClasses = 'w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';
	const normalClasses = 'border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal placeholder-slate-400';
	const errorClasses = 'border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange';

	let classes = $derived(`${baseClasses} ${sizeClasses[size]} ${hasError ? errorClasses : normalClasses} ${className}`);
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
	aria-invalid={hasError || undefined}
	aria-describedby={computedDescribedBy || undefined}
	oninput={onInput}
	onchange={onChange}
	onfocus={onFocus}
	onblur={onBlur}
/>
