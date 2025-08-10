<script>
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	
	let settings = {
		theme: 'system',
		notifications: true,
		autoSave: true,
		terminalFont: 'Monaco',
		terminalFontSize: 14
	};
	
	let isDirty = false;
	
	function updateSetting(key, value) {
		settings[key] = value;
		isDirty = true;
	}
	
	function saveSettings() {
		// TODO: Implement settings save
		console.log('Saving settings:', settings);
		isDirty = false;
	}
	
	function resetSettings() {
		settings = {
			theme: 'system',
			notifications: true,
			autoSave: true,
			terminalFont: 'Monaco',
			terminalFontSize: 14
		};
		isDirty = false;
	}
</script>

<svelte:head>
	<title>Settings - Remote-Code</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Settings</h1>
			<p class="mt-2 text-gray-600 dark:text-gray-400">Configure application settings and preferences</p>
		</div>
		{#if isDirty}
			<div class="flex items-center space-x-3">
				<Badge variant="warning" size="sm">Unsaved Changes</Badge>
				<Button onclick={resetSettings} variant="ghost">
					Reset
				</Button>
				<Button onclick={saveSettings} variant="primary">
					Save Changes
				</Button>
			</div>
		{/if}
	</div>

	<!-- Settings Sections -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Appearance Settings -->
		<div class="lg:col-span-2 space-y-6">
			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Appearance</h3>
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
							Theme
						</label>
						<select 
							bind:value={settings.theme}
							onchange={() => updateSetting('theme', settings.theme)}
							class="input-field"
						>
							<option value="light">Light</option>
							<option value="dark">Dark</option>
							<option value="system">System</option>
						</select>
						<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
							Choose your preferred color scheme
						</p>
					</div>
				</div>
			</Card>

			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Terminal</h3>
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
							Font Family
						</label>
						<select 
							bind:value={settings.terminalFont}
							onchange={() => updateSetting('terminalFont', settings.terminalFont)}
							class="input-field"
						>
							<option value="Monaco">Monaco</option>
							<option value="Menlo">Menlo</option>
							<option value="Ubuntu Mono">Ubuntu Mono</option>
							<option value="Consolas">Consolas</option>
							<option value="Courier New">Courier New</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
							Font Size
						</label>
						<input 
							type="range" 
							min="10" 
							max="20" 
							bind:value={settings.terminalFontSize}
							onchange={() => updateSetting('terminalFontSize', settings.terminalFontSize)}
							class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer"
						/>
						<div class="flex justify-between text-sm text-gray-500 dark:text-gray-400 mt-1">
							<span>10px</span>
							<span>{settings.terminalFontSize}px</span>
							<span>20px</span>
						</div>
					</div>
				</div>
			</Card>

			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Notifications</h3>
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<div>
							<label class="text-sm font-medium text-gray-700 dark:text-gray-300">
								Enable Notifications
							</label>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								Receive notifications for task completions and system events
							</p>
						</div>
						<label class="relative inline-flex items-center cursor-pointer">
							<input 
								type="checkbox" 
								bind:checked={settings.notifications}
								onchange={() => updateSetting('notifications', settings.notifications)}
								class="sr-only peer"
							/>
							<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
						</label>
					</div>
					<div class="flex items-center justify-between">
						<div>
							<label class="text-sm font-medium text-gray-700 dark:text-gray-300">
								Auto-save Settings
							</label>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								Automatically save changes as you make them
							</p>
						</div>
						<label class="relative inline-flex items-center cursor-pointer">
							<input 
								type="checkbox" 
								bind:checked={settings.autoSave}
								onchange={() => updateSetting('autoSave', settings.autoSave)}
								class="sr-only peer"
							/>
							<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
						</label>
					</div>
				</div>
			</Card>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Quick Actions</h3>
				<div class="space-y-3">
					<Button href="/agents" variant="ghost" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
						</svg>
						Manage Agents
					</Button>
					<Button href="/directories" variant="ghost" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>
						</svg>
						Base Directories
					</Button>
					<Button href="/terminal" variant="ghost" class="w-full justify-start">
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
						</svg>
						Terminal Sessions
					</Button>
				</div>
			</Card>

			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">System Info</h3>
				<div class="space-y-3 text-sm">
					<div class="flex justify-between">
						<span class="text-gray-600 dark:text-gray-400">Version</span>
						<span class="text-gray-900 dark:text-white font-mono">1.0.0</span>
					</div>
					<div class="flex justify-between">
						<span class="text-gray-600 dark:text-gray-400">Platform</span>
						<span class="text-gray-900 dark:text-white">Web</span>
					</div>
					<div class="flex justify-between">
						<span class="text-gray-600 dark:text-gray-400">Build</span>
						<span class="text-gray-900 dark:text-white font-mono">dev</span>
					</div>
				</div>
			</Card>

			<Card>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Support</h3>
				<div class="space-y-3">
					<a href="https://github.com/remote-code/remote-code" target="_blank" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
						📚 Documentation
					</a>
					<a href="https://github.com/remote-code/remote-code/issues" target="_blank" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
						🐛 Report Issues
					</a>
					<a href="https://github.com/remote-code/remote-code" target="_blank" class="block text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">
						⭐ Star on GitHub
					</a>
				</div>
			</Card>
		</div>
	</div>
</div>