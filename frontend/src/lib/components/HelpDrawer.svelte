<script lang="ts">
	import * as Sheet from './ui/sheet';
	import * as Accordion from './ui/accordion';
	import Badge from './ui/Badge.svelte';
	import Button from './ui/Button.svelte';
	import { Input as ShadcnInput } from './ui/input';
	import { HelpCircle, Search, ExternalLink, Lightbulb, BookOpen } from 'lucide-svelte';
	import { helpData } from '$lib/help-data';

	interface Props {
		/** Key into helpData — e.g. "apps", "lurk", "queue" */
		page: string;
	}

	let { page }: Props = $props();

	const data = $derived(helpData[page]);
	const sections = $derived(data?.sections ?? []);
	const quickStart = $derived(data?.quickStart);

	let open = $state(false);
	let search = $state('');
	const q = $derived(search.trim().toLowerCase());

	const filtered = $derived(
		q
			? sections
				.map(s => ({
					...s,
					tips: s.tips.filter(t => t.q.toLowerCase().includes(q) || t.a.toLowerCase().includes(q))
				}))
				.filter(s => s.tips.length > 0)
			: sections
	);
</script>

<Sheet.Root bind:open>
	<Sheet.Trigger>
		{#snippet child({ props })}
			<Button size="sm" variant="ghost" class="gap-1.5 text-muted-foreground" {...props}>
				<HelpCircle class="h-4 w-4" />
				Help
			</Button>
		{/snippet}
	</Sheet.Trigger>
	<Sheet.Content side="right" class="w-[400px] sm:w-[440px] overflow-y-auto">
		<Sheet.Header>
			<Sheet.Title class="flex items-center gap-2">
				<BookOpen class="h-5 w-5 text-primary" />
				Page Help
			</Sheet.Title>
			<Sheet.Description>Tips and guidance for this page.</Sheet.Description>
		</Sheet.Header>

		<div class="space-y-4 py-4">
			<!-- Search -->
			<div class="relative">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
				<ShadcnInput
					type="text"
					placeholder="Search help..."
					bind:value={search}
					class="pl-9 h-8 text-sm"
				/>
			</div>

			<!-- Quick Start Steps -->
			{#if quickStart && quickStart.length > 0 && !q}
				<div class="rounded-lg border border-primary/20 bg-primary/5 p-3">
					<div class="flex items-center gap-2 mb-2">
						<Lightbulb class="h-4 w-4 text-primary" />
						<span class="text-xs font-semibold text-primary uppercase tracking-wide">Quick Start</span>
					</div>
					<ol class="space-y-1.5 pl-4 list-decimal">
						{#each quickStart as step}
							<li class="text-sm text-muted-foreground leading-relaxed">{step}</li>
						{/each}
					</ol>
				</div>
			{/if}

			<!-- Help Sections -->
			{#if filtered.length === 0}
				<p class="text-sm text-muted-foreground text-center py-6">No results for "{search}"</p>
			{:else}
				{#each filtered as section}
					<div>
						<div class="flex items-center gap-2 mb-1">
							<h4 class="text-xs font-semibold text-foreground uppercase tracking-wide">{section.title}</h4>
							<Badge variant="default" class="text-[10px]">{section.tips.length}</Badge>
						</div>
						<Accordion.Root type="multiple">
							{#each section.tips as tip, i}
								<Accordion.Item value="{section.title}-{i}">
									<Accordion.Trigger class="text-sm text-start py-2">{tip.q}</Accordion.Trigger>
									<Accordion.Content>
										<p class="text-sm text-muted-foreground leading-relaxed pb-2">{tip.a}</p>
									</Accordion.Content>
								</Accordion.Item>
							{/each}
						</Accordion.Root>
					</div>
				{/each}
			{/if}

			<!-- Link to full help page -->
			<div class="pt-2 border-t border-border">
				<Button href="/help" variant="ghost" size="sm" class="w-full gap-1.5 text-muted-foreground">
					<ExternalLink class="h-3.5 w-3.5" />
					View all help topics
				</Button>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
