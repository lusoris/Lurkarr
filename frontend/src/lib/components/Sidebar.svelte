<script lang="ts">
	import { page } from '$app/state';

	interface NavItem {
		href: string;
		label: string;
		icon: string;
	}

	const nav: NavItem[] = [
		{ href: '/', label: 'Dashboard', icon: '\u2302' },
		{ href: '/apps', label: 'Apps', icon: '\u2630' },
		{ href: '/logs', label: 'Logs', icon: '\u2263' },
		{ href: '/history', label: 'History', icon: '\u231A' },
		{ href: '/scheduling', label: 'Scheduling', icon: '\u23F0' },
		{ href: '/downloads', label: 'Downloads', icon: '\u2B07' },
		{ href: '/settings', label: 'Settings', icon: '\u2699' },
		{ href: '/user', label: 'Profile', icon: '\u263A' }
	];

	let collapsed = $state(false);
</script>

<aside
	class="flex flex-col h-screen bg-surface-900 border-r border-surface-800 transition-all duration-200
		{collapsed ? 'w-16' : 'w-56'}"
>
	<!-- Logo -->
	<div class="flex items-center gap-3 px-4 h-16 border-b border-surface-800">
		<span class="text-2xl">&#x1F438;</span>
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
				<span class="text-base w-5 text-center shrink-0">{item.icon}</span>
				{#if !collapsed}
					<span>{item.label}</span>
				{/if}
			</a>
		{/each}
	</nav>

	<!-- Collapse toggle -->
	<button
		onclick={() => collapsed = !collapsed}
		class="flex items-center justify-center h-12 border-t border-surface-800 text-surface-500 hover:text-surface-200 transition-colors"
	>
		<span class="text-sm">{collapsed ? '\u276F' : '\u276E'}</span>
	</button>
</aside>
