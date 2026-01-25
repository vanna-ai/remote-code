<script lang="ts">
	interface Props {
		class?: string;
		padding?: 'none' | 'sm' | 'md' | 'lg';
		shadow?: 'none' | 'sm' | 'md' | 'lg' | 'vanna';
		border?: boolean;
		rounded?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';
		hover?: boolean;
		onclick?: () => void;
	}

	let {
		class: className = '',
		padding = 'md',
		shadow = 'vanna',
		border = true,
		rounded = '2xl',
		hover = false,
		onclick,
		children
	}: Props = $props();

	const paddingClasses = {
		none: '',
		sm: 'p-3',
		md: 'p-6',
		lg: 'p-8'
	};

	const shadowClasses = {
		none: '',
		sm: 'shadow-sm',
		md: 'shadow-md',
		lg: 'shadow-lg',
		vanna: 'shadow-vanna-card'
	};

	const roundedClasses = {
		none: '',
		sm: 'rounded-sm',
		md: 'rounded-md',
		lg: 'rounded-lg',
		xl: 'rounded-xl',
		'2xl': 'rounded-2xl'
	};

	const hoverClasses = hover ? 'hover:-translate-y-1 hover:shadow-vanna-feature transition-all duration-200 cursor-pointer' : '';
</script>

{#if onclick}
<div
	class="bg-white/80 backdrop-blur-sm {border ? 'border border-slate-200/60' : ''} {shadowClasses[shadow]} {roundedClasses[rounded]} {paddingClasses[padding]} {hoverClasses} {className}"
	onclick={onclick}
	role="button"
	tabindex="0"
>
	{@render children?.()}
</div>
{:else}
<div
	class="bg-white/80 backdrop-blur-sm {border ? 'border border-slate-200/60' : ''} {shadowClasses[shadow]} {roundedClasses[rounded]} {paddingClasses[padding]} {hover ? hoverClasses : ''} {className}"
>
	{@render children?.()}
</div>
{/if}
