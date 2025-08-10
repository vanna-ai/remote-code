<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'gray';
		size?: 'sm' | 'md' | 'lg';
		class?: string;
	}

	let {
		variant = 'primary',
		size = 'md',
		class: className = '',
		children
	}: Props = $props();

	const baseClasses = 'inline-flex items-center font-medium rounded-full';

	const variantClasses = {
		primary: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
		secondary: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
		success: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
		danger: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
		warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
		info: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-300',
		gray: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
	};

	const sizeClasses = {
		sm: 'px-2 py-0.5 text-xs',
		md: 'px-2.5 py-0.5 text-sm',
		lg: 'px-3 py-1 text-sm'
	};

	const classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;
</script>

<span class={classes}>
	{@render children?.()}
</span>