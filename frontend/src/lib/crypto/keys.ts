import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';

const DB_NAME = 'link-keys';
const STORE_NAME = 'keypair';

export function generateKeyPair(): { publicKey: string; secretKey: Uint8Array } {
	const kp = nacl.box.keyPair();
	return { publicKey: encodeBase64(kp.publicKey), secretKey: kp.secretKey };
}

/**
 * 從密碼和用戶 ID 推導出確定性的金鑰對
 * 同樣的密碼 + 用戶 ID 在任何裝置上都會產生相同的金鑰對
 */
export async function deriveKeyPairFromPassword(
	password: string,
	userId: string
): Promise<{ publicKey: string; secretKey: Uint8Array }> {
	// 使用 userId 作為 salt，確保同一用戶的金鑰一致
	const salt = new TextEncoder().encode(`link-e2e-${userId}`);

	const keyMaterial = await crypto.subtle.importKey(
		'raw',
		new TextEncoder().encode(password),
		'PBKDF2',
		false,
		['deriveBits']
	);

	// 推導 32 bytes 作為 NaCl 的 secret key
	const bits = await crypto.subtle.deriveBits(
		{ name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
		keyMaterial,
		256
	);

	const secretKey = new Uint8Array(bits);
	const keyPair = nacl.box.keyPair.fromSecretKey(secretKey);

	return {
		publicKey: encodeBase64(keyPair.publicKey),
		secretKey: keyPair.secretKey
	};
}

export async function saveSecretKey(secretKey: Uint8Array, password: string): Promise<void> {
	const salt = nacl.randomBytes(16);
	const key = await deriveKey(password, salt);
	const nonce = nacl.randomBytes(24);
	const encrypted = nacl.secretbox(secretKey, nonce, key);
	const data = {
		salt: encodeBase64(salt),
		nonce: encodeBase64(nonce),
		encrypted: encodeBase64(encrypted),
	};
	const db = await openDB();
	await putToDB(db, 'secretKey', JSON.stringify(data));
}

export async function loadSecretKey(password: string): Promise<Uint8Array | null> {
	console.log('loadSecretKey: opening IndexedDB...');
	const db = await openDB();
	const stored = await getFromDB(db, 'secretKey');
	console.log('loadSecretKey: stored data exists:', !!stored);
	if (!stored) {
		console.log('loadSecretKey: no stored key found in IndexedDB');
		return null;
	}
	const data = JSON.parse(stored);
	console.log('loadSecretKey: deriving key from password...');
	const key = await deriveKey(password, decodeBase64(data.salt));
	console.log('loadSecretKey: attempting to decrypt stored key...');
	const decrypted = nacl.secretbox.open(
		decodeBase64(data.encrypted),
		decodeBase64(data.nonce),
		key
	);
	if (decrypted) {
		console.log('loadSecretKey: decryption successful, key length:', decrypted.length);
	} else {
		console.log('loadSecretKey: decryption failed (wrong password?)');
	}
	return decrypted || null;
}

export async function hasSecretKey(): Promise<boolean> {
	const db = await openDB();
	return (await getFromDB(db, 'secretKey')) !== null;
}

export async function clearSecretKey(): Promise<void> {
	const db = await openDB();
	await deleteFromDB(db, 'secretKey');
}

async function deriveKey(password: string, salt: Uint8Array): Promise<Uint8Array> {
	const keyMaterial = await crypto.subtle.importKey(
		'raw',
		new TextEncoder().encode(password),
		'PBKDF2',
		false,
		['deriveBits']
	);
	const bits = await crypto.subtle.deriveBits(
		{ name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
		keyMaterial,
		256
	);
	return new Uint8Array(bits);
}

function openDB(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const req = indexedDB.open(DB_NAME, 1);
		req.onerror = () => reject(req.error);
		req.onsuccess = () => resolve(req.result);
		req.onupgradeneeded = () => req.result.createObjectStore(STORE_NAME);
	});
}

function putToDB(db: IDBDatabase, key: string, value: string): Promise<void> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(STORE_NAME, 'readwrite');
		tx.objectStore(STORE_NAME).put(value, key);
		tx.oncomplete = () => resolve();
		tx.onerror = () => reject(tx.error);
	});
}

function getFromDB(db: IDBDatabase, key: string): Promise<string | null> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(STORE_NAME, 'readonly');
		const req = tx.objectStore(STORE_NAME).get(key);
		req.onsuccess = () => resolve(req.result || null);
		req.onerror = () => reject(req.error);
	});
}

function deleteFromDB(db: IDBDatabase, key: string): Promise<void> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(STORE_NAME, 'readwrite');
		tx.objectStore(STORE_NAME).delete(key);
		tx.oncomplete = () => resolve();
		tx.onerror = () => reject(tx.error);
	});
}
