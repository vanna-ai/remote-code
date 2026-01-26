<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';

	let loading = $state(true);
	let hasCredentials = $state(false);
	let error = $state<string | null>(null);
	let actionInProgress = $state(false);

	onMount(async () => {
		const status = await auth.checkStatus();
		if (status) {
			hasCredentials = status.hasCredentials;
			if (status.authenticated) {
				goto('/');
			}
		}
		loading = false;
	});

	async function handleSetupPasskey() {
		actionInProgress = true;
		error = null;
		try {
			const success = await auth.registerPasskey();
			if (success) {
				goto('/');
			} else {
				error = $auth.error || 'Failed to setup passkey';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'An error occurred';
		}
		actionInProgress = false;
	}

	async function handleLogin() {
		actionInProgress = true;
		error = null;
		try {
			const success = await auth.login();
			if (success) {
				goto('/');
			} else {
				error = $auth.error || 'Failed to login';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'An error occurred';
		}
		actionInProgress = false;
	}
</script>

<svelte:head>
	<title>Login - Web Terminal</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center p-4 -mt-16">
	<div class="w-full max-w-md">
		<Card class="p-8">
			<div class="text-center mb-8">
				<div class="w-16 h-16 bg-vanna-teal/10 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-vanna-teal" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-vanna-navy">Web Terminal</h1>
				<p class="text-vanna-navy/60 mt-2">
					{#if loading}
						Checking authentication...
					{:else if !hasCredentials}
						Setup your passkey to secure this application
					{:else}
						Authenticate with your passkey
					{/if}
				</p>
			</div>

			{#if error}
				<div class="mb-6 p-4 bg-vanna-orange/10 border border-vanna-orange/30 rounded-lg">
					<p class="text-vanna-orange text-sm">{error}</p>
				</div>
			{/if}

			{#if loading}
				<div class="flex justify-center py-8">
					<svg class="animate-spin h-8 w-8 text-vanna-teal" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
						<path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
				</div>
			{:else if !hasCredentials}
				<!-- First-time setup flow -->
				<div class="space-y-6">
					<div class="bg-vanna-cream/50 rounded-lg p-4">
						<h3 class="font-medium text-vanna-navy mb-2">First-time Setup</h3>
						<p class="text-sm text-vanna-navy/70">
							This application uses passkeys (WebAuthn) for secure, passwordless authentication.
							Your passkey is stored on your device and never leaves it.
						</p>
					</div>

					<Button
						variant="primary"
						size="lg"
						class="w-full"
						onclick={handleSetupPasskey}
						loading={actionInProgress}
						disabled={actionInProgress}
					>
						<svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
						</svg>
						Setup Passkey
					</Button>
				</div>
			{:else}
				<!-- Login flow -->
				<div class="space-y-6">
					<Button
						variant="primary"
						size="lg"
						class="w-full"
						onclick={handleLogin}
						loading={actionInProgress}
						disabled={actionInProgress}
					>
						<svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
						</svg>
						Login with Passkey
					</Button>

					<div class="text-center">
						<p class="text-xs text-vanna-navy/50">
							Use your device's biometric (Face ID, Touch ID, Windows Hello) or security key to authenticate
						</p>
					</div>
				</div>
			{/if}
		</Card>

		<p class="text-center text-xs text-vanna-navy/40 mt-6">
			Secured with WebAuthn passkey authentication
		</p>
	</div>
</div>
