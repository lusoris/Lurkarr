<script lang="ts">
	import { goto } from '$app/navigation';
	import * as Command from '$lib/components/ui/command';
	import {
		LayoutDashboard, Cable, Flame, CalendarDays, History, Download,
		ListOrdered, Bell, Activity, Settings, Users, UserCircle,
		Layers, ScrollText, Film, HelpCircle
	} from 'lucide-svelte';
	import { getAuth } from '$lib/stores/auth.svelte';

	const auth = getAuth();

	interface PaletteItem {
		href: string;
		label: string;
		icon: typeof LayoutDashboard;
		group: string;
		keywords?: string;
		adminOnly?: boolean;
	}

	const items: PaletteItem[] = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard, group: 'Pages', keywords: 'home overview stats' },
		{ href: '/apps', label: 'Connections', icon: Cable, group: 'Configuration', keywords: 'apps instances sonarr radarr lidarr readarr whisparr prowlarr' },
		{ href: '/lurk', label: 'Lurk Settings', icon: Flame, group: 'Configuration', keywords: 'search missing upgrade batch cap' },
		{ href: '/queue', label: 'Queue', icon: ListOrdered, group: 'Configuration', keywords: 'cleaner strikes stalled slow blocklist seeding' },
		{ href: '/scheduling', label: 'Scheduling', icon: CalendarDays, group: 'Configuration', keywords: 'cron timer schedule days' },
		{ href: '/downloads', label: 'Downloads', icon: Download, group: 'Operations', keywords: 'active sabnzbd qbittorrent transmission' },
		{ href: '/seerr', label: 'Seerr', icon: Film, group: 'Operations', keywords: 'requests overseerr jellyseerr cleanup' },
		{ href: '/dedup', label: 'Dedup', icon: Layers, group: 'Operations', keywords: 'duplicates overlap instance groups' },
		{ href: '/notifications', label: 'Notifications', icon: Bell, group: 'Operations', keywords: 'discord telegram gotify ntfy pushover email webhook' },
		{ href: '/history', label: 'History', icon: History, group: 'Monitoring', keywords: 'past actions log' },
		{ href: '/activity', label: 'Activity', icon: ScrollText, group: 'Monitoring', keywords: 'realtime events feed' },
		{ href: '/monitoring', label: 'Monitoring', icon: Activity, group: 'Monitoring', keywords: 'prometheus metrics health status' },
		{ href: '/settings', label: 'Settings', icon: Settings, group: 'Admin', keywords: 'general config preferences' },
		{ href: '/admin/users', label: 'Users', icon: Users, group: 'Admin', keywords: 'accounts admin manage', adminOnly: true },
		{ href: '/user', label: 'Profile', icon: UserCircle, group: 'Admin', keywords: 'account password 2fa passkey' },
		{ href: '/help', label: 'Help & FAQ', icon: HelpCircle, group: 'Admin', keywords: 'documentation support troubleshoot' },
	];

	const filteredItems = $derived(items.filter(i => !i.adminOnly || auth.user?.is_admin));

	let open = $state(false);

	const groups = $derived(
		[...new Set(filteredItems.map(i => i.group))].map(g => ({
			label: g,
			items: filteredItems.filter(i => i.group === g)
		}))
	);

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			open = !open;
		}
	}

	function navigate(href: string) {
		open = false;
		goto(href);
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<Command.Dialog bind:open title="Command Palette" description="Navigate to a page">
	<Command.Input placeholder="Where to?" />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		{#each groups as group}
			<Command.Group heading={group.label}>
				{#each group.items as item}
					<Command.Item value="{item.label} {item.keywords ?? ''}" onSelect={() => navigate(item.href)}>
						<item.icon class="mr-2 h-4 w-4 shrink-0" />
						<span>{item.label}</span>
					</Command.Item>
				{/each}
			</Command.Group>
		{/each}
	</Command.List>
</Command.Dialog>
