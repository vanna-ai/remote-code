import { writable } from 'svelte/store';

export interface AuthState {
	authenticated: boolean;
	hasCredentials: boolean;
	loading: boolean;
	error: string | null;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>({
		authenticated: false,
		hasCredentials: false,
		loading: true,
		error: null
	});

	return {
		subscribe,

		async checkStatus() {
			update((state) => ({ ...state, loading: true, error: null }));
			try {
				const response = await fetch('/api/auth/status', {
					credentials: 'include'
				});
				if (response.ok) {
					const data = await response.json();
					set({
						authenticated: data.authenticated,
						hasCredentials: data.hasCredentials,
						loading: false,
						error: null
					});
					return data;
				} else {
					throw new Error('Failed to check auth status');
				}
			} catch (error) {
				update((state) => ({
					...state,
					loading: false,
					error: error instanceof Error ? error.message : 'Unknown error'
				}));
				return null;
			}
		},

		async registerPasskey(): Promise<boolean> {
			update((state) => ({ ...state, loading: true, error: null }));
			try {
				// Begin registration
				const beginRes = await fetch('/api/auth/register/begin', {
					method: 'POST',
					credentials: 'include'
				});

				if (!beginRes.ok) {
					throw new Error('Failed to begin registration');
				}

				const options = await beginRes.json();

				// Convert challenge and user ID from base64url to ArrayBuffer
				options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
				options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);

				// Convert excludeCredentials if present
				if (options.publicKey.excludeCredentials) {
					options.publicKey.excludeCredentials = options.publicKey.excludeCredentials.map(
						(cred: { id: string; type: string; transports?: string[] }) => ({
							...cred,
							id: base64urlToBuffer(cred.id)
						})
					);
				}

				// Create credential using WebAuthn API
				const credential = (await navigator.credentials.create({
					publicKey: options.publicKey
				})) as PublicKeyCredential;

				if (!credential) {
					throw new Error('Failed to create credential');
				}

				// Prepare the credential for sending to server
				const attestationResponse = credential.response as AuthenticatorAttestationResponse;
				const credentialData = {
					id: credential.id,
					rawId: bufferToBase64url(credential.rawId),
					type: credential.type,
					response: {
						clientDataJSON: bufferToBase64url(attestationResponse.clientDataJSON),
						attestationObject: bufferToBase64url(attestationResponse.attestationObject)
					}
				};

				// Finish registration
				const finishRes = await fetch('/api/auth/register/finish', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(credentialData),
					credentials: 'include'
				});

				if (!finishRes.ok) {
					const errorText = await finishRes.text();
					throw new Error(errorText || 'Failed to complete registration');
				}

				set({
					authenticated: true,
					hasCredentials: true,
					loading: false,
					error: null
				});

				return true;
			} catch (error) {
				update((state) => ({
					...state,
					loading: false,
					error: error instanceof Error ? error.message : 'Registration failed'
				}));
				return false;
			}
		},

		async login(): Promise<boolean> {
			update((state) => ({ ...state, loading: true, error: null }));
			try {
				// Begin login
				const beginRes = await fetch('/api/auth/login/begin', {
					method: 'POST',
					credentials: 'include'
				});

				if (!beginRes.ok) {
					throw new Error('Failed to begin login');
				}

				const options = await beginRes.json();

				// Convert challenge from base64url to ArrayBuffer
				options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);

				// Convert allowCredentials if present
				if (options.publicKey.allowCredentials) {
					options.publicKey.allowCredentials = options.publicKey.allowCredentials.map(
						(cred: { id: string; type: string; transports?: string[] }) => ({
							...cred,
							id: base64urlToBuffer(cred.id)
						})
					);
				}

				// Get credential using WebAuthn API
				const assertion = (await navigator.credentials.get({
					publicKey: options.publicKey
				})) as PublicKeyCredential;

				if (!assertion) {
					throw new Error('Failed to get credential');
				}

				// Prepare the assertion for sending to server
				const assertionResponse = assertion.response as AuthenticatorAssertionResponse;
				const assertionData = {
					id: assertion.id,
					rawId: bufferToBase64url(assertion.rawId),
					type: assertion.type,
					response: {
						clientDataJSON: bufferToBase64url(assertionResponse.clientDataJSON),
						authenticatorData: bufferToBase64url(assertionResponse.authenticatorData),
						signature: bufferToBase64url(assertionResponse.signature),
						userHandle: assertionResponse.userHandle
							? bufferToBase64url(assertionResponse.userHandle)
							: null
					}
				};

				// Finish login
				const finishRes = await fetch('/api/auth/login/finish', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(assertionData),
					credentials: 'include'
				});

				if (!finishRes.ok) {
					const errorText = await finishRes.text();
					throw new Error(errorText || 'Failed to complete login');
				}

				set({
					authenticated: true,
					hasCredentials: true,
					loading: false,
					error: null
				});

				return true;
			} catch (error) {
				update((state) => ({
					...state,
					loading: false,
					error: error instanceof Error ? error.message : 'Login failed'
				}));
				return false;
			}
		},

		async logout(): Promise<void> {
			try {
				await fetch('/api/auth/logout', {
					method: 'POST',
					credentials: 'include'
				});
			} catch {
				// Ignore errors during logout
			}

			set({
				authenticated: false,
				hasCredentials: true,
				loading: false,
				error: null
			});
		},

		clearError() {
			update((state) => ({ ...state, error: null }));
		}
	};
}

// Helper functions for base64url encoding/decoding
function base64urlToBuffer(base64url: string): ArrayBuffer {
	// Convert base64url to base64
	const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');

	// Pad with '=' if needed
	const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');

	// Decode to binary string
	const binary = atob(padded);

	// Convert to ArrayBuffer
	const buffer = new ArrayBuffer(binary.length);
	const bytes = new Uint8Array(buffer);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return buffer;
}

function bufferToBase64url(buffer: ArrayBuffer): string {
	const bytes = new Uint8Array(buffer);
	let binary = '';
	for (let i = 0; i < bytes.length; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	// Convert to base64 and then to base64url
	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

export const auth = createAuthStore();
