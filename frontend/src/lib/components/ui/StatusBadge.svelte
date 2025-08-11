<script lang="ts">
	interface Props {
		status: 'todo' | 'in_progress' | 'done' | 'running' | 'completed' | 'failed' | 'pending';
		size?: 'sm' | 'md' | 'lg';
		variant?: 'dot' | 'solid' | 'outline';
		class?: string;
	}

	let {
		status,
		size = 'md',
		variant = 'dot',
		class: className = ''
	}: Props = $props();

	const statusConfig = {
		todo: {
			label: 'To Do',
			color: 'gray',
			dotColor: 'bg-gray-500',
			solidColor: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
			outlineColor: 'border-gray-300 text-gray-700 dark:border-gray-600 dark:text-gray-300'
		},
		in_progress: {
			label: 'In Progress',
			color: 'blue',
			dotColor: 'bg-blue-500',
			solidColor: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
			outlineColor: 'border-blue-300 text-blue-700 dark:border-blue-600 dark:text-blue-300'
		},
		done: {
			label: 'Done',
			color: 'green',
			dotColor: 'bg-green-500',
			solidColor: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
			outlineColor: 'border-green-300 text-green-700 dark:border-green-600 dark:text-green-300'
		},
		running: {
			label: 'Running',
			color: 'yellow',
			dotColor: 'bg-yellow-500',
			solidColor: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
			outlineColor: 'border-yellow-300 text-yellow-700 dark:border-yellow-600 dark:text-yellow-300'
		},
		completed: {
			label: 'Completed',
			color: 'green',
			dotColor: 'bg-green-500',
			solidColor: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
			outlineColor: 'border-green-300 text-green-700 dark:border-green-600 dark:text-green-300'
		},
		failed: {
			label: 'Failed',
			color: 'red',
			dotColor: 'bg-red-500',
			solidColor: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
			outlineColor: 'border-red-300 text-red-700 dark:border-red-600 dark:text-red-300'
		},
		pending: {
			label: 'Pending',
			color: 'gray',
			dotColor: 'bg-gray-400',
			solidColor: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
			outlineColor: 'border-gray-300 text-gray-600 dark:border-gray-600 dark:text-gray-400'
		},
		waiting: {
			label: 'Waiting',
			color: 'yellow',
			dotColor: 'bg-yellow-400',
			solidColor: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
			outlineColor: 'border-yellow-300 text-yellow-700 dark:border-yellow-600 dark:text-yellow-300'
		},
		starting: {
			label: 'Starting',
			color: 'blue',
			dotColor: 'bg-blue-400',
			solidColor: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
			outlineColor: 'border-blue-300 text-blue-700 dark:border-blue-600 dark:text-blue-300'
		}
	};

	const sizeClasses = {
		sm: 'px-2 py-0.5 text-xs',
		md: 'px-2.5 py-1 text-sm',
		lg: 'px-3 py-1.5 text-base'
	};

	const dotSizeClasses = {
		sm: 'w-1.5 h-1.5',
		md: 'w-2 h-2',
		lg: 'w-2.5 h-2.5'
	};

	let config = $derived(statusConfig[status] || statusConfig.pending);
	let baseClasses = $derived(`inline-flex items-center gap-1.5 rounded-full font-medium ${sizeClasses[size]}`);
	
	let variantClasses = $derived({
		dot: `${config.solidColor}`,
		solid: `${config.solidColor}`,
		outline: `border ${config.outlineColor} bg-transparent`
	}[variant]);
</script>

<span class="{baseClasses} {variantClasses} {className}">
	{#if variant === 'dot' || variant === 'solid'}
		<div class="rounded-full {dotSizeClasses[size]} {config.dotColor}"></div>
	{/if}
	{config.label}
</span>
