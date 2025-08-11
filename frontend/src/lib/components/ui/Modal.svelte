<script lang="ts">
	interface Props {
		open?: boolean;
		title?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
		onClose?: () => void;
		class?: string;
	}

	let {
		open = false,
		title,
		size = 'md',
		onClose,
		class: className = '',
		children
	}: Props = $props();

	const sizeClasses = {
		sm: 'max-w-md',
		md: 'max-w-lg',
		lg: 'max-w-2xl',
		xl: 'max-w-4xl',
		'2xl': 'max-w-6xl'
	};

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget && onClose) {
			onClose();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && onClose) {
			onClose();
		}
	}
</script>

{#if open}
	<div 
		class="fixed inset-0 z-50 overflow-y-auto" 
		role="dialog" 
		aria-modal="true"
		on:keydown={handleKeydown}
	>
		<!-- Backdrop -->
		<div 
			class="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
			on:click={handleBackdropClick}
		></div>
		
		<!-- Modal container -->
		<div class="flex min-h-full items-center justify-center p-4">
			<div class="relative w-full {sizeClasses[size]} transform transition-all">
				<!-- Modal content -->
				<div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 {className}">
					{#if title}
						<!-- Header -->
						<div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
							<h3 class="text-xl font-semibold text-gray-900 dark:text-white">
								{title}
							</h3>
							{#if onClose}
								<button
									type="button"
									class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg p-2 transition-colors"
									on:click={onClose}
									aria-label="Close modal"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
									</svg>
								</button>
							{/if}
						</div>
					{/if}
					
					<!-- Content -->
					<div class="p-6">
						{@render children?.()}
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
