import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';
import type { EncryptedData } from '$lib/types';

const MIN_PADDED_LENGTH = 256;
const PADDING_BLOCK_SIZE = 64;

function padMessage(message: string): Uint8Array {
	const msgBytes = new TextEncoder().encode(message);
	const msgLen = msgBytes.length;

	let paddedLen = Math.max(MIN_PADDED_LENGTH, msgLen + 4);
	paddedLen = Math.ceil(paddedLen / PADDING_BLOCK_SIZE) * PADDING_BLOCK_SIZE;

	const padded = new Uint8Array(paddedLen);

	const view = new DataView(padded.buffer);
	view.setUint32(0, msgLen, false);

	padded.set(msgBytes, 4);

	const randomPadding = nacl.randomBytes(paddedLen - 4 - msgLen);
	padded.set(randomPadding, 4 + msgLen);

	return padded;
}

export function encryptMessage(
	message: string,
	theirPublicKey: string,
	mySecretKey: Uint8Array
): EncryptedData {
	const nonce = nacl.randomBytes(24);
	const paddedMsg = padMessage(message);
	const ciphertext = nacl.box(paddedMsg, nonce, decodeBase64(theirPublicKey), mySecretKey);

	return {
		nonce: encodeBase64(nonce),
		ciphertext: encodeBase64(ciphertext),
	};
}

export function encryptToString(
	message: string,
	theirPublicKey: string,
	mySecretKey: Uint8Array
): string {
	return JSON.stringify(encryptMessage(message, theirPublicKey, mySecretKey));
}
