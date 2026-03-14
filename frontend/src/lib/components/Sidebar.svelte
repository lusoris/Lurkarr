<script lang="ts">
	import { page } from '$app/state';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { getAuth } from '$lib/stores/auth.svelte';
	import { cn } from '$lib/lib/utils';
	import {
		LayoutDashboard, Cable, Flame, CalendarDays, History, Download,
		ListOrdered, Bell, Activity, Settings, Users, UserCircle,
		ChevronLeft, Menu, X, Heart, LogOut, Layers, ScrollText
	} from 'lucide-svelte';

	const auth = getAuth();

	interface NavItem {
		href: string;
		label: string;
		icon: typeof LayoutDashboard;
		adminOnly?: boolean;
	}

	const allNav: NavItem[] = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/apps', label: 'Connections', icon: Cable },
		{ href: '/lurk', label: 'Lurk Settings', icon: Flame },
		{ href: '/scheduling', label: 'Scheduling', icon: CalendarDays },
		{ href: '/history', label: 'History', icon: History },
		{ href: '/activity', label: 'Activity', icon: ScrollText },
		{ href: '/downloads', label: 'Downloads', icon: Download },
		{ href: '/queue', label: 'Queue', icon: ListOrdered },
		{ href: '/dedup', label: 'Dedup', icon: Layers },
		{ href: '/notifications', label: 'Notifications', icon: Bell },
		{ href: '/monitoring', label: 'Monitoring', icon: Activity },
		{ href: '/settings', label: 'Settings', icon: Settings },
		{ href: '/admin/users', label: 'Users', icon: Users, adminOnly: true },
		{ href: '/user', label: 'Profile', icon: UserCircle }
	];

	const nav = $derived(allNav.filter(item => !item.adminOnly || auth.user?.is_admin));

	let collapsed = $state(false);
	let mobileOpen = $state(false);

	function isActive(href: string): boolean {
		return page.url.pathname === href || (href !== '/' && page.url.pathname.startsWith(href));
	}

	async function handleLogout() {
		await auth.logout();
		window.location.href = '/login';
	}
</script>

<aside
	class={cn(
		'hidden md:flex flex-col h-screen bg-sidebar border-r border-sidebar-border transition-all duration-200',
		collapsed ? 'w-16' : 'w-56'
	)}
