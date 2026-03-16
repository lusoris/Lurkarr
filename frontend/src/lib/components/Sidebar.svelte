<script lang="ts">
	import { page } from '$app/state';
	import { getAuth } from '$lib/stores/auth.svelte';
	import * as S from '$lib/components/ui/sidebar';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Button } from '$lib/components/ui/button';
	import {
		LayoutDashboard, Cable, Flame, CalendarDays, History, Download,
		ListOrdered, Bell, Activity, Settings, Users, UserCircle,
		Heart, LogOut, Layers, ScrollText, Film, HelpCircle
	} from '@lucide/svelte';

	const auth = getAuth();

	interface NavItem {
		href: string;
		label: string;
		icon: typeof LayoutDashboard;
		adminOnly?: boolean;
	}

	interface NavGroup {
		label: string;
		items: NavItem[];
	}

	const groups: NavGroup[] = [
		{ label: '', items: [
			{ href: '/', label: 'Dashboard', icon: LayoutDashboard },
		]},
		{ label: 'Configuration', items: [
			{ href: '/apps', label: 'Connections', icon: Cable },
			{ href: '/lurk', label: 'Lurk Settings', icon: Flame },
			{ href: '/queue', label: 'Queue', icon: ListOrdered },
			{ href: '/scheduling', label: 'Scheduling', icon: CalendarDays },
		]},
		{ label: 'Operations', items: [
			{ href: '/downloads', label: 'Downloads', icon: Download },
			{ href: '/seerr', label: 'Seerr', icon: Film },
			{ href: '/dedup', label: 'Dedup', icon: Layers },
			{ href: '/notifications', label: 'Notifications', icon: Bell },
		]},
		{ label: 'Monitoring', items: [
			{ href: '/history', label: 'History', icon: History },
			{ href: '/activity', label: 'Activity', icon: ScrollText },
			{ href: '/monitoring', label: 'Monitoring', icon: Activity },
		]},
		{ label: 'Admin', items: [
			{ href: '/settings', label: 'Settings', icon: Settings },
			{ href: '/admin/users', label: 'Users', icon: Users, adminOnly: true },
			{ href: '/user', label: 'Profile', icon: UserCircle },
			{ href: '/help', label: 'Help', icon: HelpCircle },
		]},
	];

	const filteredGroups = $derived(groups.map(g => ({
		...g,
		items: g.items.filter(item => !item.adminOnly || auth.user?.is_admin)
	})).filter(g => g.items.length > 0));

	function isActive(href: string): boolean {
		return page.url.pathname === href || (href !== '/' && page.url.pathname.startsWith(href));
	}

	async function handleLogout() {
		await auth.logout();
		window.location.href = '/login';
	}
</script>

<S.Root collapsible="icon" variant="sidebar">
	<S.Header class="border-b border-sidebar-border">
		<div class="flex items-center justify-center overflow-hidden py-1">
			<img src="/banner.png" alt="Lurkarr" class="h-8 w-auto object-contain group-data-[collapsible=icon]:hidden" />
			<img src="/logo.png" alt="Lurkarr" class="w-8 h-8 rounded hidden group-data-[collapsible=icon]:block" />
		</div>
	</S.Header>

	<S.Content>
		{#each filteredGroups as group}
			<S.Group>
				{#if group.label}
					<S.GroupLabel>{group.label}</S.GroupLabel>
				{/if}
				<S.GroupContent>
					<S.Menu>
						{#each group.items as item}
							<S.MenuItem>
								<S.MenuButton isActive={isActive(item.href)} tooltipContent={item.label}>
									{#snippet child({ props })}
										<a href={item.href} {...props}>
											<item.icon />
											<span>{item.label}</span>
										</a>
									{/snippet}
								</S.MenuButton>
							</S.MenuItem>
						{/each}
					</S.Menu>
				</S.GroupContent>
			</S.Group>
		{/each}
	</S.Content>

	<S.Footer class="border-t border-sidebar-border">
		<S.Menu>
			<S.MenuItem>
				<S.MenuButton tooltipContent={auth.user?.username ?? 'Profile'} size="lg">
					{#snippet child({ props })}
						<a href="/user" {...props}>
							<Avatar.Root class="size-6">
								<Avatar.Fallback class="text-[10px]">{(auth.user?.username ?? '?').slice(0, 2).toUpperCase()}</Avatar.Fallback>
							</Avatar.Root>
							<div class="flex flex-col flex-1 truncate text-start leading-tight">
								<span class="text-sm truncate">{auth.user?.username ?? 'Profile'}</span>
							</div>
						</a>
					{/snippet}
				</S.MenuButton>
			</S.MenuItem>
			<S.MenuItem>
				<S.MenuButton tooltipContent="Sign Out" class="text-sidebar-foreground/60 hover:text-destructive hover:bg-destructive/10" onclick={handleLogout}>
					<LogOut />
					<span>Sign Out</span>
				</S.MenuButton>
			</S.MenuItem>
		</S.Menu>
		<div class="flex items-center justify-center gap-2 text-[10px] text-muted-foreground py-1 group-data-[collapsible=icon]:flex-col group-data-[collapsible=icon]:gap-1">
			<a href="https://github.com/lusoris" target="_blank" rel="noopener noreferrer" class="hover:text-foreground transition-colors">&copy; lusoris</a>
			<span class="group-data-[collapsible=icon]:hidden">&middot;</span>
			<a href="https://ko-fi.com/lusoris" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 text-[#FF5E5B] hover:text-[#FF7674] transition-colors">
				<Heart class="h-3 w-3" />
				<span class="group-data-[collapsible=icon]:hidden">Ko-fi</span>
			</a>
		</div>
	</S.Footer>

	<S.Rail />
</S.Root>
