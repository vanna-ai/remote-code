<script lang="ts">
	interface Props {
		title: string;
		value: string | number;
		icon: string;
		color?: 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'yellow';
		trend?: {
			value: number;
			isPositive: boolean;
		};
		href?: string;
		loading?: boolean;
	}

	let {
		title,
		value,
		icon,
		color = 'blue',
		trend,
		href,
		loading = false
	}: Props = $props();

	const colorClasses = {
		blue: {
			bg: 'bg-blue-500',
			text: 'text-blue-600',
			lightBg: 'bg-blue-50 dark:bg-blue-900/20'
		},
		green: {
			bg: 'bg-green-500',
			text: 'text-green-600',
			lightBg: 'bg-green-50 dark:bg-green-900/20'
		},
		purple: {
			bg: 'bg-purple-500',
			text: 'text-purple-600',
			lightBg: 'bg-purple-50 dark:bg-purple-900/20'
		},
		orange: {
			bg: 'bg-orange-500',
			text: 'text-orange-600',
			lightBg: 'bg-orange-50 dark:bg-orange-900/20'
		},
		red: {
			bg: 'bg-red-500',
			text: 'text-red-600',
			lightBg: 'bg-red-50 dark:bg-red-900/20'
		},
		yellow: {
			bg: 'bg-yellow-500',
			text: 'text-yellow-600',
			lightBg: 'bg-yellow-50 dark:bg-yellow-900/20'
		}
	};

	function getIcon(iconName: string): string {
		const icons: Record<string, string> = {
			terminal: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
			projects: 'M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4',
			tasks: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
			agents: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
			users: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z',
			chart: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		};
		return icons[iconName] || icons.chart;
	}

	const Component = href ? 'a' : 'div';
</script>

<svelte:element
	this={Component}
	{href}
	class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 shadow-sm hover:shadow-md transition-all duration-200 {href ? 'hover:border-gray-300 dark:hover:border-gray-600 cursor-pointer' : ''}"
>
	<div class="flex items-center">
		<div class="flex-shrink-0">
			<div class="w-12 h-12 {colorClasses[color].lightBg} rounded-lg flex items-center justify-center">
				<svg class="w-6 h-6 {colorClasses[color].text}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getIcon(icon)} />
				</svg>
			</div>
		</div>
		<div class="ml-4 flex-1">
			<div class="flex items-center justify-between">
				<div>
					<p class="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-white">
						{#if loading}
							<div class="animate-pulse bg-gray-300 dark:bg-gray-600 h-8 w-16 rounded"></div>
						{:else}
							{value}
						{/if}
					</p>
				</div>
				{#if trend}
					<div class="flex items-center text-sm {trend.isPositive ? 'text-green-600' : 'text-red-600'}">
						<svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
							{#if trend.isPositive}
								<path fill-rule="evenodd" d="M3.293 9.707a1 1 0 010-1.414l6-6a1 1 0 011.414 0l6 6a1 1 0 01-1.414 1.414L11 5.414V17a1 1 0 11-2 0V5.414L4.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
							{:else}
								<path fill-rule="evenodd" d="M16.707 10.293a1 1 0 010 1.414l-6 6a1 1 0 01-1.414 0l-6-6a1 1 0 111.414-1.414L9 14.586V3a1 1 0 012 0v11.586l4.293-4.293a1 1 0 011.414 0z" clip-rule="evenodd" />
							{/if}
						</svg>
						{Math.abs(trend.value)}%
					</div>
				{/if}
			</div>
		</div>
	</div>
</svelte:element>