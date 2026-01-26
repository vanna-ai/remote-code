<script lang="ts">
	import { getContext } from 'svelte';

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

	const formFieldContext = getContext<{ error: string; describedBy: string } | undefined>('formField');
	let hasError = $derived(error || !!formFieldContext?.error);
	let computedDescribedBy = $derived(formFieldContext?.describedBy);

	const resizeClasses = {
		none: 'resize-none',
		both: 'resize',
		horizontal: 'resize-x',
		vertical: 'resize-y'
	};

	const baseClasses = 'w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed px-3 py-2 text-sm';
	const normalClasses = 'border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal placeholder-slate-400';
	const errorClasses = 'border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange';

	let classes = $derived(`${baseClasses} ${resizeClasses[resize]} ${hasError ? errorClasses : normalClasses} ${className}`);
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
	aria-invalid={hasError || undefined}
	aria-describedby={computedDescribedBy || undefined}
	oninput={onInput}
	onchange={onChange}
	onfocus={onFocus}
	onblur={onBlur}
></textarea>
