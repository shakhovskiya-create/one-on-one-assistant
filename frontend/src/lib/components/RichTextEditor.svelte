<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Editor } from '@tiptap/core';
	import StarterKit from '@tiptap/starter-kit';
	import Link from '@tiptap/extension-link';
	import Image from '@tiptap/extension-image';
	import Underline from '@tiptap/extension-underline';
	import TextAlign from '@tiptap/extension-text-align';
	import Placeholder from '@tiptap/extension-placeholder';
	import {
		Bold,
		Italic,
		Underline as UnderlineIcon,
		Strikethrough,
		List,
		ListOrdered,
		Link as LinkIcon,
		Image as ImageIcon,
		AlignLeft,
		AlignCenter,
		AlignRight,
		Undo,
		Redo,
		Quote,
		Code
	} from 'lucide-svelte';

	interface Props {
		content?: string;
		placeholder?: string;
		onchange?: (html: string) => void;
	}

	let { content = '', placeholder = 'Введите текст...', onchange }: Props = $props();

	let element: HTMLElement;
	let editor: Editor | null = $state(null);

	onMount(() => {
		editor = new Editor({
			element: element,
			extensions: [
				StarterKit.configure({
					heading: {
						levels: [1, 2, 3]
					}
				}),
				Link.configure({
					openOnClick: false,
					HTMLAttributes: {
						class: 'text-blue-600 underline'
					}
				}),
				Image.configure({
					HTMLAttributes: {
						class: 'max-w-full h-auto rounded'
					}
				}),
				Underline,
				TextAlign.configure({
					types: ['heading', 'paragraph']
				}),
				Placeholder.configure({
					placeholder: placeholder
				})
			],
			content: content,
			onUpdate: ({ editor }) => {
				onchange?.(editor.getHTML());
			},
			editorProps: {
				attributes: {
					class: 'prose prose-sm max-w-none focus:outline-none min-h-[200px] px-4 py-3'
				}
			}
		});
	});

	onDestroy(() => {
		editor?.destroy();
	});

	function setLink() {
		if (!editor) return;
		const previousUrl = editor.getAttributes('link').href;
		const url = window.prompt('URL ссылки:', previousUrl);

		if (url === null) return;

		if (url === '') {
			editor.chain().focus().extendMarkRange('link').unsetLink().run();
			return;
		}

		editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
	}

	function addImage() {
		if (!editor) return;
		const url = window.prompt('URL изображения:');
		if (url) {
			editor.chain().focus().setImage({ src: url }).run();
		}
	}

	// Update content when prop changes externally
	$effect(() => {
		if (editor && content !== editor.getHTML()) {
			editor.commands.setContent(content);
		}
	});
</script>

<div class="border border-gray-200 rounded-lg overflow-hidden bg-white">
	<!-- Toolbar -->
	<div class="flex flex-wrap items-center gap-1 p-2 border-b border-gray-200 bg-gray-50">
		<!-- Text formatting -->
		<div class="flex items-center gap-0.5 pr-2 border-r border-gray-300">
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleBold().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('bold') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Жирный (Ctrl+B)"
			>
				<Bold size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleItalic().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('italic') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Курсив (Ctrl+I)"
			>
				<Italic size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleUnderline().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('underline') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Подчёркнутый (Ctrl+U)"
			>
				<UnderlineIcon size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleStrike().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('strike') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Зачёркнутый"
			>
				<Strikethrough size={16} />
			</button>
		</div>

		<!-- Lists -->
		<div class="flex items-center gap-0.5 px-2 border-r border-gray-300">
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleBulletList().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('bulletList') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Маркированный список"
			>
				<List size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleOrderedList().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('orderedList') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Нумерованный список"
			>
				<ListOrdered size={16} />
			</button>
		</div>

		<!-- Alignment -->
		<div class="flex items-center gap-0.5 px-2 border-r border-gray-300">
			<button
				type="button"
				onclick={() => editor?.chain().focus().setTextAlign('left').run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive({ textAlign: 'left' }) ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="По левому краю"
			>
				<AlignLeft size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().setTextAlign('center').run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive({ textAlign: 'center' }) ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="По центру"
			>
				<AlignCenter size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().setTextAlign('right').run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive({ textAlign: 'right' }) ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="По правому краю"
			>
				<AlignRight size={16} />
			</button>
		</div>

		<!-- Insert -->
		<div class="flex items-center gap-0.5 px-2 border-r border-gray-300">
			<button
				type="button"
				onclick={setLink}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('link') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Вставить ссылку"
			>
				<LinkIcon size={16} />
			</button>
			<button
				type="button"
				onclick={addImage}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors text-gray-600"
				title="Вставить изображение"
			>
				<ImageIcon size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleBlockquote().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('blockquote') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Цитата"
			>
				<Quote size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().toggleCodeBlock().run()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors {editor?.isActive('codeBlock') ? 'bg-gray-200 text-ekf-red' : 'text-gray-600'}"
				title="Блок кода"
			>
				<Code size={16} />
			</button>
		</div>

		<!-- Undo/Redo -->
		<div class="flex items-center gap-0.5 pl-2">
			<button
				type="button"
				onclick={() => editor?.chain().focus().undo().run()}
				disabled={!editor?.can().undo()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors text-gray-600 disabled:opacity-30 disabled:cursor-not-allowed"
				title="Отменить (Ctrl+Z)"
			>
				<Undo size={16} />
			</button>
			<button
				type="button"
				onclick={() => editor?.chain().focus().redo().run()}
				disabled={!editor?.can().redo()}
				class="p-1.5 rounded hover:bg-gray-200 transition-colors text-gray-600 disabled:opacity-30 disabled:cursor-not-allowed"
				title="Повторить (Ctrl+Y)"
			>
				<Redo size={16} />
			</button>
		</div>
	</div>

	<!-- Editor -->
	<div bind:this={element} class="min-h-[200px] max-h-[400px] overflow-y-auto"></div>
</div>

<style>
	:global(.ProseMirror) {
		min-height: 200px;
		padding: 1rem;
	}

	:global(.ProseMirror p.is-editor-empty:first-child::before) {
		color: #9ca3af;
		content: attr(data-placeholder);
		float: left;
		height: 0;
		pointer-events: none;
	}

	:global(.ProseMirror:focus) {
		outline: none;
	}

	:global(.ProseMirror ul) {
		list-style-type: disc;
		padding-left: 1.5rem;
	}

	:global(.ProseMirror ol) {
		list-style-type: decimal;
		padding-left: 1.5rem;
	}

	:global(.ProseMirror blockquote) {
		border-left: 3px solid #e5e7eb;
		padding-left: 1rem;
		margin: 0.5rem 0;
		color: #6b7280;
	}

	:global(.ProseMirror pre) {
		background: #1f2937;
		color: #e5e7eb;
		padding: 0.75rem 1rem;
		border-radius: 0.375rem;
		font-family: monospace;
		overflow-x: auto;
	}

	:global(.ProseMirror code) {
		background: #f3f4f6;
		padding: 0.125rem 0.25rem;
		border-radius: 0.25rem;
		font-family: monospace;
	}

	:global(.ProseMirror pre code) {
		background: none;
		padding: 0;
	}

	:global(.ProseMirror img) {
		max-width: 100%;
		height: auto;
	}

	:global(.ProseMirror a) {
		color: #2563eb;
		text-decoration: underline;
	}
</style>
