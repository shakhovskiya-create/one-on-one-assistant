<script lang="ts">
	import { onMount } from 'svelte';
	import { subordinates } from '$lib/stores/auth';

	interface Question {
		text: string;
		checked: boolean;
		notes: string;
	}

	interface Section {
		id: string;
		title: string;
		duration: number;
		questions: Question[];
		expanded: boolean;
	}

	let sections: Section[] = $state([]);
	let selectedEmployee = $state('');
	let currentSection = $state(0);
	let timer = $state(0);
	let totalTime = $state(0);
	let isRunning = $state(false);
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	const currentSectionData = $derived(sections[currentSection]);
	const sectionTimeLimit = $derived((currentSectionData?.duration || 0) * 60);
	const isOverTime = $derived(timer > sectionTimeLimit);

	function getDefaultScript(): Section[] {
		return [
			{
				id: 'checkin',
				title: 'Чекин',
				duration: 5,
				expanded: true,
				questions: [
					{ text: 'Как ты? Что нового?', checked: false, notes: '' },
					{ text: 'Как прошла неделя?', checked: false, notes: '' },
					{ text: 'Что занимает голову прямо сейчас?', checked: false, notes: '' },
					{ text: 'Как настроение команды?', checked: false, notes: '' },
				],
			},
			{
				id: 'employee_agenda',
				title: 'Повестка сотрудника',
				duration: 20,
				expanded: false,
				questions: [
					{ text: 'С чем пришел? Что хочешь обсудить?', checked: false, notes: '' },
					{ text: 'Где нужна помощь или ресурс?', checked: false, notes: '' },
					{ text: 'Что буксует и почему?', checked: false, notes: '' },
					{ text: 'Что мешает команде работать эффективнее?', checked: false, notes: '' },
					{ text: 'Какие решения ожидаются?', checked: false, notes: '' },
				],
			},
			{
				id: 'manager_agenda',
				title: 'Повестка руководителя',
				duration: 15,
				expanded: false,
				questions: [
					{ text: 'Статус по ключевым проектам', checked: false, notes: '' },
					{ text: 'Изменения в приоритетах', checked: false, notes: '' },
					{ text: 'Ожидания и сроки', checked: false, notes: '' },
					{ text: 'Обратная связь от смежных подразделений', checked: false, notes: '' },
				],
			},
			{
				id: 'development',
				title: 'Развитие сотрудника',
				duration: 10,
				expanded: false,
				questions: [
					{ text: 'Как оцениваешь свою работу за последние 2 недели?', checked: false, notes: '' },
					{ text: 'Что получилось хорошо?', checked: false, notes: '' },
					{ text: 'Что бы сделал иначе?', checked: false, notes: '' },
					{ text: 'Чему хочешь научиться?', checked: false, notes: '' },
					{ text: 'Какая поддержка нужна для роста?', checked: false, notes: '' },
				],
			},
			{
				id: 'feedback',
				title: 'Обратная связь руководителю',
				duration: 5,
				expanded: false,
				questions: [
					{ text: 'Что я мог бы делать иначе?', checked: false, notes: '' },
					{ text: 'Достаточно ли контекста и информации ты получаешь?', checked: false, notes: '' },
					{ text: 'Есть что-то, что хотел сказать, но не решался?', checked: false, notes: '' },
				],
			},
			{
				id: 'agreements',
				title: 'Договоренности',
				duration: 5,
				expanded: false,
				questions: [
					{ text: 'Фиксируем договоренности и сроки', checked: false, notes: '' },
				],
			},
		];
	}

	onMount(() => {
		sections = getDefaultScript();

		return () => {
			if (timerInterval) clearInterval(timerInterval);
		};
	});

	$effect(() => {
		if (isRunning && !timerInterval) {
			timerInterval = setInterval(() => {
				timer++;
				totalTime++;
			}, 1000);
		} else if (!isRunning && timerInterval) {
			clearInterval(timerInterval);
			timerInterval = null;
		}
	});

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	}

	function toggleSection(index: number) {
		sections = sections.map((s, i) => ({
			...s,
			expanded: i === index ? !s.expanded : s.expanded,
		}));
		currentSection = index;
		timer = 0;
	}

	function toggleQuestion(sectionIndex: number, questionIndex: number) {
		sections = sections.map((s, si) =>
			si === sectionIndex
				? {
						...s,
						questions: s.questions.map((q, qi) =>
							qi === questionIndex ? { ...q, checked: !q.checked } : q
						),
				  }
				: s
		);
	}

	function updateNotes(sectionIndex: number, questionIndex: number, notes: string) {
		sections = sections.map((s, si) =>
			si === sectionIndex
				? {
						...s,
						questions: s.questions.map((q, qi) =>
							qi === questionIndex ? { ...q, notes } : q
						),
				  }
				: s
		);
	}

	function nextSection() {
		if (currentSection < sections.length - 1) {
			sections = sections.map((s, i) => ({
				...s,
				expanded: i === currentSection + 1,
			}));
			currentSection++;
			timer = 0;
		}
	}

	function resetMeeting() {
		sections = getDefaultScript();
		currentSection = 0;
		timer = 0;
		totalTime = 0;
		isRunning = false;
	}

	function getSectionProgress(section: Section): number {
		const checked = section.questions.filter(q => q.checked).length;
		return Math.round((checked / section.questions.length) * 100);
	}
