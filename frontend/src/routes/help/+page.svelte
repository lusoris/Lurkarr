<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { Input as ShadcnInput } from '$lib/components/ui/input';
	import * as Accordion from '$lib/components/ui/accordion';
	import { helpData } from '$lib/help-data';
	import {
		Search, Rocket, Cable, Flame, ListOrdered, CalendarDays, Download,
		Film, Layers, Bell, Activity, Settings, User, HelpCircle, ChevronRight,
		BarChart3, ExternalLink
	} from 'lucide-svelte';

	const pageCards: { key: string; label: string; href: string; icon: typeof Cable; color: string }[] = [
		{ key: 'apps', label: 'Connections', href: '/apps', icon: Cable, color: 'text-blue-400' },
		{ key: 'lurk', label: 'Lurk Settings', href: '/lurk', icon: Flame, color: 'text-orange-400' },
		{ key: 'queue', label: 'Queue Cleaner', href: '/queue', icon: ListOrdered, color: 'text-yellow-400' },
		{ key: 'scheduling', label: 'Scheduling', href: '/scheduling', icon: CalendarDays, color: 'text-green-400' },
		{ key: 'downloads', label: 'Downloads', href: '/downloads', icon: Download, color: 'text-cyan-400' },
		{ key: 'seerr', label: 'Seerr', href: '/seerr', icon: Film, color: 'text-purple-400' },
		{ key: 'dedup', label: 'Dedup', href: '/dedup', icon: Layers, color: 'text-pink-400' },
		{ key: 'notifications', label: 'Notifications', href: '/notifications', icon: Bell, color: 'text-amber-400' },
		{ key: 'monitoring', label: 'Monitoring', href: '/monitoring', icon: BarChart3, color: 'text-emerald-400' },
		{ key: 'settings', label: 'Settings', href: '/settings', icon: Settings, color: 'text-slate-400' },
		{ key: 'user', label: 'Profile & Security', href: '/user', icon: User, color: 'text-indigo-400' },
	];

	const gettingStartedSteps = [
		{ label: 'Add your first *arr app', description: 'Go to Connections, click Add Connection, and register your Sonarr, Radarr, or other *arr instances.', href: '/apps' },
		{ label: 'Connect download clients', description: 'Add SABnzbd, qBittorrent, or other clients so Lurkarr can monitor active transfers.', href: '/apps' },
		{ label: 'Configure Lurk Settings', description: 'Set search modes, batch sizes, and hourly caps to control how Lurkarr searches for media.', href: '/lurk' },
		{ label: 'Create a schedule', description: 'Automate searches and queue cleaning on a recurring schedule.', href: '/scheduling' },
		{ label: 'Set up notifications', description: 'Get alerts via Discord, Telegram, Gotify, or other services when Lurkarr takes actions.', href: '/notifications' },
	];

	// Flatten all help tips for global search
	const allTips = $derived(
		Object.entries(helpData).flatMap(([page, data]) =>
			data.sections.flatMap(section =>
				section.tips.map(tip => ({ page, section: section.title, ...tip }))
			)
		)
	);

	let search = $state('');
	const q = $derived(search.trim().toLowerCase());

	const searchResults = $derived(
		q
			? allTips.filter(t => t.q.toLowerCase().includes(q) || t.a.toLowerCase().includes(q))
			: []
	);

	const totalTopics = allTips.length;

	function pageLabel(key: string): string {
		return pageCards.find(p => p.key === key)?.label ?? key;
	}
</script>

<svelte:head><title>Help - Lurkarr</title></svelte:head>

