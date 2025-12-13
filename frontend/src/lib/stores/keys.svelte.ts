import { hasSecretKey, loadSecretKey, saveSecretKey, clearSecretKey } from '$lib/crypto';

function createKeysStore() {
	let secretKey = $state<Uint8Array | null>(null);
	let publicKeyCache = $state<Record<string, string>>({});
	let hasKey = $state(false);

	async function checkHasKey(): Promise<boolean> {
		hasKey = await hasSecretKey();
		return hasKey;
	}

	async function unlock(password: string): Promise<boolean> {
		console.log('keysStore.unlock called');
		try {
			const key = await loadSecretKey(password);
			console.log('loadSecretKey result:', key ? 'key loaded (length: ' + key.length + ')' : 'null/failed');
			if (key) {
				secretKey = key;
				return true;
			}
			return false;
		} catch (e) {
			console.error('keysStore.unlock error:', e);
			return false;
		}
	}

	async function save(key: Uint8Array, password: string): Promise<void> {
		await saveSecretKey(key, password);
		secretKey = key;
		hasKey = true;
	}

	async function clear(): Promise<void> {
		await clearSecretKey();
		secretKey = null;
		hasKey = false;
	}

	function cachePublicKey(userId: string, publicKey: string): void {
		publicKeyCache[userId] = publicKey;
	}

	function getPublicKey(userId: string): string | undefined {
		return publicKeyCache[userId];
	}

	function lock(): void {
		secretKey = null;
	}

	return {
		get secretKey() {
			return secretKey;
		},
		get hasKey() {
			return hasKey;
		},
		get publicKeyCache() {
			return publicKeyCache;
		},
		checkHasKey,
		unlock,
		save,
		clear,
		lock,
		cachePublicKey,
		getPublicKey,
	};
}

export const keysStore = createKeysStore();
