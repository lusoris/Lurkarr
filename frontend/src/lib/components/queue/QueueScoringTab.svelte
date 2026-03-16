<script lang="ts">
	import Card from '$lib/components/ui/Card.svelte';
	import CollapsibleCard from '$lib/components/ui/CollapsibleCard.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import { appDisplayName, appButtonClass } from '$lib';
	import type { ScoringProfile } from '$lib/types';

	interface Props {
		app: string;
		profile: ScoringProfile | undefined;
		loaded: boolean;
		saving: boolean;
		onSave: () => Promise<void>;
		onProfileChange?: (partial: Partial<ScoringProfile>) => void;
	}

	let { app, profile, loaded, saving, onSave, onProfileChange }: Props = $props();

	let savingLocal = $state(false);

	async function save() {
		savingLocal = true;
		try {
			await onSave();
		} finally {
			savingLocal = false;
		}
	}
</script>

<p class="text-xs text-muted-foreground mb-3">Scoring profiles rank competing releases in the queue. Higher-weighted attributes contribute more to the final score. The queue cleaner uses these scores to decide which release to keep when duplicates are found.</p>
{#if profile}
	<div class="space-y-3">
		<Card>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				<Input bind:value={profile.name} label="Profile Name" />
				<Select bind:value={profile.strategy} label="Strategy">
					<option value="highest">Highest Score</option>
					<option value="adequate">Adequate Threshold</option>
				</Select>
			</div>
			{#if profile.strategy === 'adequate'}
				<Input bind:value={profile.adequate_threshold} type="number" label="Adequate Threshold" />
			{/if}
		</Card>

		<CollapsibleCard title="Preferences">
			<div class="space-y-2">
				<Toggle bind:checked={profile.prefer_higher_quality} label="Prefer Higher Quality" />
				<Toggle bind:checked={profile.prefer_larger_size} label="Prefer Larger Size" />
				<Toggle bind:checked={profile.prefer_indexer_flags} label="Prefer Indexer Flags" />
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Weights">
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				<Input bind:value={profile.custom_format_weight} type="number" label="Custom Format" hint="Authoritative — reflects your *arr quality profile" />
				<Input bind:value={profile.resolution_weight} type="number" label="Resolution" hint="2160p=4, 1080p=3, 720p=2, 480p=1" />
				<Input bind:value={profile.source_weight} type="number" label="Source" hint="Remux=5, BluRay=4, WEB-DL=3, WEBRip=2, HDTV=1" />
				<Input bind:value={profile.hdr_weight} type="number" label="HDR" hint="HDR10+=4, DV=3, HDR10=2, HDR=1" />
				<Input bind:value={profile.audio_weight} type="number" label="Audio" hint="Atmos=7, TrueHD=6, DTS-HD MA=5, FLAC=4, DDP=3" />
				<Input bind:value={profile.revision_bonus} type="number" label="Revision Bonus" hint="Flat bonus for PROPER/REPACK releases" />
				<Input bind:value={profile.size_weight} type="number" label="Size" />
				<Input bind:value={profile.age_weight} type="number" label="Age" />
				<Input bind:value={profile.seeders_weight} type="number" label="Seeders" />
			</div>
		</CollapsibleCard>

		<Button onclick={save} loading={savingLocal} class="w-full">Save Scoring Profile</Button>
	</div>
{:else if loaded}
	<Card>
		<p class="text-sm text-muted-foreground text-center py-4">No scoring profile configured for {appDisplayName(app)}.</p>
	</Card>
{:else}
	<Skeleton rows={4} height="h-10" />
{/if}
