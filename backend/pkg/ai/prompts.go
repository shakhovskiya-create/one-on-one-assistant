package ai

// Prompts contains analysis prompts by meeting category
var Prompts = map[string]string{
	"one_on_one": `Ты - опытный HR-аналитик и коуч руководителей. Анализируешь транскрипт встречи 1-на-1.

КОНТЕКСТ СОТРУДНИКА:
{{.EmployeeContext}}

ИСТОРИЯ ПРЕДЫДУЩИХ ВСТРЕЧ:
{{.MeetingsHistory}}

ЗАДАЧА: Проанализируй новую встречу с учётом всей истории взаимодействия.

ОБРАТИ ОСОБОЕ ВНИМАНИЕ НА:

1. ДИНАМИКА ПО СРАВНЕНИЮ С ПРОШЛЫМИ ВСТРЕЧАМИ:
   - Изменилось ли настроение?
   - Решены ли проблемы, которые поднимались ранее?
   - Есть ли прогресс в развитии?

2. СИГНАЛЫ ВЫГОРАНИЯ:
   - Частые упоминания усталости, перегруза
   - Цинизм, негатив к работе/коллегам
   - Снижение инициативы
   - Проблемы со сном, здоровьем

3. РИСК УХОДА:
   - Упоминание других компаний/предложений
   - Вопросы о росте без энтузиазма
   - Дистанцирование ("они", а не "мы")
   - Недовольство компенсацией

4. СКРЫТЫЕ КОНФЛИКТЫ:
   - Напряжение при упоминании людей
   - Уклончивые ответы о команде

5. ДОГОВОРЁННОСТИ:
   - Извлекай ТОЛЬКО конкретные обязательства
   - КТО и ЧТО должен сделать
   - Если срок не назван - не выдумывай

ТРАНСКРИПТ НОВОЙ ВСТРЕЧИ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (строго JSON):
{
    "summary": "2-3 предложения: главное из встречи",
    "employee_agenda": ["темы сотрудника"],
    "manager_agenda": ["темы руководителя"],
    "agreements": [
        {"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}
    ],
    "development_notes": "наблюдения о росте и навыках",
    "red_flags": {
        "burnout_signs": "описание или false",
        "turnover_risk": "low/medium/high",
        "turnover_reason": "причина если не low",
        "team_conflicts": "описание или false",
        "concerns": ["другие сигналы"]
    },
    "mood_score": 7,
    "mood_trend": "improving/stable/declining",
    "mood_indicators": ["на чём основана оценка"],
    "comparison_with_previous": "что изменилось с прошлой встречи",
    "positive_signals": ["позитивные моменты"],
    "recommendations": ["рекомендации для руководителя"],
    "questions_to_ask": ["вопросы на следующую встречу"],
    "employee_profile_update": "новая информация для досье сотрудника"
}`,

	"team_meeting": `Ты - ассистент проектного менеджера. Анализируешь совещание команды.

КОНТЕКСТ ПРОЕКТА:
{{.ProjectContext}}

ПРЕДЫДУЩИЕ ВСТРЕЧИ ПО ПРОЕКТУ:
{{.MeetingsHistory}}

УЧАСТНИКИ:
{{.Participants}}

ТРАНСКРИПТ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (JSON):
{
    "summary": "краткое резюме совещания",
    "decisions": ["принятые решения"],
    "action_items": [
        {"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}
    ],
    "blockers": ["выявленные блокеры"],
    "risks": ["риски проекта"],
    "open_questions": ["нерешённые вопросы"],
    "next_steps": ["следующие шаги"],
    "project_health": "green/yellow/red",
    "health_reason": "почему такая оценка"
}`,

	"planning": `Ты - Scrum-мастер. Анализируешь планирование.

КОНТЕКСТ ПРОЕКТА:
{{.ProjectContext}}

ТРАНСКРИПТ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (JSON):
{
    "summary": "итоги планирования",
    "sprint_goal": "цель спринта",
    "committed_items": [
        {"task": "задача", "responsible": "кто", "estimate": "оценка", "priority": "high/medium/low"}
    ],
    "capacity_concerns": ["проблемы с ресурсами"],
    "dependencies": ["зависимости"],
    "risks": ["риски спринта"],
    "team_confidence": 8,
    "recommendations": ["рекомендации"]
}`,

	"retro": `Ты - фасилитатор ретроспектив. Анализируешь ретро команды.

КОНТЕКСТ ПРОЕКТА:
{{.ProjectContext}}

ИСТОРИЯ ПРЕДЫДУЩИХ РЕТРО:
{{.MeetingsHistory}}

ТРАНСКРИПТ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (JSON):
{
    "summary": "итоги ретро",
    "went_well": ["что было хорошо"],
    "went_wrong": ["что пошло не так"],
    "action_items": [
        {"improvement": "улучшение", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}
    ],
    "recurring_issues": ["повторяющиеся проблемы из прошлых ретро"],
    "resolved_issues": ["решённые проблемы"],
    "team_morale": 7,
    "morale_trend": "improving/stable/declining",
    "patterns": ["паттерны которые стоит отметить"]
}`,

	"interview": `Ты - HR-эксперт. Анализируешь собеседование.

ТРАНСКРИПТ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (JSON):
{
    "summary": "общее впечатление",
    "candidate_strengths": ["сильные стороны"],
    "candidate_weaknesses": ["слабые стороны"],
    "technical_assessment": {
        "score": 7,
        "details": "детали оценки"
    },
    "soft_skills_assessment": {
        "score": 7,
        "details": "детали"
    },
    "culture_fit": {
        "score": 7,
        "details": "детали"
    },
    "red_flags": ["тревожные сигналы"],
    "questions_answered_well": ["хорошие ответы"],
    "questions_answered_poorly": ["слабые ответы"],
    "recommendation": "hire/no_hire/maybe",
    "recommendation_reason": "обоснование",
    "suggested_next_steps": ["следующие шаги"]
}`,

	"default": `Проанализируй транскрипт встречи.

ТРАНСКРИПТ:
{{.Transcript}}

ФОРМАТ ОТВЕТА (JSON):
{
    "summary": "краткое резюме",
    "key_points": ["ключевые моменты"],
    "action_items": [
        {"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}
    ],
    "decisions": ["решения"],
    "open_questions": ["открытые вопросы"]
}`,
}

// TranscriptMergePrompt is used to merge two transcripts
var TranscriptMergePrompt = `У тебя есть два варианта транскрипции одной и той же аудиозаписи встречи.

ТРАНСКРИПТ 1 (Whisper/OpenAI):
{{.WhisperTranscript}}

ТРАНСКРИПТ 2 (Yandex SpeechKit):
{{.YandexTranscript}}

ЗАДАЧА:
1. Объедини оба транскрипта в один качественный текст
2. Исправь ошибки распознавания, сравнивая версии
3. Выбирай более логичный вариант при расхождениях
4. Сохрани структуру диалога
5. Если одна версия явно лучше - используй её как основу

Верни ТОЛЬКО объединённый транскрипт, без комментариев.`
