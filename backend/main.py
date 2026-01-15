import os
import json
import tempfile
import asyncio
import base64
from datetime import datetime, date, timedelta
from typing import Optional, List
from fastapi import FastAPI, UploadFile, File, HTTPException, BackgroundTasks, Query, Form, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import uuid
from pydantic import BaseModel
from openai import OpenAI
from anthropic import Anthropic
from supabase import create_client, Client
import httpx
import uvicorn

app = FastAPI(title="Meeting Assistant API", version="3.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Clients
openai_client: Optional[OpenAI] = None
anthropic_client: Optional[Anthropic] = None
supabase: Optional[Client] = None

# API Keys
TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")
YANDEX_API_KEY = os.getenv("YANDEX_API_KEY", "")
YANDEX_FOLDER_ID = os.getenv("YANDEX_FOLDER_ID", "")
CONNECTOR_API_KEY = os.getenv("CONNECTOR_API_KEY", "")


# ============ ON-PREM CONNECTOR ============

class ConnectorManager:
    """Manages WebSocket connection to on-prem connector"""

    def __init__(self):
        self.connector: Optional[WebSocket] = None
        self.pending_requests: dict = {}
        self.connected = False

    async def connect(self, websocket: WebSocket, api_key: str):
        """Accept connector connection"""
        if CONNECTOR_API_KEY and api_key != CONNECTOR_API_KEY:
            await websocket.close(code=4001, reason="Invalid API key")
            return False

        await websocket.accept()
        self.connector = websocket
        self.connected = True
        print("On-prem connector connected")
        return True

    def disconnect(self):
        """Handle connector disconnect"""
        self.connector = None
        self.connected = False
        # Fail all pending requests
        for request_id, future in self.pending_requests.items():
            if not future.done():
                future.set_exception(Exception("Connector disconnected"))
        self.pending_requests.clear()
        print("On-prem connector disconnected")

    async def send_command(self, command: str, params: dict = None, timeout: float = 30.0) -> dict:
        """Send command to connector and wait for response"""
        if not self.connected or not self.connector:
            raise HTTPException(status_code=503, detail="On-prem connector not connected")

        request_id = str(uuid.uuid4())

        message = {
            "command": command,
            "request_id": request_id,
            "params": params or {}
        }

        # Create future for response
        loop = asyncio.get_event_loop()
        future = loop.create_future()
        self.pending_requests[request_id] = future

        try:
            await self.connector.send_json(message)

            # Wait for response with timeout
            result = await asyncio.wait_for(future, timeout=timeout)
            return result

        except asyncio.TimeoutError:
            self.pending_requests.pop(request_id, None)
            raise HTTPException(status_code=504, detail="Connector request timeout")
        except Exception as e:
            self.pending_requests.pop(request_id, None)
            raise HTTPException(status_code=500, detail=str(e))

    def handle_response(self, data: dict):
        """Handle response from connector"""
        request_id = data.get("request_id")
        if request_id and request_id in self.pending_requests:
            future = self.pending_requests.pop(request_id)
            if not future.done():
                if data.get("success"):
                    future.set_result(data.get("result"))
                else:
                    future.set_exception(Exception(data.get("error", "Unknown error")))


connector_manager = ConnectorManager()

def init_clients():
    global openai_client, anthropic_client, supabase
    
    if os.getenv("OPENAI_API_KEY"):
        openai_client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))
    
    if os.getenv("ANTHROPIC_API_KEY"):
        anthropic_client = Anthropic(api_key=os.getenv("ANTHROPIC_API_KEY"))
    
    if os.getenv("SUPABASE_URL") and os.getenv("SUPABASE_KEY"):
        supabase = create_client(
            os.getenv("SUPABASE_URL"),
            os.getenv("SUPABASE_KEY")
        )

@app.on_event("startup")
async def startup():
    init_clients()


# ============ MODELS ============

class Employee(BaseModel):
    id: Optional[str] = None
    name: str
    position: str
    meeting_frequency: str = "weekly"
    meeting_day: Optional[str] = None
    development_priorities: Optional[str] = None
    created_at: Optional[datetime] = None

class ProjectCreate(BaseModel):
    name: str
    description: Optional[str] = None
    status: str = "active"
    start_date: Optional[str] = None
    end_date: Optional[str] = None

class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None

class MeetingCreate(BaseModel):
    title: Optional[str] = None
    employee_id: Optional[str] = None  # Для 1-на-1
    project_id: Optional[str] = None
    category_code: str = "team_meeting"
    date: str
    participant_ids: List[str] = []

class Agreement(BaseModel):
    id: Optional[str] = None
    meeting_id: str
    task: str
    responsible: str
    deadline: Optional[date] = None
    status: str = "pending"

class TaskCreate(BaseModel):
    title: str
    description: Optional[str] = None
    status: str = "backlog"
    priority: int = 3
    flag_color: Optional[str] = None
    assignee_id: Optional[str] = None
    co_assignee_id: Optional[str] = None
    creator_id: Optional[str] = None
    meeting_id: Optional[str] = None
    project_id: Optional[str] = None
    parent_id: Optional[str] = None
    is_epic: bool = False
    due_date: Optional[str] = None
    tags: List[str] = []

class TaskUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    priority: Optional[int] = None
    flag_color: Optional[str] = None
    assignee_id: Optional[str] = None
    co_assignee_id: Optional[str] = None
    parent_id: Optional[str] = None
    is_epic: Optional[bool] = None
    due_date: Optional[str] = None

class TaskLinkCreate(BaseModel):
    source_task_id: str
    target_task_id: str
    link_type: str

class TagCreate(BaseModel):
    name: str
    color: str = "gray"

class TaskCommentCreate(BaseModel):
    task_id: str
    author_id: str
    content: str

class TelegramLinkRequest(BaseModel):
    employee_id: str
    telegram_username: str


# ============ PROMPTS BY CATEGORY ============

