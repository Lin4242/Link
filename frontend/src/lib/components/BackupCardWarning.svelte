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
	<div class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
		<div class="bg-slate-800 border border-white/10 rounded-xl max-w-md w-full p-6 shadow-2xl">
			<div class="flex items-center gap-3 mb-4">
				<div class="w-12 h-12 bg-red-500/20 rounded-xl flex items-center justify-center">
					<svg class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				</div>
				<h2 class="text-lg font-semibold text-red-400">主卡撤銷警告</h2>
			</div>

			<div class="space-y-3 text-sm text-slate-300">
				<p class="text-red-400 font-medium">
					您正在使用附卡登入，此操作將：
				</p>
				<ul class="space-y-2 ml-1">
					<li class="flex items-start gap-2">
						<span class="text-red-400 mt-0.5">•</span>
						<span><strong class="text-white">立即撤銷主卡</strong> - 主卡將永久失效</span>
					</li>
					<li class="flex items-start gap-2">
						<span class="text-red-400 mt-0.5">•</span>
						<span><strong class="text-white">登出所有裝置</strong> - 所有現有連線將中斷</span>
					</li>
					<li class="flex items-start gap-2">
						<span class="text-red-400 mt-0.5">•</span>
						<span><strong class="text-white">附卡升級為主卡</strong> - 此卡將成為唯一登入卡</span>
					</li>
				</ul>
				<div class="bg-amber-500/10 border border-amber-500/20 p-3 rounded-xl mt-4">
					<p class="text-amber-400 text-xs">
						此操作無法撤銷！帳號將進入「單卡狀態」，無法再使用附卡撤銷功能。
					</p>
				</div>
			</div>

			<div class="mt-6 flex gap-3">
				<button
					onclick={onCancel}
					disabled={loading}
					class="flex-1 bg-slate-700 text-slate-300 py-3 px-4 rounded-xl hover:bg-slate-600 transition-colors disabled:opacity-50 border border-white/5"
				>
					取消
				</button>
				<button
					onclick={onConfirm}
					disabled={loading}
					class="flex-1 bg-gradient-to-r from-red-500 to-red-600 text-white py-3 px-4 rounded-xl hover:from-red-600 hover:to-red-700 transition-all disabled:opacity-50 shadow-lg shadow-red-500/20"
				>
					{loading ? '處理中...' : '確認撤銷'}
				</button>
			</div>
		</div>
	</div>
{/if}
