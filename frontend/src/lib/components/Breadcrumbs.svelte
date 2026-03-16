<script lang="ts">
	import { page } from '$app/state';
	import * as Breadcrumb from './ui/breadcrumb';
	import { Home } from 'lucide-svelte';

	/** Maps route segments to display labels. */
	const labels: Record<string, string> = {
		apps: 'Connections',
		lurk: 'Lurk Settings',
		queue: 'Queue',
		scheduling: 'Scheduling',
		downloads: 'Downloads',
		seerr: 'Seerr',
		dedup: 'Dedup',
		notifications: 'Notifications',
		history: 'History',
		activity: 'Activity',
		monitoring: 'Monitoring',
		settings: 'Settings',
		user: 'Profile',
		admin: 'Admin',
		users: 'Users',
		help: 'Help',
	};

	/** Maps top-level route segments to their sidebar category. */
	const categories: Record<string, string> = {
		apps: 'Configuration',
		lurk: 'Configuration',
		queue: 'Configuration',
		scheduling: 'Configuration',
		downloads: 'Operations',
		seerr: 'Operations',
		dedup: 'Operations',
		notifications: 'Operations',
		history: 'Monitoring',
		activity: 'Monitoring',
		monitoring: 'Monitoring',
		settings: 'Admin',
		user: 'Admin',
		admin: 'Admin',
		help: 'Admin',
	};

	const crumbs = $derived.by(() => {
		const path = page.url.pathname;
		if (path === '/') return [];

		const segments = path.split('/').filter(Boolean);
		const topSegment = segments[0];
		const category = categories[topSegment];

		const result: { label: string; href?: string; isLast: boolean }[] = [];

		// Add category crumb (non-navigable) if it exists
		if (category) {
			result.push({ label: category, isLast: false });
		}

		// Add each path segment
		segments.forEach((seg, i) => {
			result.push({
				label: labels[seg] ?? seg.charAt(0).toUpperCase() + seg.slice(1),
				href: '/' + segments.slice(0, i + 1).join('/'),
				isLast: i === segments.length - 1,
			});
		});

		return result;
	});
</script>

{#if crumbs.length > 0}
	<Breadcrumb.Root class="mb-4">
		<Breadcrumb.List>
			<Breadcrumb.Item>
				<Breadcrumb.Link href="/" class="flex items-center gap-1 text-muted-foreground hover:text-foreground">
					<Home class="h-3.5 w-3.5" />
					<span class="sr-only">Dashboard</span>
				</Breadcrumb.Link>
			</Breadcrumb.Item>

			{#each crumbs as crumb}
				<Breadcrumb.Separator />
				<Breadcrumb.Item>
					{#if crumb.isLast}
						<Breadcrumb.Page>{crumb.label}</Breadcrumb.Page>
					{:else if crumb.href}
						<Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
					{:else}
						<span class="text-muted-foreground text-sm">{crumb.label}</span>
					{/if}
				</Breadcrumb.Item>
			{/each}
		</Breadcrumb.List>
	</Breadcrumb.Root>
{/if}