PROMPTS = {
    "one_on_one": """Ты - опытный HR-аналитик и коуч руководителей. Анализируешь транскрипт встречи 1-на-1.

КОНТЕКСТ СОТРУДНИКА:
{employee_context}

ИСТОРИЯ ПРЕДЫДУЩИХ ВСТРЕЧ:
{meetings_history}

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
{transcript}

ФОРМАТ ОТВЕТА (строго JSON):
{{
    "summary": "2-3 предложения: главное из встречи",
    "employee_agenda": ["темы сотрудника"],
    "manager_agenda": ["темы руководителя"],
    "agreements": [
        {{"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}}
    ],
    "development_notes": "наблюдения о росте и навыках",
    "red_flags": {{
        "burnout_signs": "описание или false",
        "turnover_risk": "low/medium/high",
        "turnover_reason": "причина если не low",
        "team_conflicts": "описание или false",
        "concerns": ["другие сигналы"]
    }},
    "mood_score": 7,
    "mood_trend": "improving/stable/declining",
    "mood_indicators": ["на чём основана оценка"],
    "comparison_with_previous": "что изменилось с прошлой встречи",
    "positive_signals": ["позитивные моменты"],
    "recommendations": ["рекомендации для руководителя"],
    "questions_to_ask": ["вопросы на следующую встречу"],
    "employee_profile_update": "новая информация для досье сотрудника"
}}""",

    "team_meeting": """Ты - ассистент проектного менеджера. Анализируешь совещание команды.

КОНТЕКСТ ПРОЕКТА:
{project_context}

ПРЕДЫДУЩИЕ ВСТРЕЧИ ПО ПРОЕКТУ:
{meetings_history}

УЧАСТНИКИ:
{participants}

ТРАНСКРИПТ:
{transcript}

ФОРМАТ ОТВЕТА (JSON):
{{
    "summary": "краткое резюме совещания",
    "decisions": ["принятые решения"],
    "action_items": [
        {{"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}}
    ],
    "blockers": ["выявленные блокеры"],
    "risks": ["риски проекта"],
    "open_questions": ["нерешённые вопросы"],
    "next_steps": ["следующие шаги"],
    "project_health": "green/yellow/red",
    "health_reason": "почему такая оценка"
}}""",

    "planning": """Ты - Scrum-мастер. Анализируешь планирование.

КОНТЕКСТ ПРОЕКТА:
{project_context}

ТРАНСКРИПТ:
{transcript}

ФОРМАТ ОТВЕТА (JSON):
{{
    "summary": "итоги планирования",
    "sprint_goal": "цель спринта",
    "committed_items": [
        {{"task": "задача", "responsible": "кто", "estimate": "оценка", "priority": "high/medium/low"}}
    ],
    "capacity_concerns": ["проблемы с ресурсами"],
    "dependencies": ["зависимости"],
    "risks": ["риски спринта"],
    "team_confidence": 1-10,
    "recommendations": ["рекомендации"]
}}""",

    "retro": """Ты - фасилитатор ретроспектив. Анализируешь ретро команды.

КОНТЕКСТ ПРОЕКТА:
{project_context}

ИСТОРИЯ ПРЕДЫДУЩИХ РЕТРО:
{meetings_history}

ТРАНСКРИПТ:
{transcript}

ФОРМАТ ОТВЕТА (JSON):
{{
    "summary": "итоги ретро",
    "went_well": ["что было хорошо"],
    "went_wrong": ["что пошло не так"],
    "action_items": [
        {{"improvement": "улучшение", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}}
    ],
    "recurring_issues": ["повторяющиеся проблемы из прошлых ретро"],
    "resolved_issues": ["решённые проблемы"],
    "team_morale": 1-10,
    "morale_trend": "improving/stable/declining",
    "patterns": ["паттерны которые стоит отметить"]
}}""",

    "interview": """Ты - HR-эксперт. Анализируешь собеседование.

ТРАНСКРИПТ:
{transcript}

ФОРМАТ ОТВЕТА (JSON):
{{
    "summary": "общее впечатление",
    "candidate_strengths": ["сильные стороны"],
    "candidate_weaknesses": ["слабые стороны"],
    "technical_assessment": {{
        "score": 1-10,
        "details": "детали оценки"
    }},
    "soft_skills_assessment": {{
        "score": 1-10,
        "details": "детали"
    }},
    "culture_fit": {{
        "score": 1-10,
        "details": "детали"
    }},
    "red_flags": ["тревожные сигналы"],
    "questions_answered_well": ["хорошие ответы"],
    "questions_answered_poorly": ["слабые ответы"],
    "recommendation": "hire/no_hire/maybe",
    "recommendation_reason": "обоснование",
    "suggested_next_steps": ["следующие шаги"]
}}""",

    "default": """Проанализируй транскрипт встречи.

ТРАНСКРИПТ:
{transcript}

ФОРМАТ ОТВЕТА (JSON):
{{
    "summary": "краткое резюме",
    "key_points": ["ключевые моменты"],
    "action_items": [
        {{"task": "задача", "responsible": "кто", "deadline": "YYYY-MM-DD или null"}}
    ],
    "decisions": ["решения"],
    "open_questions": ["открытые вопросы"]
}}"""
}

TRANSCRIPT_MERGE_PROMPT = """У тебя есть два варианта транскрипции одной и той же аудиозаписи встречи.

ТРАНСКРИПТ 1 (Whisper/OpenAI):
{whisper_transcript}

ТРАНСКРИПТ 2 (Yandex SpeechKit):
{yandex_transcript}

ЗАДАЧА:
1. Объедини оба транскрипта в один качественный текст
2. Исправь ошибки распознавания, сравнивая версии
3. Выбирай более логичный вариант при расхождениях
4. Сохрани структуру диалога
5. Если одна версия явно лучше - используй её как основу

Верни ТОЛЬКО объединённый транскрипт, без комментариев."""


# ============ TELEGRAM HELPER ============

async def send_telegram_message(chat_id: int, text: str):
    if not TELEGRAM_BOT_TOKEN:
        return False
    
    url = f"https://api.telegram.org/bot{TELEGRAM_BOT_TOKEN}/sendMessage"
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(url, json={
                "chat_id": chat_id,
                "text": text,
                "parse_mode": "HTML"
            })
            return response.status_code == 200
        except Exception as e:
            print(f"Telegram error: {e}")
            return False


# ============ TRANSCRIPTION SERVICES ============

async def transcribe_whisper(file_path: str) -> str:
    """Транскрипция через OpenAI Whisper"""
    if not openai_client:
        return ""
    
    try:
        with open(file_path, "rb") as audio_file:
            transcript = openai_client.audio.transcriptions.create(
                model="whisper-1",
                file=audio_file,
                language="ru",
                response_format="text",
                prompt="Это рабочая встреча. Обсуждаются проекты, задачи, KPI, спринты, дедлайны."
            )
        return transcript
    except Exception as e:
        print(f"Whisper error: {e}")
        return ""


