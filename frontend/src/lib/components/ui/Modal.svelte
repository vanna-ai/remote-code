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

	let modalElement: HTMLDivElement | undefined = $state();
	let previousActiveElement: Element | null = null;

	const sizeClasses = {
		sm: 'max-w-md',
		md: 'max-w-lg',
		lg: 'max-w-2xl',
		xl: 'max-w-4xl',
		'2xl': 'max-w-6xl'
	};

	// Focus management: trap focus and restore on close
	$effect(() => {
		if (open && modalElement) {
			previousActiveElement = document.activeElement;
			// Focus the first focusable element in the modal
			const focusable = modalElement.querySelectorAll<HTMLElement>(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			);
			if (focusable.length > 0) {
				focusable[0].focus();
			}
		} else if (!open && previousActiveElement instanceof HTMLElement) {
			previousActiveElement.focus();
			previousActiveElement = null;
		}
	});

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget && onClose) {
			onClose();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && onClose) {
			onClose();
		}

		// Focus trap
		if (event.key === 'Tab' && modalElement) {
			const focusable = modalElement.querySelectorAll<HTMLElement>(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			);
			if (focusable.length === 0) return;

			const first = focusable[0];
			const last = focusable[focusable.length - 1];

			if (event.shiftKey && document.activeElement === first) {
				event.preventDefault();
				last.focus();
			} else if (!event.shiftKey && document.activeElement === last) {
				event.preventDefault();
				first.focus();
			}
		}
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="fixed inset-0 z-50 overflow-y-auto"
		role="dialog"
		aria-modal="true"
		tabindex="-1"
		onkeydown={handleKeydown}
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 bg-vanna-navy/50 backdrop-blur-sm transition-opacity"
			onclick={handleBackdropClick}
			aria-hidden="true"
		></div>

		<!-- Modal container -->
		<div class="flex min-h-full items-center justify-center p-4">
			<div class="relative w-full {sizeClasses[size]} transform transition-all">
				<!-- Modal content -->
				<div bind:this={modalElement} class="relative bg-white rounded-2xl shadow-vanna-feature border border-slate-200/60 {className}">
					{#if title}
						<!-- Header -->
						<div class="flex items-center justify-between p-6 border-b border-slate-200">
							<h3 class="text-xl font-semibold text-vanna-navy font-serif">
								{title}
							</h3>
							{#if onClose}
								<button
									type="button"
									class="text-slate-400 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg p-2 transition-colors"
									onclick={onClose}
									aria-label="Close modal"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
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
