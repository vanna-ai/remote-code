<script lang="ts">
	interface Props {
		status: 'todo' | 'in_progress' | 'done' | 'running' | 'completed' | 'failed' | 'pending' | 'waiting' | 'starting';
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
			color: 'slate',
			dotColor: 'bg-slate-400',
			solidColor: 'bg-slate-100 text-slate-600',
			outlineColor: 'border-slate-300 text-slate-600'
		},
		in_progress: {
			label: 'In Progress',
			color: 'teal',
			dotColor: 'bg-vanna-teal',
			solidColor: 'bg-vanna-teal/10 text-vanna-teal',
			outlineColor: 'border-vanna-teal/30 text-vanna-teal'
		},
		done: {
			label: 'Done',
			color: 'teal',
			dotColor: 'bg-vanna-teal',
			solidColor: 'bg-vanna-teal/10 text-vanna-teal',
			outlineColor: 'border-vanna-teal/30 text-vanna-teal'
		},
		running: {
			label: 'Running',
			color: 'teal',
			dotColor: 'bg-vanna-teal',
			solidColor: 'bg-vanna-teal/10 text-vanna-teal',
			outlineColor: 'border-vanna-teal/30 text-vanna-teal'
		},
		completed: {
			label: 'Completed',
			color: 'teal',
			dotColor: 'bg-vanna-teal',
			solidColor: 'bg-vanna-teal/10 text-vanna-teal',
			outlineColor: 'border-vanna-teal/30 text-vanna-teal'
		},
		failed: {
			label: 'Failed',
			color: 'orange',
			dotColor: 'bg-vanna-orange',
			solidColor: 'bg-vanna-orange/10 text-vanna-orange',
			outlineColor: 'border-vanna-orange/30 text-vanna-orange'
		},
		pending: {
			label: 'Pending',
			color: 'slate',
			dotColor: 'bg-slate-400',
			solidColor: 'bg-slate-100 text-slate-400',
			outlineColor: 'border-slate-300 text-slate-400'
		},
		waiting: {
			label: 'Waiting',
			color: 'slate',
			dotColor: 'bg-slate-400',
			solidColor: 'bg-slate-100 text-slate-500',
			outlineColor: 'border-slate-300 text-slate-500'
		},
		starting: {
			label: 'Starting',
			color: 'teal',
			dotColor: 'bg-vanna-teal/70',
			solidColor: 'bg-vanna-teal/10 text-vanna-teal',
			outlineColor: 'border-vanna-teal/30 text-vanna-teal'
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
