import { api } from '$lib/api';
import { type AppType } from '$lib';
import type { AppInstance } from '$lib/types';

export type { AppInstance };

let cache = $state<Record<string, AppInstance[]>>({});
let selectedApp = $state<string>('sonarr');
let selectedInstance = $state<string>('');
let fetching = $state(false);
let fetched = false;

async function fetchInstances(force = false) {
	if (fetching) return;
	if (fetched && !force) return;
	fetching = true;
	try {
		cache = await api.get<Record<string, AppInstance[]>>('/instances');
	} catch {
		cache = {};
	}
	fetched = true;
	fetching = false;
}

export function getInstances() {
	return {
		get selectedApp() { return selectedApp; },
		set selectedApp(v: string) {
			selectedApp = v;
			selectedInstance = '';
		},
		get selectedInstance() { return selectedInstance; },
		set selectedInstance(v: string) { selectedInstance = v; },
		get cache() { return cache; },
		get loading() { return fetching; },

		/** Instances for the currently selected app type. */
		get currentInstances(): AppInstance[] {
			// For whisparr UI tab, merge whisparr + eros instances
			if (selectedApp === 'whisparr') {
				return [...(cache['whisparr'] ?? []), ...(cache['eros'] ?? [])];
			}
			return cache[selectedApp] ?? [];
		},

		/** Fetch all instances (call once on app init or layout mount). */
		fetch: fetchInstances,

		/** Force-refetch instances (after CRUD mutations). */
		refetch: () => fetchInstances(true),

		/** Resolve instance name by ID from cache. */
		instanceName(instanceId: string): string {
			for (const list of Object.values(cache)) {
				const inst = list.find((i) => i.id === instanceId);
				if (inst) return inst.name;
			}
			return instanceId;
		}
	};
}
