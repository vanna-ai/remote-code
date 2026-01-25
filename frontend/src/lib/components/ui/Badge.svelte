<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'gray' | 'magenta';
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
		primary: 'bg-vanna-teal/10 text-vanna-teal',
		secondary: 'bg-vanna-cream text-vanna-navy',
		success: 'bg-vanna-teal/10 text-vanna-teal',
		danger: 'bg-vanna-orange/10 text-vanna-orange',
		warning: 'bg-vanna-orange/10 text-vanna-orange',
		info: 'bg-vanna-teal/10 text-vanna-teal',
		gray: 'bg-slate-100 text-slate-600',
		magenta: 'bg-vanna-magenta/10 text-vanna-magenta'
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
