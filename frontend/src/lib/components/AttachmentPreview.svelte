<script lang="ts">
	import { X, Download, ZoomIn, ZoomOut, RotateCw } from 'lucide-svelte';

	interface Props {
		name: string;
		contentType: string;
		content: string; // Base64 encoded
		onclose: () => void;
		ondownload?: () => void;
	}

	let { name, contentType, content, onclose, ondownload }: Props = $props();

	let zoom = $state(1);
	let rotation = $state(0);

	const isImage = $derived(contentType.startsWith('image/'));
	const isPdf = $derived(contentType.includes('pdf'));
	const isText = $derived(contentType.startsWith('text/') || contentType.includes('json') || contentType.includes('xml'));
	const isVideo = $derived(contentType.startsWith('video/'));
	const isAudio = $derived(contentType.startsWith('audio/'));

	const dataUrl = $derived(`data:${contentType};base64,${content}`);

	// Decode text content
	const textContent = $derived(() => {
		if (!isText) return '';
		try {
			return atob(content);
		} catch {
			return 'Не удалось декодировать содержимое';
		}
	});

	function zoomIn() {
		zoom = Math.min(zoom + 0.25, 3);
	}

	function zoomOut() {
		zoom = Math.max(zoom - 0.25, 0.25);
	}

	function rotate() {
		rotation = (rotation + 90) % 360;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onclose();
		} else if (event.key === '+' || event.key === '=') {
			zoomIn();
		} else if (event.key === '-') {
			zoomOut();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Backdrop -->
<div
	class="fixed inset-0 bg-black/80 z-50 flex items-center justify-center"
	onclick={onclose}
	role="dialog"
	aria-modal="true"
>
	<!-- Modal container -->
	<div
		class="relative max-w-[90vw] max-h-[90vh] flex flex-col"
		onclick={(e) => e.stopPropagation()}
	>
		<!-- Header -->
		<div class="flex items-center justify-between px-4 py-3 bg-gray-900 rounded-t-lg">
			<div class="flex items-center gap-3">
				<span class="text-white font-medium truncate max-w-[300px]">{name}</span>
				<span class="text-gray-400 text-sm">({contentType})</span>
			</div>
			<div class="flex items-center gap-2">
				{#if isImage}
					<button
						onclick={zoomOut}
						class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
						title="Уменьшить (−)"
					>
						<ZoomOut size={20} />
					</button>
					<span class="text-gray-400 text-sm min-w-[4rem] text-center">{Math.round(zoom * 100)}%</span>
					<button
						onclick={zoomIn}
						class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
						title="Увеличить (+)"
					>
						<ZoomIn size={20} />
					</button>
					<button
						onclick={rotate}
						class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
						title="Повернуть"
					>
						<RotateCw size={20} />
					</button>
					<div class="w-px h-6 bg-gray-700 mx-2"></div>
				{/if}
				{#if ondownload}
					<button
						onclick={ondownload}
						class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
						title="Скачать"
					>
						<Download size={20} />
					</button>
				{/if}
				<button
					onclick={onclose}
					class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
					title="Закрыть (Esc)"
				>
					<X size={20} />
				</button>
			</div>
		</div>

		<!-- Content -->
		<div class="bg-gray-800 rounded-b-lg overflow-auto" style="max-height: calc(90vh - 60px);">
			{#if isImage}
				<div class="flex items-center justify-center p-4 min-h-[300px]">
					<img
						src={dataUrl}
						alt={name}
						style="transform: scale({zoom}) rotate({rotation}deg); transition: transform 0.2s;"
						class="max-w-full max-h-[80vh] object-contain"
					/>
				</div>
			{:else if isPdf}
				<div class="w-full h-[80vh]">
					<iframe
						src={dataUrl}
						class="w-full h-full"
						title={name}
					></iframe>
				</div>
			{:else if isText}
				<div class="p-4">
					<pre class="text-gray-200 text-sm font-mono whitespace-pre-wrap break-words max-h-[70vh] overflow-auto">{textContent()}</pre>
				</div>
			{:else if isVideo}
				<div class="flex items-center justify-center p-4">
					<video
						src={dataUrl}
						controls
						class="max-w-full max-h-[80vh]"
					>
						<track kind="captions" />
					</video>
				</div>
			{:else if isAudio}
				<div class="flex items-center justify-center p-8">
					<audio src={dataUrl} controls class="w-full max-w-md">
						<track kind="captions" />
					</audio>
				</div>
			{:else}
				<div class="flex flex-col items-center justify-center p-8 text-center">
					<div class="text-6xl mb-4">📄</div>
					<p class="text-gray-300 mb-2">Предпросмотр недоступен</p>
					<p class="text-gray-500 text-sm mb-4">Тип файла: {contentType}</p>
					{#if ondownload}
						<button
							onclick={ondownload}
							class="flex items-center gap-2 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
						>
							<Download size={18} />
							Скачать файл
						</button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>