<div class="space-y-8">
	<PageHeader title="Help Center">
		{#snippet actions()}
			<Badge variant="default">{totalTopics} topics</Badge>
		{/snippet}
	</PageHeader>

	<!-- Global Search -->
	<div class="relative max-w-xl">
		<Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
		<ShadcnInput
			type="text"
			placeholder="Search all help topics..."
			bind:value={search}
			class="pl-9 h-10"
		/>
	</div>

	<!-- Search Results -->
	{#if q}
		{#if searchResults.length === 0}
			<Card>
				<p class="text-center text-muted-foreground py-6">No results for "{search}". Try different keywords.</p>
			</Card>
		{:else}
			<div class="space-y-2">
				<p class="text-sm text-muted-foreground">{searchResults.length} result{searchResults.length !== 1 ? 's' : ''} found</p>
				<Accordion.Root type="multiple">
					{#each searchResults as result, i}
						<Accordion.Item value="search-{i}">
							<Accordion.Trigger class="text-sm text-start">
								<span class="flex items-center gap-2">
									<Badge variant="outline" class="text-[10px] shrink-0">{pageLabel(result.page)}</Badge>
									{result.q}
								</span>
							</Accordion.Trigger>
							<Accordion.Content>
								<p class="text-sm text-muted-foreground leading-relaxed">{result.a}</p>
							</Accordion.Content>
						</Accordion.Item>
					{/each}
				</Accordion.Root>
			</div>
		{/if}
	{:else}
		<!-- Getting Started Guide -->
		<section>
			<div class="flex items-center gap-2 mb-4">
				<Rocket class="h-5 w-5 text-primary" />
				<h2 class="text-base font-semibold text-foreground">Getting Started</h2>
			</div>
			<div class="grid gap-3">
				{#each gettingStartedSteps as step, i}
					<a href={step.href} class="group block">
						<Card class="!p-3 flex items-start gap-3 transition-colors group-hover:border-primary/40">
							<span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-primary text-xs font-bold shrink-0 mt-0.5">{i + 1}</span>
							<div class="min-w-0">
								<p class="text-sm font-medium text-foreground group-hover:text-primary transition-colors">{step.label}</p>
								<p class="text-xs text-muted-foreground mt-0.5">{step.description}</p>
							</div>
							<ChevronRight class="h-4 w-4 text-muted-foreground shrink-0 ml-auto mt-1 opacity-0 group-hover:opacity-100 transition-opacity" />
						</Card>
					</a>
				{/each}
			</div>
		</section>

		<!-- Page Help Cards -->
		<section>
			<div class="flex items-center gap-2 mb-4">
				<HelpCircle class="h-5 w-5 text-primary" />
				<h2 class="text-base font-semibold text-foreground">Help by Page</h2>
				<span class="text-xs text-muted-foreground">— click any page to visit it, each has a Help button in the header</span>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each pageCards as card}
					{@const data = helpData[card.key]}
					{@const tipCount = data?.sections.reduce((n, s) => n + s.tips.length, 0) ?? 0}
					<a href={card.href} class="group block">
						<Card class="!p-4 h-full transition-colors group-hover:border-primary/40">
							<div class="flex items-center gap-2 mb-2">
								<card.icon class="h-4 w-4 {card.color} shrink-0" />
								<span class="text-sm font-medium text-foreground group-hover:text-primary transition-colors">{card.label}</span>
								<Badge variant="outline" class="ml-auto text-[10px]">{tipCount}</Badge>
							</div>
							{#if data?.quickStart}
								<p class="text-xs text-muted-foreground line-clamp-2">{data.quickStart[0]}</p>
							{:else if data?.sections[0]?.tips[0]}
								<p class="text-xs text-muted-foreground line-clamp-2">{data.sections[0].tips[0].q}</p>
							{/if}
						</Card>
					</a>
				{/each}
			</div>
		</section>

		<!-- All Topics (expandable) -->
		<section>
			<div class="flex items-center gap-2 mb-4">
				<ExternalLink class="h-5 w-5 text-primary" />
				<h2 class="text-base font-semibold text-foreground">All Topics</h2>
			</div>
			<div class="space-y-4">
				{#each Object.entries(helpData) as [key, data]}
					{@const card = pageCards.find(c => c.key === key)}
					{#if card}
						<Card>
							<div class="flex items-center gap-2 mb-3">
								<card.icon class="h-4 w-4 {card.color} shrink-0" />
								<h3 class="text-sm font-semibold text-foreground">{card.label}</h3>
								<Badge variant="default" class="ml-auto text-[10px]">{data.sections.reduce((n, s) => n + s.tips.length, 0)}</Badge>
							</div>
							{#each data.sections as section}
								{#if data.sections.length > 1}
									<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mt-3 mb-1">{section.title}</p>
								{/if}
								<Accordion.Root type="multiple">
									{#each section.tips as tip, i}
										<Accordion.Item value="{key}-{section.title}-{i}">
											<Accordion.Trigger class="text-sm text-start">{tip.q}</Accordion.Trigger>
											<Accordion.Content>
												<p class="text-sm text-muted-foreground leading-relaxed">{tip.a}</p>
											</Accordion.Content>
										</Accordion.Item>
									{/each}
								</Accordion.Root>
							{/each}
						</Card>
					{/if}
				{/each}
			</div>
		</section>
	{/if}
</div>
