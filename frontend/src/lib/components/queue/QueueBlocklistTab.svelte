<script lang="ts">
	import Card from '$lib/components/ui/Card.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import Badge from '$lib/components/ui/Badge.svelte';
	import type { BlocklistEntry } from '$lib/types';

	interface Props {
		app: string;
		blocklist: BlocklistEntry[];
		loading: boolean;
	}

	let { app, blocklist, loading }: Props = $props();

	const blocklistLogColumns: Column<BlocklistEntry>[] = [
		{ key: 'title', header: 'Title', sortable: true },
		{ key: 'reason', header: 'Reason', sortable: true },
		{ key: 'blocklisted_at', header: 'Date', sortable: true }
	];
</script>

{#if loading}
	<Skeleton rows={4} height="h-10" />
{:else if blocklist.length === 0}
	<EmptyState title="No blocklist entries" description="Items removed from the queue will appear here." />
{:else}
	<DataTable data={blocklist} columns={blocklistLogColumns} searchable pageSize={50} noun="entries">
		{#snippet row(entry)}
			<T.Row>
				<T.Cell class="text-foreground max-w-xs truncate">{entry.title}</T.Cell>
				<T.Cell><Badge variant="error">{entry.reason}</Badge></T.Cell>
				<T.Cell class="text-muted-foreground text-xs">{new Date(entry.blocklisted_at).toLocaleString()}</T.Cell>
			</T.Row>
		{/snippet}
	</DataTable>
{/if}