async def transcribe_yandex(file_path: str) -> str:
    """Транскрипция через Yandex SpeechKit"""
    if not YANDEX_API_KEY or not YANDEX_FOLDER_ID:
        return ""
    
    try:
        # Читаем аудио файл
        with open(file_path, "rb") as f:
            audio_data = f.read()
        
        # Определяем формат
        ext = os.path.splitext(file_path)[1].lower()
        audio_format = "oggopus"
        if ext in [".mp3"]:
            audio_format = "mp3"
        elif ext in [".wav"]:
            audio_format = "lpcm"
        
        # Синхронное распознавание (для коротких записей до 30 сек)
        # Для длинных нужно использовать асинхронное API
        
        url = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
        headers = {
            "Authorization": f"Api-Key {YANDEX_API_KEY}",
        }
        params = {
            "folderId": YANDEX_FOLDER_ID,
            "lang": "ru-RU",
            "format": audio_format,
            "sampleRateHertz": 48000,
        }
        
        async with httpx.AsyncClient(timeout=120) as client:
            response = await client.post(
                url,
                headers=headers,
                params=params,
                content=audio_data
            )
            
            if response.status_code == 200:
                result = response.json()
                return result.get("result", "")
            else:
                print(f"Yandex STT error: {response.status_code} - {response.text}")
                return ""
                
    except Exception as e:
        print(f"Yandex error: {e}")
        return ""


async def merge_transcripts(whisper: str, yandex: str) -> str:
    """Объединение транскриптов через Claude"""
    if not anthropic_client:
        return whisper or yandex
    
    if not whisper:
        return yandex
    if not yandex:
        return whisper
    
    try:
        prompt = TRANSCRIPT_MERGE_PROMPT.format(
            whisper_transcript=whisper,
            yandex_transcript=yandex
        )
        
        message = anthropic_client.messages.create(
            model="claude-sonnet-4-20250514",
            max_tokens=8000,
            messages=[{"role": "user", "content": prompt}]
        )
        
        return message.content[0].text
    except Exception as e:
        print(f"Merge error: {e}")
        return whisper or yandex


# ============ CONTEXT BUILDERS ============

def get_employee_context(employee_id: str) -> str:
    """Собираем полное досье на сотрудника"""
    if not supabase:
        return "Контекст недоступен"
    
    # Основная инфо
    emp = supabase.table("employees").select("*").eq("id", employee_id).single().execute()
    if not emp.data:
        return "Сотрудник не найден"
    
    context = f"""
ИМЯ: {emp.data.get('name')}
ДОЛЖНОСТЬ: {emp.data.get('position')}
ПРИОРИТЕТЫ РАЗВИТИЯ: {emp.data.get('development_priorities') or 'не указаны'}
"""
    
    # Статистика по задачам
    tasks = supabase.table("tasks").select("status, due_date").eq("assignee_id", employee_id).execute()
    if tasks.data:
        total = len(tasks.data)
        done = len([t for t in tasks.data if t["status"] == "done"])
        in_progress = len([t for t in tasks.data if t["status"] == "in_progress"])
        today = date.today().isoformat()
        overdue = len([t for t in tasks.data if t.get("due_date") and t["due_date"] < today and t["status"] != "done"])
        
        context += f"""
СТАТИСТИКА ЗАДАЧ:
- Всего: {total}
- Выполнено: {done}
- В работе: {in_progress}
- Просрочено: {overdue}
"""
    
    return context


def get_employee_meetings_history(employee_id: str, limit: int = 5) -> str:
    """История встреч сотрудника (все типы)"""
    if not supabase:
        return "История недоступна"
    
    # 1-на-1 встречи
    one_on_ones = supabase.table("meetings")\
        .select("date, summary, mood_score, analysis")\
        .eq("employee_id", employee_id)\
        .order("date", desc=True)\
        .limit(limit)\
        .execute()
    
    # Встречи где участвовал
    participations = supabase.table("meeting_participants")\
        .select("meeting_id, meetings(date, title, summary, category_id, meeting_categories(name))")\
        .eq("employee_id", employee_id)\
        .order("created_at", desc=True)\
        .limit(limit)\
        .execute()
    
    history = []
    
    for m in one_on_ones.data:
        mood = m.get("mood_score", "?")
        summary = m.get("summary", "")[:200]
        flags = ""
        if m.get("analysis") and m["analysis"].get("red_flags"):
            rf = m["analysis"]["red_flags"]
            if rf.get("burnout_signs"):
                flags += " ⚠️ ВЫГОРАНИЕ"
            if rf.get("turnover_risk") in ["medium", "high"]:
                flags += f" ⚠️ РИСК УХОДА: {rf['turnover_risk']}"
        
        history.append(f"[1-на-1 {m['date']}] Настроение: {mood}/10{flags}\n{summary}")
    
    for p in participations.data:
        if p.get("meetings"):
            m = p["meetings"]
            cat = m.get("meeting_categories", {}).get("name", "Встреча")
            history.append(f"[{cat} {m.get('date')}] {m.get('title', '')}\n{m.get('summary', '')[:150]}")
    
    if not history:
        return "Предыдущих встреч не найдено"
    
    return "\n\n".join(history[:limit])


def get_project_context(project_id: str) -> str:
    """Контекст проекта"""
    if not supabase or not project_id:
        return "Проект не указан"
    
    proj = supabase.table("projects").select("*").eq("id", project_id).single().execute()
    if not proj.data:
        return "Проект не найден"
    
    # Статистика задач проекта
    tasks = supabase.table("tasks").select("status").eq("project_id", project_id).execute()
    
    context = f"""
ПРОЕКТ: {proj.data.get('name')}
ОПИСАНИЕ: {proj.data.get('description') or 'нет'}
СТАТУС: {proj.data.get('status')}
ДАТЫ: {proj.data.get('start_date')} - {proj.data.get('end_date') or 'не определено'}
"""
    
    if tasks.data:
        total = len(tasks.data)
        done = len([t for t in tasks.data if t["status"] == "done"])
        progress = int((done / total) * 100) if total > 0 else 0
        context += f"ПРОГРЕСС: {progress}% ({done}/{total} задач выполнено)"
    
    return context


def get_project_meetings_history(project_id: str, limit: int = 5) -> str:
    """История встреч по проекту"""
    if not supabase or not project_id:
        return ""
    
    meetings = supabase.table("meetings")\
        .select("date, title, summary, category_id, meeting_categories(name)")\
        .eq("project_id", project_id)\
        .order("date", desc=True)\
        .limit(limit)\
        .execute()
    
    if not meetings.data:
        return "Предыдущих встреч по проекту нет"
    
    history = []
    for m in meetings.data:
        cat = m.get("meeting_categories", {}).get("name", "")
        history.append(f"[{cat} {m['date']}] {m.get('title', '')}\n{m.get('summary', '')[:200]}")
    
    return "\n\n".join(history)


