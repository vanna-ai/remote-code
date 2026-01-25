<script lang="ts">
	import Badge from './Badge.svelte';

	interface Props {
		task: {
			id: number;
			task_title?: string;
			task_id?: number;
			status?: string;
			agent_name?: string;
			agent_id?: number;
			created_at?: string;
		};
		onDelete?: (id: number) => void;
		onStatusChange?: (id: number, newStatus: string) => void;
	}

	let { task, onDelete, onStatusChange }: Props = $props();

	function getStatusColor(status?: string) {
		switch (status?.toLowerCase()) {
			case 'completed': return 'success';
			case 'running': return 'primary';
			case 'waiting': return 'warning';
			case 'failed': return 'danger';
			case 'pending': return 'secondary';
			default: return 'secondary';
		}
	}

	function formatTimeAgo(dateString?: string) {
		if (!dateString) return '';
		const date = new Date(dateString);
		const now = new Date();
		const diffInMinutes = Math.floor((now - date) / (1000 * 60));

		if (diffInMinutes < 1) return 'Just now';
		if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
		if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`;
		return `${Math.floor(diffInMinutes / 1440)}d ago`;
	}

	let showDropdown = $state(false);
</script>

<div class="bg-white/80 backdrop-blur-sm rounded-2xl border border-slate-200/60 p-4 shadow-vanna-card hover:shadow-vanna-feature hover:-translate-y-1 transition-all duration-200 group">
	<div class="flex items-start justify-between mb-3">
		<h4 class="text-sm font-medium text-vanna-navy line-clamp-2">
			{task.task_title || `Task ${task.task_id}`}
		</h4>
		<div class="relative">
			<button
				onclick={() => showDropdown = !showDropdown}
				class="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded hover:bg-vanna-cream/50"
			>
				<svg class="w-4 h-4 text-slate-400" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"/>
				</svg>
			</button>

			{#if showDropdown}
				<div class="absolute right-0 mt-1 w-48 bg-white rounded-xl shadow-vanna-card border border-slate-200/60 z-10">
					<div class="py-1">
						<a href="/task-executions/{task.id}" class="block px-4 py-2 text-sm text-vanna-navy hover:bg-vanna-cream/50">
							View Details
						</a>
						{#if onStatusChange}
							<button onclick={() => onStatusChange?.(task.id, 'pending')} class="block w-full text-left px-4 py-2 text-sm text-vanna-navy hover:bg-vanna-cream/50">
								Mark as Pending
							</button>
							<button onclick={() => onStatusChange?.(task.id, 'running')} class="block w-full text-left px-4 py-2 text-sm text-vanna-navy hover:bg-vanna-cream/50">
								Mark as Running
							</button>
							<button onclick={() => onStatusChange?.(task.id, 'completed')} class="block w-full text-left px-4 py-2 text-sm text-vanna-navy hover:bg-vanna-cream/50">
								Mark as Completed
							</button>
						{/if}
						{#if onDelete}
							<button onclick={() => onDelete?.(task.id)} class="block w-full text-left px-4 py-2 text-sm text-vanna-orange hover:bg-vanna-cream/50">
								Delete Task
							</button>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	</div>

	<div class="space-y-2 mb-3">
		<div class="flex items-center text-xs text-slate-500">
			<svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
			</svg>
			{task.agent_name || `Agent ${task.agent_id}`}
		</div>
		{#if task.created_at}
			<div class="flex items-center text-xs text-slate-500">
				<svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				{formatTimeAgo(task.created_at)}
			</div>
		{/if}
	</div>

	<div class="flex items-center justify-between">
		<Badge variant={getStatusColor(task.status)} size="sm">
			{#if task.status?.toLowerCase() === 'waiting'}
				<svg class="w-3 h-3 mr-1 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
				</svg>
			{:else if task.status?.toLowerCase() === 'running'}
				<div class="w-2 h-2 bg-current rounded-full mr-1 animate-pulse"></div>
			{:else if task.status?.toLowerCase() === 'completed'}
				<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
				</svg>
			{:else if task.status?.toLowerCase() === 'failed'}
				<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
				</svg>
			{/if}
			{task.status || 'Unknown'}
		</Badge>

		{#if task.status?.toLowerCase() === 'waiting'}
			<a href="/task-executions/{task.id}" class="text-xs text-vanna-orange hover:text-vanna-orange/80 font-medium">
				Check Session
			</a>
		{/if}
	</div>
</div>

<!-- Click outside to close dropdown -->
<svelte:window onclick={(e) => {
	if (!e.target.closest('.relative')) {
		showDropdown = false;
	}
}} />
