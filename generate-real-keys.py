#!/usr/bin/env python3
import nacl.utils
import nacl.public
import base64

# 生成 F 的密鑰對
f_private_key = nacl.utils.random(nacl.public.PrivateKey.SIZE)
f_box = nacl.public.PrivateKey(f_private_key)
f_public_key = f_box.public_key

# 生成 N 的密鑰對  
n_private_key = nacl.utils.random(nacl.public.PrivateKey.SIZE)
n_box = nacl.public.PrivateKey(n_private_key)
n_public_key = n_box.public_key

# 轉換為 Base64
f_public_b64 = base64.b64encode(bytes(f_public_key)).decode('utf-8')
f_private_b64 = base64.b64encode(f_private_key).decode('utf-8')

n_public_b64 = base64.b64encode(bytes(n_public_key)).decode('utf-8')
n_private_b64 = base64.b64encode(n_private_key).decode('utf-8')

print("F 的密鑰對:")
print(f"公鑰: {f_public_b64}")
print(f"私鑰: {f_private_b64}")
print()
print("N 的密鑰對:")
print(f"公鑰: {n_public_b64}")
print(f"私鑰: {n_private_b64}")
print()
print("SQL 更新語句:")
print(f"UPDATE users SET public_key = '{f_public_b64}' WHERE nickname = 'F';")
print(f"UPDATE users SET public_key = '{n_public_b64}' WHERE nickname = 'N';")