>
	<!-- Logo / Banner -->
	<div class="flex items-center justify-center px-2 py-3 border-b border-sidebar-border overflow-hidden">
		{#if collapsed}
			<img src="/logo.png" alt="Lurkarr" class="w-10 h-10 rounded transition-all" />
		{:else}
			<img src="/banner.png" alt="Lurkarr" class="w-full h-auto max-h-16 object-contain transition-all" />
		{/if}
	</div>

	<!-- Navigation -->
	<nav class="flex-1 py-3 space-y-0.5 overflow-y-auto">
		{#each nav as item}
			{@const active = isActive(item.href)}
			<a
				href={item.href}
				class={cn(
					'flex items-center gap-3 mx-2 px-3 py-2 rounded-md text-sm transition-colors',
					active
						? 'bg-sidebar-primary/15 text-sidebar-primary font-medium'
						: 'text-sidebar-foreground/60 hover:text-sidebar-foreground hover:bg-sidebar-accent'
				)}
			>
				<item.icon class="h-5 w-5 shrink-0" />
				{#if !collapsed}
					<span>{item.label}</span>
				{/if}
			</a>
		{/each}

		<!-- Sign Out (inside scrollable nav) -->
		<div class="!mt-3 pt-3 mx-2 border-t border-sidebar-border/50">
			<button
				onclick={handleLogout}
				class={cn(
					'flex items-center gap-3 w-full px-3 py-2 rounded-md text-sm transition-colors text-sidebar-foreground/60 hover:text-destructive hover:bg-destructive/10',
					collapsed && 'justify-center'
				)}
				title="Sign out"
			>
				<LogOut class="h-5 w-5 shrink-0" />
				{#if !collapsed}
					<span>Sign Out</span>
				{/if}
			</button>
		</div>
	</nav>

	<!-- Footer + Collapse -->
	<div class="flex items-center border-t border-sidebar-border">
		<div class={cn(
			'flex-1 text-muted-foreground transition-all',
			collapsed ? 'flex flex-col items-center gap-0.5 py-1.5' : 'px-3 py-1.5 text-center text-[10px] leading-relaxed'
		)}>
			{#if collapsed}
				<a href="https://github.com/lusoris" target="_blank" rel="noopener noreferrer" title="© lusoris · AGPL-3.0" class="hover:text-foreground transition-colors text-xs">©</a>
				<a href="https://ko-fi.com/lusoris" target="_blank" rel="noopener noreferrer" title="Support on Ko-fi" class="text-[#FF5E5B] hover:text-[#FF7674] transition-colors">
					<Heart class="h-3 w-3" />
				</a>
			{:else}
				<p>&copy; <a href="https://github.com/lusoris" target="_blank" rel="noopener noreferrer" class="hover:text-foreground transition-colors">lusoris</a> &middot; <a href="https://www.gnu.org/licenses/agpl-3.0.html" target="_blank" rel="noopener noreferrer" class="hover:text-foreground transition-colors">AGPL-3.0</a></p>
				<a href="https://ko-fi.com/lusoris" target="_blank" rel="noopener noreferrer" class="inline-flex items-center justify-center gap-1 text-[#FF5E5B] hover:text-[#FF7674] transition-colors">
					<Heart class="h-3 w-3" />
					Support on Ko-fi
				</a>
			{/if}
		</div>
		<button
			onclick={() => collapsed = !collapsed}
			aria-label="Toggle sidebar"
			class="shrink-0 flex items-center justify-center w-10 h-10 text-muted-foreground hover:text-foreground transition-colors"
		>
			<ChevronLeft class={cn('h-4 w-4 transition-transform', collapsed && 'rotate-180')} />
		</button>
	</div>
</aside>

<!-- Mobile top bar -->
<div class="md:hidden flex items-center gap-3 px-4 h-14 bg-sidebar border-b border-sidebar-border shrink-0">
	<button onclick={() => mobileOpen = true} aria-label="Open menu" class="text-muted-foreground hover:text-foreground">
		<Menu class="h-6 w-6" />
	</button>
	<img src="/banner.png" alt="Lurkarr" class="h-8 w-auto object-contain" />
</div>

<!-- Mobile overlay -->
{#if mobileOpen}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 md:hidden"
		onkeydown={(e) => { if (e.key === 'Escape') mobileOpen = false; }}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => mobileOpen = false}></div>

		<nav
			class="absolute inset-y-0 left-0 w-64 bg-sidebar border-r border-sidebar-border flex flex-col"
			transition:fly={{ x: -256, duration: 200, easing: cubicOut }}
		>
			<div class="flex items-center justify-between px-4 py-3 border-b border-sidebar-border">
				<img src="/banner.png" alt="Lurkarr" class="h-8 w-auto object-contain" />
				<button onclick={() => mobileOpen = false} aria-label="Close menu" class="text-muted-foreground hover:text-foreground">
					<X class="h-5 w-5" />
				</button>
			</div>

			<div class="flex-1 py-3 space-y-0.5 overflow-y-auto">
				{#each nav as item}
					{@const active = isActive(item.href)}
					<a
						href={item.href}
						onclick={() => mobileOpen = false}
						class={cn(
							'flex items-center gap-3 mx-2 px-3 py-2 rounded-md text-sm transition-colors',
							active
								? 'bg-sidebar-primary/15 text-sidebar-primary font-medium'
								: 'text-sidebar-foreground/60 hover:text-sidebar-foreground hover:bg-sidebar-accent'
						)}
					>
						<item.icon class="h-5 w-5 shrink-0" />
						<span>{item.label}</span>
					</a>
				{/each}
			</div>

			<!-- Logout -->
			<button
				onclick={handleLogout}
				class="flex items-center gap-3 mx-2 px-3 py-2 rounded-md text-sm transition-colors text-sidebar-foreground/60 hover:text-destructive hover:bg-destructive/10"
			>
				<LogOut class="h-5 w-5 shrink-0" />
				<span>Sign Out</span>
			</button>

			<div class="px-4 py-3 border-t border-sidebar-border text-[10px] text-muted-foreground leading-relaxed text-center">
				<p>&copy; <a href="https://github.com/lusoris" target="_blank" rel="noopener noreferrer" class="hover:text-foreground">lusoris</a> &middot; <a href="https://www.gnu.org/licenses/agpl-3.0.html" target="_blank" rel="noopener noreferrer" class="hover:text-foreground">AGPL-3.0</a></p>
				<a href="https://ko-fi.com/lusoris" target="_blank" rel="noopener noreferrer" class="inline-flex items-center justify-center gap-1.5 text-[#FF5E5B] hover:text-[#FF7674] transition-colors mt-1">
					<Heart class="h-3.5 w-3.5" />
					Support on Ko-fi
				</a>
			</div>
		</nav>
	</div>
{/if}
