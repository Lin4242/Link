import { describe, it, expect } from 'vitest';
import nacl from 'tweetnacl';
import { encodeBase64 } from 'tweetnacl-util';
import { encryptMessage, encryptToString } from '../encrypt';
import { decryptMessage, decryptFromString } from '../decrypt';

describe('Crypto Module', () => {
	// Generate test key pairs
	const aliceKeyPair = nacl.box.keyPair();
	const bobKeyPair = nacl.box.keyPair();

	const alicePublicKey = encodeBase64(aliceKeyPair.publicKey);
	const bobPublicKey = encodeBase64(bobKeyPair.publicKey);

	describe('encryptMessage and decryptMessage', () => {
		it('should encrypt and decrypt a simple message', () => {
			const message = 'Hello, Bob!';

			// Alice encrypts message for Bob
			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);

			expect(encrypted).toHaveProperty('nonce');
			expect(encrypted).toHaveProperty('ciphertext');
			expect(encrypted.nonce.length).toBeGreaterThan(0);
			expect(encrypted.ciphertext.length).toBeGreaterThan(0);

			// Bob decrypts message from Alice
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should handle empty messages', () => {
			const message = '';

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should handle unicode messages', () => {
			const message = '你好世界！🎉 端對端加密測試';

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should handle long messages', () => {
			const message = 'A'.repeat(1000);

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should fail to decrypt with wrong key', () => {
			const message = 'Secret message';
			const wrongKeyPair = nacl.box.keyPair();

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);

			// Try to decrypt with wrong key
			const decrypted = decryptMessage(encrypted, alicePublicKey, wrongKeyPair.secretKey);

			expect(decrypted).toBeNull();
		});
	});

	describe('encryptToString and decryptFromString', () => {
		it('should encrypt to JSON string and decrypt from JSON string', () => {
			const message = 'Test message';

			const encryptedStr = encryptToString(message, bobPublicKey, aliceKeyPair.secretKey);

			expect(typeof encryptedStr).toBe('string');
			expect(() => JSON.parse(encryptedStr)).not.toThrow();

			const decrypted = decryptFromString(encryptedStr, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should return null for invalid JSON', () => {
			const decrypted = decryptFromString('not valid json', alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBeNull();
		});
	});

	describe('Padding', () => {
		it('should pad messages to minimum 256 bytes', () => {
			const shortMessage = 'Hi';

			const encrypted = encryptMessage(shortMessage, bobPublicKey, aliceKeyPair.secretKey);

			// The ciphertext should be at least 256 bytes (the minimum padded length)
			// plus authentication overhead (16 bytes for Poly1305)
			const ciphertextBytes = Uint8Array.from(atob(encrypted.ciphertext), (c) => c.charCodeAt(0));
			expect(ciphertextBytes.length).toBeGreaterThanOrEqual(256 + 16);

			// But it should still decrypt correctly
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);
			expect(decrypted).toBe(shortMessage);
		});

		it('should pad messages to 64-byte boundaries', () => {
			// Create messages of various lengths and verify padding
			const testMessages = ['A', 'AB', 'ABC'.repeat(100)];

			for (const message of testMessages) {
				const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
				const ciphertextBytes = Uint8Array.from(atob(encrypted.ciphertext), (c) => c.charCodeAt(0));

				// Padded length should be multiple of 64 (after subtracting 16-byte auth tag)
				const paddedLength = ciphertextBytes.length - 16;
				expect(paddedLength % 64).toBe(0);

				// Should still decrypt correctly
				const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);
				expect(decrypted).toBe(message);
			}
		});

		it('should produce different ciphertexts for same message due to random padding', () => {
			const message = 'Same message';

			const encrypted1 = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const encrypted2 = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);

			// Ciphertexts should be different due to random nonce and random padding
			expect(encrypted1.ciphertext).not.toBe(encrypted2.ciphertext);
			expect(encrypted1.nonce).not.toBe(encrypted2.nonce);

			// But both should decrypt to the same message
			const decrypted1 = decryptMessage(encrypted1, alicePublicKey, bobKeyPair.secretKey);
			const decrypted2 = decryptMessage(encrypted2, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted1).toBe(message);
			expect(decrypted2).toBe(message);
		});

		it('should prevent length analysis by having consistent ciphertext sizes for similar messages', () => {
			// Short messages should all have the same ciphertext length
			const shortMessages = ['a', 'ab', 'abc', 'abcd'];
			const ciphertextLengths = new Set<number>();

			for (const msg of shortMessages) {
				const encrypted = encryptMessage(msg, bobPublicKey, aliceKeyPair.secretKey);
				const ciphertextBytes = Uint8Array.from(atob(encrypted.ciphertext), (c) => c.charCodeAt(0));
				ciphertextLengths.add(ciphertextBytes.length);
			}

			// All short messages should have the same padded length (minimum 256 + 16 auth tag)
			expect(ciphertextLengths.size).toBe(1);
		});
	});

	describe('Edge Cases', () => {
		it('should handle special characters', () => {
			const message = '!@#$%^&*()_+-=[]{}|;:\'",.<>?/\\`~';

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should handle newlines and whitespace', () => {
			const message = 'Line 1\nLine 2\r\nLine 3\t\tTabbed';

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});

		it('should handle null bytes', () => {
			const message = 'Before\x00After';

			const encrypted = encryptMessage(message, bobPublicKey, aliceKeyPair.secretKey);
			const decrypted = decryptMessage(encrypted, alicePublicKey, bobKeyPair.secretKey);

			expect(decrypted).toBe(message);
		});
	});
});
