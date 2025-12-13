import nacl from 'tweetnacl';
import { decodeBase64 } from 'tweetnacl-util';
import type { EncryptedData } from '$lib/types';

function unpadMessage(padded: Uint8Array): string {
	const view = new DataView(padded.buffer, padded.byteOffset, padded.byteLength);
	const msgLen = view.getUint32(0, false);

	if (msgLen > padded.length - 4) {
		throw new Error('Invalid padded message');
	}

	const msgBytes = padded.slice(4, 4 + msgLen);
	return new TextDecoder().decode(msgBytes);
}

export function decryptMessage(
	encrypted: EncryptedData,
	theirPublicKey: string,
	mySecretKey: Uint8Array
): string | null {
	try {
		console.log('decryptMessage called:', {
			hasCiphertext: !!encrypted?.ciphertext,
			hasNonce: !!encrypted?.nonce,
			theirPublicKeyLength: theirPublicKey?.length,
			mySecretKeyLength: mySecretKey?.length
		});

		const decrypted = nacl.box.open(
			decodeBase64(encrypted.ciphertext),
			decodeBase64(encrypted.nonce),
			decodeBase64(theirPublicKey),
			mySecretKey
		);

		if (!decrypted) {
			console.log('nacl.box.open returned null - decryption failed');
			return null;
		}

		const result = unpadMessage(decrypted);
		console.log('Decryption succeeded, content length:', result.length);
		return result;
	} catch (e) {
		console.log('decryptMessage exception:', e);
		return null;
	}
}

export function decryptFromString(
	encryptedContent: string,
	theirPublicKey: string,
	mySecretKey: Uint8Array
): string | null {
	try {
		return decryptMessage(JSON.parse(encryptedContent), theirPublicKey, mySecretKey);
	} catch {
		return null;
	}
}
