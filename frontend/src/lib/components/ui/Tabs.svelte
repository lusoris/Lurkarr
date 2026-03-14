<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import * as ShadcnTabs from './tabs';

	interface Tab {
		value: string;
		label: string;
		icon?: string;
		activeClass?: string;
	}

	interface Props {
		tabs: Tab[];
		value?: string;
		class?: string;
		onchange?: (value: string) => void;
	}

	let {
		tabs,
		value = $bindable(''),
		class: className = '',
		onchange
	}: Props = $props();

	function handleValueChange(v: string) {
		value = v;
		onchange?.(v);
	}
</script>

<ShadcnTabs.Root {value} onValueChange={handleValueChange}>
	<ShadcnTabs.List class={cn('overflow-x-auto', className)}>
		{#each tabs as tab}
			<ShadcnTabs.Trigger
				value={tab.value}
				class={cn(value === tab.value && tab.activeClass)}
			>
				{#if tab.icon}
					<img src={tab.icon} alt="" class="w-4 h-4 rounded-sm" />
				{/if}
				{tab.label}
			</ShadcnTabs.Trigger>
		{/each}
	</ShadcnTabs.List>
</ShadcnTabs.Root>
