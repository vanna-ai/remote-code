<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'ghost';
		size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
		disabled?: boolean;
		loading?: boolean;
		class?: string;
		href?: string;
		type?: 'button' | 'submit' | 'reset';
		onclick?: () => void;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		class: className = '',
		href,
		type = 'button',
		onclick,
		children
	}: Props = $props();

	const baseClasses = 'inline-flex items-center justify-center font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';

	const variantClasses = {
		primary: 'bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500',
		secondary: 'bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500',
		success: 'bg-green-600 hover:bg-green-700 text-white focus:ring-green-500',
		danger: 'bg-red-600 hover:bg-red-700 text-white focus:ring-red-500',
		warning: 'bg-yellow-600 hover:bg-yellow-700 text-white focus:ring-yellow-500',
		info: 'bg-cyan-600 hover:bg-cyan-700 text-white focus:ring-cyan-500',
		ghost: 'bg-transparent hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300 focus:ring-gray-500'
	};

	const sizeClasses = {
		xs: 'px-2.5 py-1.5 text-xs',
		sm: 'px-3 py-2 text-sm',
		md: 'px-4 py-2 text-sm',
		lg: 'px-4 py-2 text-base',
		xl: 'px-6 py-3 text-base'
	};

	const classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;
</script>

{#if href}
	<a {href} class={classes}>
		{#if loading}
			<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
		{/if}
		{@render children?.()}
	</a>
{:else}
	<button {type} {disabled} {onclick} class={classes}>
		{#if loading}
			<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
		{/if}
		{@render children?.()}
	</button>
{/if}