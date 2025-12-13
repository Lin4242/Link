<script lang="ts">
	interface Props {
		show?: boolean;
		onConfirm?: () => void;
		onCancel?: () => void;
		loading?: boolean;
	}

	let { show = false, onConfirm, onCancel, loading = false }: Props = $props();
</script>

{#if show}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-lg max-w-md w-full p-6 shadow-xl">
			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 bg-red-100 rounded-full flex items-center justify-center">
					<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				</div>
				<h2 class="text-lg font-semibold text-red-700">附卡撤銷警告</h2>
			</div>

			<div class="space-y-3 text-sm text-gray-600">
				<p class="text-red-600 font-medium">
					您正在使用附卡（備援卡）登入，此操作將：
				</p>
				<ul class="list-disc list-inside space-y-1 ml-2">
					<li><strong>立即撤銷主卡</strong> - 主卡將永久失效</li>
					<li><strong>登出所有裝置</strong> - 所有現有連線將中斷</li>
					<li><strong>附卡升級為主卡</strong> - 此卡將成為您唯一的登入卡</li>
				</ul>
				<p class="text-yellow-700 font-medium bg-yellow-50 p-3 rounded">
					注意：此操作無法撤銷！帳號將進入「單卡狀態」，無法再使用附卡撤銷功能。
				</p>
			</div>

			<div class="mt-6 flex gap-3">
				<button
					onclick={onCancel}
					disabled={loading}
					class="flex-1 bg-gray-200 text-gray-700 py-2 px-4 rounded-lg hover:bg-gray-300 transition-colors disabled:opacity-50"
				>
					取消
				</button>
				<button
					onclick={onConfirm}
					disabled={loading}
					class="flex-1 bg-red-600 text-white py-2 px-4 rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
				>
					{loading ? '處理中...' : '確認撤銷主卡'}
				</button>
			</div>
		</div>
	</div>
{/if}
