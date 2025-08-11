<script lang="ts">
	interface Props {
		class?: string;
		padding?: 'none' | 'sm' | 'md' | 'lg';
		shadow?: 'none' | 'sm' | 'md' | 'lg';
		border?: boolean;
		rounded?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
		onclick?: () => void;
	}

	let {
		class: className = '',
		padding = 'md',
		shadow = 'sm',
		border = true,
		rounded = 'lg',
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
		lg: 'shadow-lg'
	};

	const roundedClasses = {
		none: '',
		sm: 'rounded-sm',
		md: 'rounded-md',
		lg: 'rounded-lg',
		xl: 'rounded-xl'
	};
</script>

{#if onclick}
<div
	class="bg-white dark:bg-gray-800 {border ? 'border border-gray-200 dark:border-gray-700' : ''} {shadowClasses[shadow]} {roundedClasses[rounded]} {paddingClasses[padding]} {className}"
	onclick={onclick}
	role="button"
	tabindex="0"
>
	{@render children?.()}
</div>
{:else}
<div
	class="bg-white dark:bg-gray-800 {border ? 'border border-gray-200 dark:border-gray-700' : ''} {shadowClasses[shadow]} {roundedClasses[rounded]} {paddingClasses[padding]} {className}"
>
	{@render children?.()}
</div>
{/if}