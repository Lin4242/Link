<script lang="ts">
	interface Props {
		show?: boolean;
		onClose?: () => void;
	}

	let { show = true, onClose }: Props = $props();
</script>

{#if show}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-lg max-w-md w-full p-6 shadow-xl">
			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 bg-yellow-100 rounded-full flex items-center justify-center">
					<svg
						class="w-6 h-6 text-yellow-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				</div>
				<h2 class="text-lg font-semibold text-gray-900">安全提醒</h2>
			</div>

			<div class="space-y-3 text-sm text-gray-600">
				<p>
					<strong>私鑰儲存在本地裝置</strong>
				</p>
				<ul class="list-disc list-inside space-y-1 ml-2">
					<li>您的私鑰僅存於此瀏覽器</li>
					<li>清除瀏覽器資料會導致私鑰遺失</li>
					<li>私鑰遺失後，歷史訊息將無法解密</li>
				</ul>
				<p class="text-yellow-700 font-medium">
					建議：請勿在公用電腦上使用，並定期備份重要訊息。
				</p>
			</div>

			<button
				onclick={onClose}
				class="mt-6 w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors"
			>
				我了解了
			</button>
		</div>
	</div>
{/if}
