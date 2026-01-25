<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'ghost';
		size?: 'xs' | 'sm' | 'md' | 'lg';
		disabled?: boolean;
		loading?: boolean;
		class?: string;
		type?: 'button' | 'submit' | 'reset';
		title?: string;
		onclick?: () => void;
	}

	let {
		variant = 'ghost',
		size = 'md',
		disabled = false,
		loading = false,
		class: className = '',
		type = 'button',
		title,
		onclick,
		children
	}: Props = $props();

	const baseClasses = 'inline-flex items-center justify-center font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';

	const variantClasses = {
		primary: 'bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal',
		secondary: 'bg-vanna-cream hover:bg-vanna-cream/80 text-vanna-navy focus:ring-vanna-teal',
		success: 'bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal',
		danger: 'bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange',
		warning: 'bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange',
		info: 'bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal',
		ghost: 'bg-transparent hover:bg-vanna-cream/50 text-vanna-navy focus:ring-vanna-teal'
	};

	const sizeClasses = {
		xs: 'p-1',
		sm: 'p-1.5',
		md: 'p-2',
		lg: 'p-2.5'
	};

	const iconSizeClasses = {
		xs: 'w-3 h-3',
		sm: 'w-4 h-4',
		md: 'w-5 h-5',
		lg: 'w-6 h-6'
	};

	let classes = $derived(`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`);
</script>

<button
	{type}
	{disabled}
	{title}
	{onclick}
	class={classes}
	aria-label={title}
>
	{#if loading}
		<svg class="animate-spin {iconSizeClasses[size]}" fill="none" viewBox="0 0 24 24">
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
			<path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
		</svg>
	{:else}
		<div class="{iconSizeClasses[size]}">
			{@render children?.()}
		</div>
	{/if}
</button>