</script>

<svelte:head>
	<title>Скрипт встречи - EKF Team Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-2xl font-bold text-gray-900">Скрипт встречи 1-на-1</h1>
		<select
			bind:value={selectedEmployee}
			class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
		>
			<option value="">Выберите сотрудника</option>
			{#each $subordinates as emp}
				<option value={emp.id}>{emp.name}</option>
			{/each}
		</select>
	</div>

	<!-- Timer Panel -->
	<div class="bg-white rounded-lg shadow-sm border p-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-6">
				<div class="text-center">
					<p class="text-sm text-gray-500">Секция</p>
					<p class="text-3xl font-mono font-bold {isOverTime ? 'text-red-600' : 'text-gray-900'}">
						{formatTime(timer)}
					</p>
					<p class="text-xs text-gray-400">из {currentSectionData?.duration || 0} мин</p>
				</div>
				<div class="text-center border-l pl-6">
					<p class="text-sm text-gray-500">Всего</p>
					<p class="text-3xl font-mono font-bold text-gray-900">{formatTime(totalTime)}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<button
					onclick={() => isRunning = !isRunning}
					class="p-3 rounded-full transition-colors
						{isRunning
							? 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200'
							: 'bg-green-100 text-green-700 hover:bg-green-200'}"
				>
					{#if isRunning}
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					{:else}
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					{/if}
				</button>
				<button
					onclick={resetMeeting}
					class="p-3 rounded-full bg-gray-100 text-gray-700 hover:bg-gray-200"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
				</button>
			</div>
		</div>
	</div>

	<!-- Sections -->
	<div class="space-y-4">
		{#each sections as section, sectionIndex}
			<div class="bg-white rounded-lg shadow-sm border overflow-hidden
				{sectionIndex === currentSection ? 'ring-2 ring-ekf-red' : ''}">
				<button
					onclick={() => toggleSection(sectionIndex)}
					class="w-full p-4 flex items-center justify-between hover:bg-gray-50"
				>
					<div class="flex items-center gap-3">
						<svg class="w-5 h-5 transform transition-transform {section.expanded ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
						<span class="font-semibold">{section.title}</span>
						<span class="text-sm text-gray-500">({section.duration} мин)</span>
					</div>
					<div class="flex items-center gap-3">
						<div class="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
							<div
								class="h-full bg-green-500 transition-all"
								style="width: {getSectionProgress(section)}%"
							></div>
						</div>
						<span class="text-sm text-gray-500 w-10">{getSectionProgress(section)}%</span>
					</div>
				</button>

				{#if section.expanded}
					<div class="border-t p-4 space-y-3">
						{#each section.questions as question, questionIndex}
							<div class="space-y-2">
								<label class="flex items-start gap-3 cursor-pointer">
									<input
										type="checkbox"
										checked={question.checked}
										onchange={() => toggleQuestion(sectionIndex, questionIndex)}
										class="mt-1 h-5 w-5 rounded border-gray-300 text-ekf-red focus:ring-ekf-red"
									/>
									<span class="flex-1 {question.checked ? 'text-gray-400 line-through' : 'text-gray-700'}">
										{question.text}
									</span>
								</label>
								<textarea
									value={question.notes}
									oninput={(e) => updateNotes(sectionIndex, questionIndex, (e.target as HTMLTextAreaElement).value)}
									placeholder="Заметки..."
									class="w-full ml-8 p-2 text-sm border rounded resize-none focus:ring-1 focus:ring-ekf-red focus:border-ekf-red"
									rows="2"
								></textarea>
							</div>
						{/each}

						{#if sectionIndex < sections.length - 1}
							<button
								onclick={nextSection}
								class="mt-4 w-full py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 flex items-center justify-center gap-2"
							>
								Следующая секция
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
								</svg>
							</button>
						{/if}
					</div>
				{/if}
			</div>
		{/each}
	</div>

	<!-- Progress Overview -->
	<div class="bg-white rounded-lg shadow-sm border p-4">
		<h3 class="font-semibold mb-3">Прогресс встречи</h3>
		<div class="flex gap-2">
			{#each sections as section, index}
				<div
					class="flex-1 h-2 rounded-full
						{getSectionProgress(section) === 100 ? 'bg-green-500' :
						index === currentSection ? 'bg-ekf-red' : 'bg-gray-200'}"
				></div>
			{/each}
		</div>
		<div class="flex justify-between mt-2 text-xs text-gray-500">
			{#each sections as section}
				<span>{section.title.split(' ')[0]}</span>
			{/each}
		</div>
	</div>
</div>
