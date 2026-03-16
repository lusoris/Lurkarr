<script lang="ts" module>
	export interface Column<T = any> {
		key: string;
		header: string;
		sortable?: boolean;
		/** Extract a sortable/filterable value from the row. Defaults to (row as any)[key]. */
		accessor?: (row: T) => string | number;
		class?: string;
		headerClass?: string;
	}
</script>

<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-svelte';
	import * as Table from './table';
	import * as Pagination from './pagination';
	import { Input as ShadcnInput } from './input';
	import Button from './Button.svelte';

	interface Props {
		data: T[];
		columns: Column<T>[];
		/** Snippet to render each row. Receives (item, index). */
		row: Snippet<[T, number]>;
		/** Enable text search filtering. */
		searchable?: boolean;
		searchPlaceholder?: string;
		/** External search value (two-way bindable). */
		search?: string;
		/** Rows per page (0 = no pagination). */
		pageSize?: number;
		/** External page value (two-way bindable, 1-based). */
		page?: number;
		/** If provided, used for total count display (for server-side pagination). */
		totalItems?: number;
		/** Label shown next to pagination count. */
		noun?: string;
		class?: string;
		/** Extra content above the table (e.g. filters). */
		toolbar?: Snippet;
	}

	let {
		data,
		columns,
		row,
		searchable = false,
		searchPlaceholder = 'Search...',
		search = $bindable(''),
		pageSize = 0,
		page = $bindable(1),
		totalItems,
		noun = 'items',
		class: className = '',
		toolbar
	}: Props = $props();

	// --- Sorting ---
	let sortKey = $state('');
	let sortDir = $state<'asc' | 'desc'>('asc');

	function toggleSort(key: string) {
		if (sortKey === key) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortKey = key;
			sortDir = 'asc';
		}
		page = 1;
	}

	function getValue(item: T, col: Column<T>): string | number {
		if (col.accessor) return col.accessor(item);
		return (item as Record<string, any>)[col.key] ?? '';
	}

	// --- Filtered + sorted data ---
	const processed = $derived.by(() => {
		let result = data;

		// Client-side search filter
		if (searchable && search.trim()) {
			const q = search.trim().toLowerCase();
			result = result.filter((item) =>
				columns.some((col) => String(getValue(item, col)).toLowerCase().includes(q))
			);
		}

		// Sort
		if (sortKey) {
			const col = columns.find((c) => c.key === sortKey);
			if (col) {
				result = [...result].sort((a, b) => {
					const va = getValue(a, col);
					const vb = getValue(b, col);
					const cmp = typeof va === 'number' && typeof vb === 'number'
						? va - vb
						: String(va).localeCompare(String(vb));
					return sortDir === 'asc' ? cmp : -cmp;
				});
			}
		}

		return result;
	});

	// --- Pagination ---
	const isServerPaged = $derived(totalItems !== undefined);
	const totalCount = $derived(isServerPaged ? totalItems! : processed.length);
	const totalPages = $derived(pageSize > 0 ? Math.max(1, Math.ceil(totalCount / pageSize)) : 1);
	const paged = $derived(
		pageSize > 0 && !isServerPaged
			? processed.slice((page - 1) * pageSize, page * pageSize)
			: processed
	);

	// Clamp page if out of range
	$effect(() => {
		if (page > totalPages) page = totalPages;
		if (page < 1) page = 1;
	});
</script>

<div class={cn('space-y-3', className)}>
	{#if searchable || toolbar}
		<div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
			{#if searchable}
				<ShadcnInput
					type="text"
					placeholder={searchPlaceholder}
					bind:value={search}
					class="w-full sm:max-w-xs"
				/>
			{/if}
			{#if toolbar}
				{@render toolbar()}
			{/if}
		</div>
	{/if}

	<div class="rounded-xl border border-border overflow-hidden">
		<Table.Root>
			<Table.Header>
				<Table.Row class="bg-muted/50">
					{#each columns as col}
						<Table.Head class={cn(col.headerClass)}>
							{#if col.sortable}
								<Button
									type="button"
									variant="ghost"
									size="sm"
									class="h-auto -ml-2 px-2 py-1 font-medium"
									onclick={() => toggleSort(col.key)}
								>
									{col.header}
									{#if sortKey === col.key}
										{#if sortDir === 'asc'}
											<ArrowUp class="h-3 w-3 ml-1" />
										{:else}
											<ArrowDown class="h-3 w-3 ml-1" />
										{/if}
									{:else}
										<ArrowUpDown class="h-3 w-3 ml-1 opacity-30" />
									{/if}
								</Button>
							{:else}
								{col.header}
							{/if}
						</Table.Head>
					{/each}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each paged as item, i}
					{@render row(item, (page - 1) * (pageSize || 1) + i)}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="text-center text-muted-foreground py-8">
							{searchable && search ? 'No results match your search.' : 'No data.'}
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>

	{#if pageSize > 0 && totalCount > 0}
		<div class="flex items-center justify-between text-sm text-muted-foreground">
			<span>{totalCount.toLocaleString()} {noun}</span>
			{#if totalPages > 1}
				<Pagination.Root count={totalCount} perPage={pageSize} bind:page siblingCount={1}>
					{#snippet children({ pages })}
						<Pagination.Content>
							<Pagination.Item>
								<Pagination.Previous />
							</Pagination.Item>
							{#each pages as p (p.key)}
								{#if p.type === 'ellipsis'}
									<Pagination.Item>
										<Pagination.Ellipsis />
									</Pagination.Item>
								{:else}
									<Pagination.Item>
										<Pagination.Link page={p} isActive={page === p.value} />
									</Pagination.Item>
								{/if}
							{/each}
							<Pagination.Item>
								<Pagination.Next />
							</Pagination.Item>
						</Pagination.Content>
					{/snippet}
				</Pagination.Root>
			{/if}
		</div>
	{/if}
</div>