def get_participants_info(participant_ids: List[str]) -> str:
    """Информация об участниках"""
    if not supabase or not participant_ids:
        return ""
    
    employees = supabase.table("employees")\
        .select("name, position")\
        .in_("id", participant_ids)\
        .execute()
    
    if not employees.data:
        return ""
    
    return "\n".join([f"- {e['name']} ({e['position']})" for e in employees.data])


# ============ API ENDPOINTS ============

@app.get("/")
async def root():
    return {"status": "ok", "service": "Meeting Assistant API", "version": "3.0.0"}

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "openai": openai_client is not None,
        "anthropic": anthropic_client is not None,
        "supabase": supabase is not None,
        "telegram": bool(TELEGRAM_BOT_TOKEN),
        "yandex_stt": bool(YANDEX_API_KEY and YANDEX_FOLDER_ID)
    }


# ============ PROJECTS ============

@app.get("/projects")
async def list_projects(status: Optional[str] = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("projects").select("*")
    if status:
        query = query.eq("status", status)
    
    result = query.order("created_at", desc=True).execute()
    return result.data

@app.post("/projects")
async def create_project(project: ProjectCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("projects").insert(project.dict()).execute()
    return result.data[0]

@app.get("/projects/{project_id}")
async def get_project(project_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("projects").select("*").eq("id", project_id).single().execute()
    
    # Добавляем статистику
    tasks = supabase.table("tasks").select("status").eq("project_id", project_id).execute()
    meetings = supabase.table("meetings").select("id").eq("project_id", project_id).execute()
    
    data = result.data
    data["task_count"] = len(tasks.data)
    data["meeting_count"] = len(meetings.data)
    
    if tasks.data:
        done = len([t for t in tasks.data if t["status"] == "done"])
        data["progress"] = int((done / len(tasks.data)) * 100)
    else:
        data["progress"] = 0
    
    return data

@app.put("/projects/{project_id}")
async def update_project(project_id: str, project: ProjectUpdate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    data = {k: v for k, v in project.dict().items() if v is not None}
    result = supabase.table("projects").update(data).eq("id", project_id).execute()
    return result.data[0]

@app.delete("/projects/{project_id}")
async def delete_project(project_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("projects").delete().eq("id", project_id).execute()
    return {"status": "deleted"}


# ============ MEETING CATEGORIES ============

@app.get("/meeting-categories")
async def list_meeting_categories():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("meeting_categories").select("*").order("name").execute()
    return result.data


# ============ EMPLOYEES ============

@app.get("/employees")
async def list_employees():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("employees").select("*").order("name").execute()
    return result.data

@app.post("/employees")
async def create_employee(employee: Employee):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    data = employee.dict(exclude={'id', 'created_at'})
    result = supabase.table("employees").insert(data).execute()
    return result.data[0]

@app.get("/employees/{employee_id}")
async def get_employee(employee_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("employees").select("*").eq("id", employee_id).single().execute()
    if not result.data:
        raise HTTPException(status_code=404, detail="Employee not found")
    return result.data

@app.put("/employees/{employee_id}")
async def update_employee(employee_id: str, employee: Employee):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    data = employee.dict(exclude={'id', 'created_at'})
    result = supabase.table("employees").update(data).eq("id", employee_id).execute()
    return result.data[0]

@app.delete("/employees/{employee_id}")
async def delete_employee(employee_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("employees").delete().eq("id", employee_id).execute()
    return {"status": "deleted"}

@app.get("/employees/{employee_id}/dossier")
async def get_employee_dossier(employee_id: str):
    """Полное досье сотрудника"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Основная инфо
    emp = supabase.table("employees").select("*").eq("id", employee_id).single().execute()
    if not emp.data:
        raise HTTPException(status_code=404, detail="Employee not found")
    
    # Все 1-на-1
    one_on_ones = supabase.table("meetings")\
        .select("*")\
        .eq("employee_id", employee_id)\
        .order("date", desc=True)\
        .execute()
    
    # Все участия
    participations = supabase.table("meeting_participants")\
        .select("meetings(*)")\
        .eq("employee_id", employee_id)\
        .execute()
    
    # Задачи
    tasks = supabase.table("tasks")\
        .select("*")\
        .eq("assignee_id", employee_id)\
        .execute()
    
    # Mood trend
    mood_history = [
        {"date": m["date"], "score": m["mood_score"]}
        for m in one_on_ones.data if m.get("mood_score")
    ]
    
    # Red flags history
    red_flags = []
    for m in one_on_ones.data:
        if m.get("analysis") and m["analysis"].get("red_flags"):
            flags = m["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") != "low":
                red_flags.append({
                    "date": m["date"],
                    "flags": flags
                })
    
    return {
        "employee": emp.data,
        "one_on_one_count": len(one_on_ones.data),
        "project_meetings_count": len(participations.data),
        "tasks": {
            "total": len(tasks.data),
            "done": len([t for t in tasks.data if t["status"] == "done"]),
            "in_progress": len([t for t in tasks.data if t["status"] == "in_progress"]),
        },
        "mood_history": mood_history,
        "red_flags_history": red_flags,
        "recent_meetings": one_on_ones.data[:5]
    }


# ============ MEETINGS ============

@app.get("/meetings")
async def list_meetings(
    employee_id: Optional[str] = None,
    project_id: Optional[str] = None,
    category_code: Optional[str] = None,
    limit: int = 50
):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("meetings").select(
        "*, employees(name, position), projects(name), meeting_categories(code, name)"
    )
    
    if employee_id:
        query = query.eq("employee_id", employee_id)
    
    if project_id:
        query = query.eq("project_id", project_id)
    
    if category_code:
        # Сначала получаем ID категории
        cat = supabase.table("meeting_categories").select("id").eq("code", category_code).single().execute()
        if cat.data:
            query = query.eq("category_id", cat.data["id"])
    
    result = query.order("date", desc=True).limit(limit).execute()
    return result.data

@app.get("/meetings/{meeting_id}")
async def get_meeting(meeting_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("meetings").select(
        "*, employees(name, position), projects(name), meeting_categories(code, name), agreements(*)"
    ).eq("id", meeting_id).single().execute()
    
    # Получаем участников
    participants = supabase.table("meeting_participants")\
        .select("*, employees(id, name, position)")\
        .eq("meeting_id", meeting_id)\
        .execute()
    
    data = result.data
    data["participants"] = [p["employees"] for p in participants.data]
    
    return data


# ============ MAIN PROCESSING PIPELINE ============

@app.post("/process-meeting")
async def process_meeting(
    category_code: str = Form("one_on_one"),
    employee_id: Optional[str] = Form(None),
    project_id: Optional[str] = Form(None),
    meeting_date: str = Form(...),
    title: Optional[str] = Form(None),
    participant_ids: Optional[str] = Form(None),  # JSON array as string
    file: UploadFile = File(...),
    background_tasks: BackgroundTasks = None
):
    """
    Универсальная обработка встреч любого типа
    """
    if not openai_client or not anthropic_client:
        raise HTTPException(status_code=500, detail="API clients not configured")
    
    # Парсим участников
    participants = []
    if participant_ids:
        try:
            participants = json.loads(participant_ids)
        except:
            pass
    
    # Сохраняем файл
    suffix = os.path.splitext(file.filename)[1]
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name
    
    try:
        # 1. ДВОЙНАЯ ТРАНСКРИПЦИЯ
        whisper_task = transcribe_whisper(tmp_path)
        yandex_task = transcribe_yandex(tmp_path)
        
        whisper_transcript, yandex_transcript = await asyncio.gather(
            whisper_task, yandex_task
        )
        
        # 2. ОБЪЕДИНЕНИЕ ТРАНСКРИПТОВ
        if whisper_transcript and yandex_transcript:
            merged_transcript = await merge_transcripts(whisper_transcript, yandex_transcript)
        else:
            merged_transcript = whisper_transcript or yandex_transcript
        
        if not merged_transcript:
            raise HTTPException(status_code=500, detail="Transcription failed")
        
        # 3. ПОЛУЧАЕМ КАТЕГОРИЮ И ПРОМПТ
        category = None
        if supabase:
            cat_result = supabase.table("meeting_categories")\
                .select("*")\
                .eq("code", category_code)\
                .single()\
                .execute()
            category = cat_result.data
        
        prompt_template = PROMPTS.get(category_code, PROMPTS["default"])
        
        # 4. СОБИРАЕМ КОНТЕКСТ
        context_data = {
            "transcript": merged_transcript,
            "employee_context": "",
            "meetings_history": "",
            "project_context": "",
            "participants": ""
        }
        
        if employee_id:
            context_data["employee_context"] = get_employee_context(employee_id)
            context_data["meetings_history"] = get_employee_meetings_history(employee_id)
        
        if project_id:
            context_data["project_context"] = get_project_context(project_id)
            if not context_data["meetings_history"]:
                context_data["meetings_history"] = get_project_meetings_history(project_id)
        
        if participants:
            context_data["participants"] = get_participants_info(participants)
        
        # 5. АНАЛИЗ
        prompt = prompt_template.format(**context_data)
        
        message = anthropic_client.messages.create(
            model="claude-sonnet-4-20250514",
            max_tokens=8000,
            messages=[{"role": "user", "content": prompt}]
        )
        
        response_text = message.content[0].text
        
        # Парсим JSON
        if "```json" in response_text:
            response_text = response_text.split("```json")[1].split("```")[0]
        elif "```" in response_text:
            response_text = response_text.split("```")[1].split("```")[0]
        
        analysis = json.loads(response_text.strip())
        
        # 6. СОХРАНЯЕМ В БД
        meeting_id = None
        if supabase:
            meeting_data = {
                "title": title or f"{category_code} - {meeting_date}",
                "employee_id": employee_id,
                "project_id": project_id,
                "category_id": category["id"] if category else None,
                "date": meeting_date,
                "transcript_whisper": whisper_transcript,
                "transcript_yandex": yandex_transcript,
                "transcript_merged": merged_transcript,
                "transcript": merged_transcript,  # для совместимости
                "summary": analysis.get("summary", ""),
                "mood_score": analysis.get("mood_score"),
                "analysis": analysis
            }
            
            meeting_result = supabase.table("meetings").insert(meeting_data).execute()
            meeting_id = meeting_result.data[0]["id"]
            
            # Добавляем участников
            if participants:
                for pid in participants:
                    supabase.table("meeting_participants").insert({
                        "meeting_id": meeting_id,
                        "employee_id": pid
                    }).execute()
            
            # Сохраняем договорённости/action items
            action_items = analysis.get("agreements", []) or analysis.get("action_items", [])
            for item in action_items:
                task_text = item.get("task") or item.get("improvement", "")
                if task_text:
                    agreement_data = {
                        "meeting_id": meeting_id,
                        "task": task_text,
                        "responsible": item.get("responsible", ""),
                        "deadline": item.get("deadline"),
                        "status": "pending"
                    }
                    supabase.table("agreements").insert(agreement_data).execute()
                    
                    # Создаём задачу
                    task_data = {
                        "title": task_text,
                        "description": f"Из встречи: {title or category_code} от {meeting_date}",
                        "status": "todo",
                        "priority": 3,
                        "meeting_id": meeting_id,
                        "project_id": project_id,
                        "due_date": item.get("deadline")
                    }
                    
                    # Пытаемся найти ответственного
                    if item.get("responsible") and employee_id:
                        task_data["assignee_id"] = employee_id
                    
                    supabase.table("tasks").insert(task_data).execute()
            
            analysis["meeting_id"] = meeting_id
        
        return {
            "meeting_id": meeting_id,
            "transcript": {
                "whisper": whisper_transcript[:500] + "..." if len(whisper_transcript) > 500 else whisper_transcript,
                "yandex": yandex_transcript[:500] + "..." if len(yandex_transcript) > 500 else yandex_transcript,
                "merged": merged_transcript
            },
            "analysis": analysis
        }
        
    finally:
        os.unlink(tmp_path)


# ============ TASKS ============

@app.get("/tasks")
async def list_tasks(
    assignee_id: Optional[str] = None,
    project_id: Optional[str] = None,
    status: Optional[str] = None,
    parent_id: Optional[str] = None,
    is_epic: Optional[bool] = None,
    include_subtasks: bool = False
):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("tasks").select(
        "*, assignee:employees!tasks_assignee_id_fkey(id, name), "
        "co_assignee:employees!tasks_co_assignee_id_fkey(id, name), "
        "creator:employees!tasks_creator_id_fkey(id, name), "
        "project:projects(id, name)"
    )
    
    if assignee_id:
        query = query.eq("assignee_id", assignee_id)
    
    if project_id:
        query = query.eq("project_id", project_id)
    
    if status:
        query = query.eq("status", status)
    
    if parent_id:
        query = query.eq("parent_id", parent_id)
    elif not include_subtasks:
        query = query.is_("parent_id", "null")
    
    if is_epic is not None:
        query = query.eq("is_epic", is_epic)
    
    result = query.order("created_at", desc=True).execute()
    
    # Get tags
    tasks = result.data
    if tasks:
        task_ids = [t["id"] for t in tasks]
        tags_result = supabase.table("task_tags").select("task_id, tags(*)").in_("task_id", task_ids).execute()
        
        tags_map = {}
        for tt in tags_result.data:
            if tt["task_id"] not in tags_map:
                tags_map[tt["task_id"]] = []
            tags_map[tt["task_id"]].append(tt["tags"])
        
        for task in tasks:
            task["tags"] = tags_map.get(task["id"], [])
    
    return tasks

@app.post("/tasks")
async def create_task(task: TaskCreate, background_tasks: BackgroundTasks = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    data = task.dict(exclude={'tags'})
    if data.get('due_date'):
        data['original_due_date'] = data['due_date']
    
    result = supabase.table("tasks").insert(data).execute()
    task_id = result.data[0]["id"]
    
    # Add tags
    if task.tags:
        for tag_name in task.tags:
            tag_result = supabase.table("tags").select("id").eq("name", tag_name).execute()
            if tag_result.data:
                supabase.table("task_tags").insert({
                    "task_id": task_id,
                    "tag_id": tag_result.data[0]["id"]
                }).execute()
    
    # Notify
    if task.assignee_id and TELEGRAM_BOT_TOKEN and background_tasks:
        background_tasks.add_task(notify_new_task, task_id, task.assignee_id, task.title)
    
    return result.data[0]

@app.get("/tasks/{task_id}")
async def get_task(task_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("tasks").select(
        "*, assignee:employees!tasks_assignee_id_fkey(id, name), "
        "co_assignee:employees!tasks_co_assignee_id_fkey(id, name), "
        "creator:employees!tasks_creator_id_fkey(id, name), "
        "project:projects(id, name)"
    ).eq("id", task_id).single().execute()
    
    task = result.data
    
    # Tags
    tags_result = supabase.table("task_tags").select("tags(*)").eq("task_id", task_id).execute()
    task["tags"] = [tt["tags"] for tt in tags_result.data]
    
    # Subtasks
    if task["is_epic"]:
        subtasks = supabase.table("tasks").select("*").eq("parent_id", task_id).execute()
        task["subtasks"] = subtasks.data
        if subtasks.data:
            done_count = len([s for s in subtasks.data if s["status"] == "done"])
            task["progress"] = int((done_count / len(subtasks.data)) * 100)
        else:
            task["progress"] = 0
    
    # Comments
    comments = supabase.table("task_comments").select("*, author:employees(name)").eq("task_id", task_id).order("created_at").execute()
    task["comments"] = comments.data
    
    # History
    history = supabase.table("task_history").select("*").eq("task_id", task_id).order("created_at", desc=True).limit(20).execute()
    task["history"] = history.data
    
    return task

@app.put("/tasks/{task_id}")
async def update_task(task_id: str, task: TaskUpdate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    current = supabase.table("tasks").select("*").eq("id", task_id).single().execute()
    current_data = current.data
    
    data = {k: v for k, v in task.dict().items() if v is not None}
    
    # History
    for field, new_value in data.items():
        old_value = current_data.get(field)
        if str(old_value) != str(new_value):
            supabase.table("task_history").insert({
                "task_id": task_id,
                "field_name": field,
                "old_value": str(old_value) if old_value else None,
                "new_value": str(new_value) if new_value else None
            }).execute()
    
    if data.get("status") == "done" and current_data.get("status") != "done":
        data["completed_at"] = datetime.now().isoformat()
    
    result = supabase.table("tasks").update(data).eq("id", task_id).execute()
    return result.data[0]

@app.delete("/tasks/{task_id}")
async def delete_task(task_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("tasks").delete().eq("id", task_id).execute()
    return {"status": "deleted"}


# ============ KANBAN ============

@app.get("/kanban")
async def get_kanban(assignee_id: Optional[str] = None, project_id: Optional[str] = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("tasks").select(
        "*, assignee:employees!tasks_assignee_id_fkey(id, name), tags:task_tags(tags(*))"
    ).is_("parent_id", "null")
    
    if assignee_id:
        query = query.eq("assignee_id", assignee_id)
    
    if project_id:
        query = query.eq("project_id", project_id)
    
    result = query.execute()
    
    kanban = {
        "backlog": [],
        "todo": [],
        "in_progress": [],
        "review": [],
        "done": []
    }
    
    for task in result.data:
        status = task.get("status", "backlog")
        if status in kanban:
            task["tags"] = [tt["tags"] for tt in task.get("tags", []) if tt.get("tags")]
            kanban[status].append(task)
    
    for status in kanban:
        kanban[status].sort(key=lambda x: (x.get("priority", 3), x.get("created_at", "")))
    
    return kanban

@app.put("/kanban/move")
async def move_task_kanban(task_id: str, new_status: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    current = supabase.table("tasks").select("status").eq("id", task_id).single().execute()
    old_status = current.data.get("status")
    
    data = {"status": new_status}
    if new_status == "done" and old_status != "done":
        data["completed_at"] = datetime.now().isoformat()
    
    result = supabase.table("tasks").update(data).eq("id", task_id).execute()
    
    if old_status != new_status:
        supabase.table("task_history").insert({
            "task_id": task_id,
            "field_name": "status",
            "old_value": old_status,
            "new_value": new_status
        }).execute()
    
    return result.data[0]


# ============ AGREEMENTS ============

@app.get("/agreements")
async def list_agreements(employee_id: Optional[str] = None, status: Optional[str] = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("agreements").select("*, meetings(date, employee_id, employees(name))")
    
    if status:
        query = query.eq("status", status)
    
    result = query.order("deadline").execute()
    
    if employee_id:
        result.data = [
            a for a in result.data 
            if a.get("meetings", {}).get("employee_id") == employee_id
        ]
    
    return result.data

@app.put("/agreements/{agreement_id}")
async def update_agreement(agreement_id: str, data: dict):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("agreements").update(data).eq("id", agreement_id).execute()
    return result.data[0]


# ============ TAGS ============

@app.get("/tags")
async def list_tags():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("tags").select("*").order("name").execute()
    return result.data

@app.post("/tags")
async def create_tag(tag: TagCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("tags").insert(tag.dict()).execute()
    return result.data[0]


# ============ COMMENTS ============

@app.post("/task-comments")
async def create_comment(comment: TaskCommentCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("task_comments").insert(comment.dict()).execute()
    return result.data[0]


# ============ ANALYTICS ============

@app.get("/analytics/dashboard")
async def get_dashboard():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    employees = supabase.table("employees").select("*").execute()
    projects = supabase.table("projects").select("*").eq("status", "active").execute()
    
    recent_meetings = supabase.table("meetings")\
        .select("*, employees(name), meeting_categories(name)")\
        .order("date", desc=True)\
        .limit(10)\
        .execute()
    
    tasks = supabase.table("tasks").select("status, due_date").execute()
    
    today = date.today().isoformat()
    task_summary = {
        "total": len(tasks.data),
        "done": len([t for t in tasks.data if t["status"] == "done"]),
        "in_progress": len([t for t in tasks.data if t["status"] == "in_progress"]),
        "overdue": len([t for t in tasks.data if t.get("due_date") and t["due_date"] < today and t["status"] != "done"])
    }
    
    # Red flags
    red_flags = []
    for meeting in recent_meetings.data:
        if meeting.get("analysis") and meeting["analysis"].get("red_flags"):
            flags = meeting["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") in ["medium", "high"]:
                red_flags.append({
                    "employee": meeting.get("employees", {}).get("name"),
                    "date": meeting["date"],
                    "flags": flags
                })
    
    return {
        "employees": employees.data,
        "projects": projects.data,
        "recent_meetings": recent_meetings.data,
        "task_summary": task_summary,
        "red_flags": red_flags
    }

@app.get("/analytics/employee/{employee_id}")
async def get_employee_analytics(employee_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    meetings = supabase.table("meetings")\
        .select("date, mood_score, analysis")\
        .eq("employee_id", employee_id)\
        .order("date")\
        .execute()
    
    mood_history = [
        {"date": m["date"], "score": m["mood_score"]}
        for m in meetings.data if m.get("mood_score")
    ]
    
    red_flags_history = []
    for m in meetings.data:
        if m.get("analysis") and m["analysis"].get("red_flags"):
            flags = m["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") != "low":
                red_flags_history.append({
                    "date": m["date"],
                    "flags": flags
                })
    
    tasks = supabase.table("tasks").select("status").eq("assignee_id", employee_id).execute()
    
    return {
        "mood_history": mood_history,
        "red_flags_history": red_flags_history,
        "task_stats": {
            "total": len(tasks.data),
            "done": len([t for t in tasks.data if t["status"] == "done"]),
            "in_progress": len([t for t in tasks.data if t["status"] == "in_progress"])
        },
        "total_meetings": len(meetings.data)
    }


# ============ TELEGRAM ============

@app.post("/telegram/link")
async def link_telegram(data: TelegramLinkRequest):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("telegram_users").upsert({
        "employee_id": data.employee_id,
        "telegram_username": data.telegram_username,
        "telegram_chat_id": 0,
        "notifications_enabled": True
    }).execute()
    
    return {"status": "pending", "message": "Отправьте /start боту"}

@app.post("/telegram/webhook")
async def telegram_webhook(data: dict):
    if not supabase or not TELEGRAM_BOT_TOKEN:
        return {"ok": True}
    
    message = data.get("message", {})
    chat_id = message.get("chat", {}).get("id")
    text = message.get("text", "")
    username = message.get("from", {}).get("username", "")
    
    if text == "/start":
        user = supabase.table("telegram_users").select("*").eq("telegram_username", username).execute()
        
        if user.data:
            supabase.table("telegram_users").update({
                "telegram_chat_id": chat_id
            }).eq("telegram_username", username).execute()
            
            await send_telegram_message(chat_id, "✅ Аккаунт привязан!")
        else:
            await send_telegram_message(chat_id, "❌ Username не найден в системе")
    
    elif text == "/tasks":
        user = supabase.table("telegram_users").select("employee_id").eq("telegram_chat_id", chat_id).execute()
        
        if user.data:
            tasks = supabase.table("tasks").select("title, status, due_date")\
                .eq("assignee_id", user.data[0]["employee_id"])\
                .neq("status", "done")\
                .limit(10).execute()
            
            if tasks.data:
                msg = "📋 <b>Ваши задачи:</b>\n\n"
                for t in tasks.data:
                    emoji = {"backlog": "📝", "todo": "📌", "in_progress": "🔄", "review": "👀"}.get(t["status"], "")
                    msg += f"{emoji} {t['title']}\n"
                await send_telegram_message(chat_id, msg)
            else:
                await send_telegram_message(chat_id, "✨ Нет активных задач!")
    
    return {"ok": True}


# ============ HELPERS ============

async def notify_new_task(task_id: str, assignee_id: str, title: str):
    if not supabase:
        return
    
    user = supabase.table("telegram_users").select("telegram_chat_id")\
        .eq("employee_id", assignee_id)\
        .eq("notifications_enabled", True).execute()
    
    if user.data and user.data[0].get("telegram_chat_id"):
        await send_telegram_message(
            user.data[0]["telegram_chat_id"],
            f"📌 <b>Новая задача:</b>\n\n{title}"
        )


# ============ ON-PREM CONNECTOR WEBSOCKET ============

@app.websocket("/ws/connector")
async def connector_websocket(websocket: WebSocket, token: Optional[str] = None):
    """WebSocket endpoint for on-prem connector"""
    # Get API key from query param or headers
    api_key = token or websocket.headers.get("authorization", "").replace("Bearer ", "")

    if not await connector_manager.connect(websocket, api_key):
        return

    try:
        while True:
            data = await websocket.receive_json()

            if data.get("type") == "heartbeat":
                # Connector heartbeat - just acknowledge
                continue

            if data.get("type") == "response":
                # Response to our command
                connector_manager.handle_response(data)

    except WebSocketDisconnect:
        connector_manager.disconnect()
    except Exception as e:
        print(f"Connector WebSocket error: {e}")
        connector_manager.disconnect()


@app.get("/connector/status")
async def connector_status():
    """Check if on-prem connector is connected"""
    return {
        "connected": connector_manager.connected,
        "pending_requests": len(connector_manager.pending_requests)
    }


# ============ AD INTEGRATION ============

@app.post("/ad/sync")
async def sync_ad_users():
    """Sync users from Active Directory with pagination"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    synced = 0
    total_from_ad = 0
    offset = 0
    batch_size = 100

    # Fetch users in batches
    while True:
        result = await connector_manager.send_command("sync_users", {
            "offset": offset,
            "limit": batch_size
        }, timeout=60.0)

        users = result.get("users", [])
        total_from_ad = result.get("total", 0)

        for user in users:
            if not user.get("email"):
                continue

            # Check if exists
            existing = supabase.table("employees").select("id").eq("email", user["email"]).execute()

            employee_data = {
                "name": user.get("name", ""),
                "email": user.get("email"),
                "position": user.get("title", ""),
                "department": user.get("department", ""),
                "ad_dn": user.get("dn"),
                "manager_dn": user.get("manager_dn"),
                "ad_login": user.get("login")
            }

            if existing.data:
                supabase.table("employees").update(employee_data).eq("id", existing.data[0]["id"]).execute()
            else:
                supabase.table("employees").insert(employee_data).execute()

            synced += 1

        # Check if more pages
        if not result.get("has_more", False):
            break

        offset += batch_size

    # Update manager relationships
    employees = supabase.table("employees").select("id, manager_dn").execute()
    dn_to_id = {e.get("ad_dn"): e["id"] for e in supabase.table("employees").select("id, ad_dn").execute().data if e.get("ad_dn")}

    for emp in employees.data:
        if emp.get("manager_dn") and emp["manager_dn"] in dn_to_id:
            supabase.table("employees").update({
                "manager_id": dn_to_id[emp["manager_dn"]]
            }).eq("id", emp["id"]).execute()

    return {
        "synced": synced,
        "total_from_ad": total_from_ad
    }


@app.post("/ad/authenticate")
async def authenticate_ad_user(username: str = Form(...), password: str = Form(...)):
    """Authenticate user against AD"""
    result = await connector_manager.send_command("authenticate", {
        "username": username,
        "password": password
    })

    if result.get("authenticated"):
        user = result.get("user", {})

        # Find or create employee
        if supabase and user.get("email"):
            existing = supabase.table("employees").select("*").eq("email", user["email"]).execute()

            if existing.data:
                employee = existing.data[0]
            else:
                # Auto-create from AD
                emp_data = {
                    "name": user.get("name", ""),
                    "email": user.get("email"),
                    "position": user.get("title", ""),
                    "ad_dn": user.get("dn"),
                    "ad_login": user.get("login")
                }
                employee = supabase.table("employees").insert(emp_data).execute().data[0]

            # Generate session token (simplified - use proper JWT in production)
            session_token = str(uuid.uuid4())

            return {
                "authenticated": True,
                "employee": employee,
                "token": session_token
            }

    return {"authenticated": False, "error": result.get("error", "Authentication failed")}


@app.get("/ad/subordinates/{employee_id}")
async def get_ad_subordinates(employee_id: str):
    """Get subordinates for a manager from AD"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    # Get manager's DN
    employee = supabase.table("employees").select("ad_dn").eq("id", employee_id).single().execute()
    if not employee.data or not employee.data.get("ad_dn"):
        raise HTTPException(status_code=404, detail="Employee not found or not synced from AD")

    result = await connector_manager.send_command("get_subordinates", {
        "manager_dn": employee.data["ad_dn"]
    })

    return result


# ============ EXCHANGE/CALENDAR INTEGRATION ============

@app.get("/calendar/{employee_id}")
async def get_employee_calendar(
    employee_id: str,
    days_back: int = 7,
    days_forward: int = 30
):
    """Get calendar events from Exchange"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    employee = supabase.table("employees").select("email").eq("id", employee_id).single().execute()
    if not employee.data or not employee.data.get("email"):
        raise HTTPException(status_code=404, detail="Employee email not found")

    result = await connector_manager.send_command("get_calendar", {
        "email": employee.data["email"],
        "days_back": days_back,
        "days_forward": days_forward
    }, timeout=60.0)

    return result


@app.post("/calendar/meeting")
async def create_calendar_meeting(
    organizer_id: str = Form(...),
    subject: str = Form(...),
    start: str = Form(...),
    end: str = Form(...),
    attendee_ids: str = Form("[]"),  # JSON array
    body: str = Form(""),
    location: str = Form("")
):
    """Create meeting in Exchange and sync to app"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    # Get organizer email
    organizer = supabase.table("employees").select("email").eq("id", organizer_id).single().execute()
    if not organizer.data or not organizer.data.get("email"):
        raise HTTPException(status_code=404, detail="Organizer not found")

    # Get attendee emails
    attendee_list = json.loads(attendee_ids)
    attendee_emails = []
    if attendee_list:
        attendees = supabase.table("employees").select("email").in_("id", attendee_list).execute()
        attendee_emails = [a["email"] for a in attendees.data if a.get("email")]

    # Create in Exchange
    result = await connector_manager.send_command("create_meeting", {
        "organizer_email": organizer.data["email"],
        "subject": subject,
        "start": start,
        "end": end,
        "attendees": attendee_emails,
        "body": body,
        "location": location
    }, timeout=30.0)

    if result:
        # Also create in our database
        meeting_data = {
            "title": subject,
            "date": start.split("T")[0],
            "exchange_id": result.get("id"),
            "category_id": None  # Will be set when processed
        }

        meeting = supabase.table("meetings").insert(meeting_data).execute()

        # Add participants
        for aid in attendee_list:
            supabase.table("meeting_participants").insert({
                "meeting_id": meeting.data[0]["id"],
                "employee_id": aid
            }).execute()

        result["app_meeting_id"] = meeting.data[0]["id"]

    return result


@app.get("/calendar/free-slots")
async def find_free_slots(
    attendee_ids: str = Query(...),  # Comma-separated IDs
    duration_minutes: int = Query(60),
    start: str = Query(...),
    end: str = Query(...)
):
    """Find free time slots for all attendees"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    ids = [id.strip() for id in attendee_ids.split(",")]
    attendees = supabase.table("employees").select("email").in_("id", ids).execute()
    emails = [a["email"] for a in attendees.data if a.get("email")]

    if not emails:
        raise HTTPException(status_code=400, detail="No valid attendee emails found")

    result = await connector_manager.send_command("find_free_slots", {
        "emails": emails,
        "duration_minutes": duration_minutes,
        "start": start,
        "end": end
    }, timeout=60.0)

    return result


@app.post("/calendar/sync")
async def sync_calendar_meetings(employee_id: str = Form(...)):
    """Sync meetings from Exchange to app"""
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")

    employee = supabase.table("employees").select("email").eq("id", employee_id).single().execute()
    if not employee.data or not employee.data.get("email"):
        raise HTTPException(status_code=404, detail="Employee email not found")

    events = await connector_manager.send_command("get_calendar", {
        "email": employee.data["email"],
        "days_back": 30,
        "days_forward": 60
    }, timeout=120.0)

    synced = 0
    for event in events:
        if not event.get("id"):
            continue

        # Check if already exists
        existing = supabase.table("meetings").select("id").eq("exchange_id", event["id"]).execute()

        if not existing.data:
            meeting_data = {
                "title": event.get("subject", "Untitled"),
                "date": event.get("start", "").split("T")[0] if event.get("start") else None,
                "exchange_id": event["id"],
                "exchange_data": event
            }

            supabase.table("meetings").insert(meeting_data).execute()
            synced += 1

    return {"synced": synced, "total_events": len(events)}


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
