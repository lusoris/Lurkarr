import { api } from '$lib/api';
import { appTypes, type AppType } from '$lib';

export interface AppInstance {
	id: string;
	app_type: string;
	name: string;
	enabled: boolean;
}

let cache = $state<Record<string, AppInstance[]>>({});
let selectedApp = $state<string>('sonarr');
let selectedInstance = $state<string>('');
let fetching = $state(false);

async function fetchInstances() {
	if (fetching) return;
	fetching = true;
	const result: Record<string, AppInstance[]> = {};
	await Promise.all(
		appTypes.map(async (app) => {
			try {
				result[app] = await api.get<AppInstance[]>(`/instances/${app}`);
			} catch {
				result[app] = [];
			}
		})
	);
	cache = result;
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
