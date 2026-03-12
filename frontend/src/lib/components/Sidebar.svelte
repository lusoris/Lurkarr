<script lang="ts">
	import { page } from '$app/state';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface NavItem {
		href: string;
		label: string;
		icon: string;
	}

	const nav: NavItem[] = [
		{ href: '/', label: 'Dashboard', icon: 'dashboard' },
		{ href: '/apps', label: 'Apps', icon: 'apps' },
		{ href: '/history', label: 'History', icon: 'history' },
		{ href: '/scheduling', label: 'Scheduling', icon: 'scheduling' },
		{ href: '/downloads', label: 'Downloads', icon: 'downloads' },
		{ href: '/queue', label: 'Queue', icon: 'queue' },
		{ href: '/notifications', label: 'Notifications', icon: 'notifications' },
		{ href: '/monitoring', label: 'Monitoring', icon: 'monitoring' },
		{ href: '/settings', label: 'Settings', icon: 'settings' },
		{ href: '/user', label: 'Profile', icon: 'profile' }
	];

	let collapsed = $state(false);
	let mobileOpen = $state(false);

	/* SVG paths (Heroicons outline 24x24) */
	const iconPaths: Record<string, string> = {
		dashboard: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z',
		apps: 'M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9',
		history: 'M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z',
		scheduling: 'M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5',
		downloads: 'M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3',
		queue: 'M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 010 3.75H5.625a1.875 1.875 0 010-3.75z',
		notifications: 'M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0',
		monitoring: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z',
		settings: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
		profile: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z'
	};
</script>

<aside
	class="hidden md:flex flex-col h-screen bg-surface-900 border-r border-surface-800 transition-all duration-200
		{collapsed ? 'w-16' : 'w-56'}"
>
	<!-- Logo -->
	<div class="flex items-center gap-3 px-4 h-16 border-b border-surface-800">
		<img src="/logo.png" alt="Lurkarr" class="w-8 h-8 rounded" />
		{#if !collapsed}
			<span class="font-bold text-lurk-400 text-lg tracking-tight">Lurkarr</span>
		{/if}
	</div>

	<!-- Navigation -->
	<nav class="flex-1 py-3 space-y-0.5 overflow-y-auto">
		{#each nav as item}
			<a
				href={item.href}
				class="flex items-center gap-3 mx-2 px-3 py-2.5 rounded-lg text-sm transition-colors
					{page.url.pathname === item.href || (item.href !== '/' && page.url.pathname.startsWith(item.href))
						? 'bg-lurk-600/20 text-lurk-400 font-medium'
						: 'text-surface-400 hover:text-surface-100 hover:bg-surface-800'}"
			>
				<svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d={iconPaths[item.icon]} />
				</svg>
				{#if !collapsed}
					<span>{item.label}</span>
				{/if}
			</a>
		{/each}
	</nav>

	<!-- Collapse toggle -->
	<button
		onclick={() => collapsed = !collapsed}
		aria-label="Toggle sidebar"
		class="flex items-center justify-center h-12 border-t border-surface-800 text-surface-500 hover:text-surface-200 transition-colors"
	>
		<svg class="w-4 h-4 transition-transform {collapsed ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
		</svg>
	</button>
</aside>

<!-- Mobile top bar -->
<div class="md:hidden flex items-center gap-3 px-4 h-14 bg-surface-900 border-b border-surface-800 shrink-0">
	<button onclick={() => mobileOpen = true} aria-label="Open menu" class="text-surface-300 hover:text-surface-100">
		<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
		</svg>
	</button>
	<img src="/logo.png" alt="Lurkarr" class="w-7 h-7 rounded" />
	<span class="font-bold text-lurk-400 text-lg tracking-tight">Lurkarr</span>
</div>

<!-- Mobile overlay -->
{#if mobileOpen}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 md:hidden"
		onkeydown={(e) => { if (e.key === 'Escape') mobileOpen = false; }}
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => mobileOpen = false}></div>

		<!-- Drawer -->
		<nav
			class="absolute inset-y-0 left-0 w-64 bg-surface-900 border-r border-surface-800 flex flex-col"
			transition:fly={{ x: -256, duration: 200, easing: cubicOut }}
		>
			<div class="flex items-center justify-between px-4 h-14 border-b border-surface-800">
				<div class="flex items-center gap-3">
					<img src="/logo.png" alt="Lurkarr" class="w-8 h-8 rounded" />
					<span class="font-bold text-lurk-400 text-lg tracking-tight">Lurkarr</span>
				</div>
				<button onclick={() => mobileOpen = false} aria-label="Close menu" class="text-surface-400 hover:text-surface-100">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="flex-1 py-3 space-y-0.5 overflow-y-auto">
				{#each nav as item}
					<a
						href={item.href}
						onclick={() => mobileOpen = false}
						class="flex items-center gap-3 mx-2 px-3 py-2.5 rounded-lg text-sm transition-colors
							{page.url.pathname === item.href || (item.href !== '/' && page.url.pathname.startsWith(item.href))
								? 'bg-lurk-600/20 text-lurk-400 font-medium'
								: 'text-surface-400 hover:text-surface-100 hover:bg-surface-800'}"
					>
						<svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d={iconPaths[item.icon]} />
						</svg>
						<span>{item.label}</span>
					</a>
				{/each}
			</div>
		</nav>
	</div>
{/if}
