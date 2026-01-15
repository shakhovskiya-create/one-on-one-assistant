import os
import json
import tempfile
import asyncio
from datetime import datetime, date, timedelta
from typing import Optional, List
from fastapi import FastAPI, UploadFile, File, HTTPException, BackgroundTasks, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from openai import OpenAI
from anthropic import Anthropic
from supabase import create_client, Client
import httpx
import uvicorn

app = FastAPI(title="1-on-1 Assistant API", version="2.0.0")

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

# Telegram Bot Token
TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")

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

class Meeting(BaseModel):
    id: Optional[str] = None
    employee_id: str
    date: date
    duration_minutes: Optional[int] = None
    mood_score: Optional[int] = None
    transcript: Optional[str] = None
    summary: Optional[str] = None
    analysis: Optional[dict] = None
    created_at: Optional[datetime] = None

class Agreement(BaseModel):
    id: Optional[str] = None
    meeting_id: str
    task: str
    responsible: str
    deadline: Optional[date] = None
    status: str = "pending"

class MeetingAnalysis(BaseModel):
    summary: str
    employee_agenda: List[str]
    manager_agenda: List[str]
    agreements: List[dict]
    development_notes: str
    red_flags: dict
    mood_score: int
    mood_change: Optional[int] = None
    recommendations: List[str]

# ============ TASK MODELS ============

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

class DeadlineRequestCreate(BaseModel):
    task_id: str
    requester_id: str
    requested_date: str
    reason: str

class DeadlineRequestReview(BaseModel):
    status: str
    reviewer_id: str
    review_comment: Optional[str] = None

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

# ============ SCRIPT DATA ============

MEETING_SCRIPT = {
    "sections": [
        {
            "id": "checkin",
            "title": "Чекин",
            "duration": 5,
            "questions": [
                "Как ты? Что нового?",
                "Как прошла неделя?",
                "Что занимает голову прямо сейчас?",
                "Как настроение команды?"
            ]
        },
        {
            "id": "employee_agenda",
            "title": "Повестка сотрудника",
            "duration": 20,
            "questions": [
                "С чем пришел? Что хочешь обсудить?",
                "Где нужна помощь или ресурс?",
                "Что буксует и почему?",
                "Что мешает команде работать эффективнее?",
                "Какие решения ожидаются?"
            ]
        },
        {
            "id": "manager_agenda",
            "title": "Повестка руководителя",
            "duration": 15,
            "questions": [
                "Статус по ключевым проектам",
                "Изменения в приоритетах",
                "Ожидания и сроки",
                "Обратная связь от смежных подразделений"
            ]
        },
        {
            "id": "development",
            "title": "Развитие сотрудника",
            "duration": 10,
            "questions": [
                "Как оцениваешь свою работу за последние 2 недели?",
                "Что получилось хорошо?",
                "Что бы сделал иначе?",
                "Чему хочешь научиться?",
                "Какая поддержка нужна для роста?"
            ]
        },
        {
            "id": "feedback",
            "title": "Обратная связь руководителю",
            "duration": 5,
            "questions": [
                "Что я мог бы делать иначе, чтобы тебе было проще работать?",
                "Достаточно ли контекста и информации ты получаешь?",
                "Есть что-то, что хотел сказать, но не решался?"
            ]
        },
        {
            "id": "agreements",
            "title": "Договоренности",
            "duration": 5,
            "questions": [
                "Фиксируем договоренности и сроки"
            ]
        }
    ]
}

ANALYSIS_PROMPT = """Ты - ассистент руководителя, анализирующий транскрипт встречи 1-на-1.

Проанализируй транскрипт и верни JSON со следующей структурой:
{
    "summary": "Краткое резюме встречи в 2-3 предложениях",
    "employee_agenda": ["Список тем, которые поднял сотрудник"],
    "manager_agenda": ["Список тем, которые поднял руководитель"],
    "agreements": [
        {
            "task": "Описание задачи",
            "responsible": "Кто отвечает (сотрудник/руководитель/имя)",
            "deadline": "Срок в формате YYYY-MM-DD или null"
        }
    ],
    "development_notes": "Заметки по развитию сотрудника",
    "red_flags": {
        "burnout_signs": false или "описание признаков",
        "turnover_risk": "low/medium/high",
        "team_conflicts": false или "описание"
    },
    "mood_score": число от 1 до 10,
    "recommendations": ["Рекомендации для следующей встречи"]
}

Важно:
- Выделяй конкретные договоренности с четкими сроками
- Обращай внимание на признаки выгорания: усталость, цинизм, снижение продуктивности
- Оценивай риск ухода по косвенным признакам: недовольство, упоминание других предложений
- Mood score основывай на тоне, энергии, позитивных/негативных высказываниях

Транскрипт встречи:
{transcript}

Верни только валидный JSON без дополнительного текста."""


# ============ TELEGRAM HELPER ============

async def send_telegram_message(chat_id: int, text: str):
    """Send message via Telegram Bot API"""
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


# ============ API ENDPOINTS ============

@app.get("/")
async def root():
    return {"status": "ok", "service": "1-on-1 Assistant API", "version": "2.0.0"}

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "openai": openai_client is not None,
        "anthropic": anthropic_client is not None,
        "supabase": supabase is not None,
        "telegram": bool(TELEGRAM_BOT_TOKEN)
    }

@app.get("/script")
async def get_script():
    return MEETING_SCRIPT


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

@app.get("/employees/{employee_id}")
async def get_employee(employee_id: str):
    result = supabase.table("employees").select("*").eq("id", employee_id).single().execute()
    if not result.data:
        raise HTTPException(status_code=404, detail="Employee not found")
    return result.data

# ============ MEETINGS ============

@app.get("/meetings")
async def list_meetings(employee_id: Optional[str] = None, limit: int = 50):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("meetings").select("*, employees(name, position)")
    
    if employee_id:
        query = query.eq("employee_id", employee_id)
    
    result = query.order("date", desc=True).limit(limit).execute()
    return result.data

@app.post("/meetings")
async def create_meeting(meeting: Meeting):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    data = meeting.dict(exclude={'id', 'created_at'})
    data['date'] = str(data['date'])
    result = supabase.table("meetings").insert(data).execute()
    return result.data[0]

@app.get("/meetings/{meeting_id}")
async def get_meeting(meeting_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("meetings").select("*, employees(name, position), agreements(*)").eq("id", meeting_id).single().execute()
    return result.data


# ============ TRANSCRIPTION ============

@app.post("/transcribe")
async def transcribe_audio(file: UploadFile = File(...)):
    if not openai_client:
        raise HTTPException(status_code=500, detail="OpenAI not configured")
    
    suffix = os.path.splitext(file.filename)[1]
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name
    
    try:
        with open(tmp_path, "rb") as audio_file:
            transcript = openai_client.audio.transcriptions.create(
                model="whisper-1",
                file=audio_file,
                language="ru",
                response_format="text"
            )
        return {"transcript": transcript}
    finally:
        os.unlink(tmp_path)


# ============ ANALYSIS ============

@app.post("/analyze")
async def analyze_transcript(data: dict):
    if not anthropic_client:
        raise HTTPException(status_code=500, detail="Anthropic not configured")
    
    transcript = data.get("transcript", "")
    if not transcript:
        raise HTTPException(status_code=400, detail="Transcript required")
    
    prompt = ANALYSIS_PROMPT.format(transcript=transcript)
    
    message = anthropic_client.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=4096,
        messages=[{"role": "user", "content": prompt}]
    )
    
    response_text = message.content[0].text
    
    try:
        if "```json" in response_text:
            response_text = response_text.split("```json")[1].split("```")[0]
        elif "```" in response_text:
            response_text = response_text.split("```")[1].split("```")[0]
        
        analysis = json.loads(response_text.strip())
        return analysis
    except json.JSONDecodeError as e:
        raise HTTPException(status_code=500, detail=f"Failed to parse analysis: {str(e)}")


# ============ FULL PROCESSING PIPELINE ============

@app.post("/process-meeting")
async def process_meeting(
    employee_id: str,
    meeting_date: str,
    file: UploadFile = File(...),
    background_tasks: BackgroundTasks = None
):
    if not openai_client or not anthropic_client:
        raise HTTPException(status_code=500, detail="API clients not configured")
    
    # 1. Transcribe
    suffix = os.path.splitext(file.filename)[1]
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name
    
    try:
        with open(tmp_path, "rb") as audio_file:
            transcript = openai_client.audio.transcriptions.create(
                model="whisper-1",
                file=audio_file,
                language="ru",
                response_format="text"
            )
    finally:
        os.unlink(tmp_path)
    
    # 2. Analyze
    prompt = ANALYSIS_PROMPT.format(transcript=transcript)
    
    message = anthropic_client.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=4096,
        messages=[{"role": "user", "content": prompt}]
    )
    
    response_text = message.content[0].text
    
    if "```json" in response_text:
        response_text = response_text.split("```json")[1].split("```")[0]
    elif "```" in response_text:
        response_text = response_text.split("```")[1].split("```")[0]
    
    analysis = json.loads(response_text.strip())
    
    # 3. Get previous meeting mood for comparison
    if supabase:
        prev_meetings = supabase.table("meetings")\
            .select("mood_score")\
            .eq("employee_id", employee_id)\
            .order("date", desc=True)\
            .limit(1)\
            .execute()
        
        if prev_meetings.data:
            prev_mood = prev_meetings.data[0].get("mood_score")
            if prev_mood:
                analysis["mood_change"] = analysis["mood_score"] - prev_mood
    
    # 4. Save to database
    meeting_id = None
    if supabase:
        meeting_data = {
            "employee_id": employee_id,
            "date": meeting_date,
            "transcript": transcript,
            "summary": analysis["summary"],
            "mood_score": analysis["mood_score"],
            "analysis": analysis
        }
        
        meeting_result = supabase.table("meetings").insert(meeting_data).execute()
        meeting_id = meeting_result.data[0]["id"]
        
        # Save agreements and create tasks
        for agreement in analysis.get("agreements", []):
            agreement_data = {
                "meeting_id": meeting_id,
                "task": agreement["task"],
                "responsible": agreement["responsible"],
                "deadline": agreement.get("deadline"),
                "status": "pending"
            }
            supabase.table("agreements").insert(agreement_data).execute()
            
            # Auto-create task from agreement
            task_data = {
                "title": agreement["task"],
                "description": f"Задача из встречи 1-on-1 от {meeting_date}",
                "status": "todo",
                "priority": 3,
                "assignee_id": employee_id,
                "meeting_id": meeting_id,
                "due_date": agreement.get("deadline")
            }
            supabase.table("tasks").insert(task_data).execute()
        
        analysis["meeting_id"] = meeting_id
    
    return {
        "transcript": transcript,
        "analysis": analysis
    }


# ============ AGREEMENTS ============

@app.get("/agreements")
async def list_agreements(
    employee_id: Optional[str] = None,
    status: Optional[str] = None
):
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


# ============ TASKS ============

@app.get("/tasks")
async def list_tasks(
    assignee_id: Optional[str] = None,
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
        "creator:employees!tasks_creator_id_fkey(id, name)"
    )
    
    if assignee_id:
        query = query.or_(f"assignee_id.eq.{assignee_id},co_assignee_id.eq.{assignee_id}")
    
    if status:
        query = query.eq("status", status)
    
    if parent_id:
        query = query.eq("parent_id", parent_id)
    elif not include_subtasks:
        query = query.is_("parent_id", "null")
    
    if is_epic is not None:
        query = query.eq("is_epic", is_epic)
    
    result = query.order("created_at", desc=True).execute()
    
    # Get tags for each task
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
    
    # Get blocking info
    if tasks:
        links_result = supabase.table("task_links").select("*").execute()
        
        blocking_map = {}
        blocked_by_map = {}
        for link in links_result.data:
            if link["link_type"] == "blocks":
                if link["source_task_id"] not in blocking_map:
                    blocking_map[link["source_task_id"]] = []
                blocking_map[link["source_task_id"]].append(link["target_task_id"])
            elif link["link_type"] == "blocked_by":
                if link["source_task_id"] not in blocked_by_map:
                    blocked_by_map[link["source_task_id"]] = []
                blocked_by_map[link["source_task_id"]].append(link["target_task_id"])
        
        for task in tasks:
            task["blocks"] = blocking_map.get(task["id"], [])
            task["blocked_by"] = blocked_by_map.get(task["id"], [])
            task["is_blocking"] = len(task["blocks"]) > 0
    
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
    
    # Notify assignee via Telegram
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
        "creator:employees!tasks_creator_id_fkey(id, name)"
    ).eq("id", task_id).single().execute()
    
    task = result.data
    
    # Get tags
    tags_result = supabase.table("task_tags").select("tags(*)").eq("task_id", task_id).execute()
    task["tags"] = [tt["tags"] for tt in tags_result.data]
    
    # Get subtasks if epic
    if task["is_epic"]:
        subtasks = supabase.table("tasks").select("*").eq("parent_id", task_id).execute()
        task["subtasks"] = subtasks.data
        
        # Calculate progress
        if subtasks.data:
            done_count = len([s for s in subtasks.data if s["status"] == "done"])
            task["progress"] = int((done_count / len(subtasks.data)) * 100)
        else:
            task["progress"] = 0
    
    # Get links
    links = supabase.table("task_links").select("*, target:tasks!task_links_target_task_id_fkey(id, title, status)").eq("source_task_id", task_id).execute()
    task["links"] = links.data
    
    # Get comments
    comments = supabase.table("task_comments").select("*, author:employees(name)").eq("task_id", task_id).order("created_at").execute()
    task["comments"] = comments.data
    
    # Get history
    history = supabase.table("task_history").select("*, changed_by:employees(name)").eq("task_id", task_id).order("created_at", desc=True).limit(20).execute()
    task["history"] = history.data
    
    return task

@app.put("/tasks/{task_id}")
async def update_task(task_id: str, task: TaskUpdate, background_tasks: BackgroundTasks = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get current task for history
    current = supabase.table("tasks").select("*").eq("id", task_id).single().execute()
    current_data = current.data
    
    data = {k: v for k, v in task.dict().items() if v is not None}
    
    # Record history for changed fields
    for field, new_value in data.items():
        old_value = current_data.get(field)
        if str(old_value) != str(new_value):
            supabase.table("task_history").insert({
                "task_id": task_id,
                "field_name": field,
                "old_value": str(old_value) if old_value else None,
                "new_value": str(new_value) if new_value else None
            }).execute()
    
    # If completing task
    if data.get("status") == "done" and current_data.get("status") != "done":
        data["completed_at"] = datetime.now().isoformat()
        
        # Check if parent epic should be completed
        if current_data.get("parent_id"):
            await check_epic_completion(current_data["parent_id"])
    
    result = supabase.table("tasks").update(data).eq("id", task_id).execute()
    return result.data[0]

@app.delete("/tasks/{task_id}")
async def delete_task(task_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("tasks").delete().eq("id", task_id).execute()
    return {"status": "deleted"}


# ============ TASK LINKS (BLOCKING) ============

@app.post("/task-links")
async def create_task_link(link: TaskLinkCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Check for circular dependency
    if link.link_type == "blocks":
        existing = supabase.table("task_links").select("*")\
            .eq("source_task_id", link.target_task_id)\
            .eq("target_task_id", link.source_task_id)\
            .eq("link_type", "blocks").execute()
        
        if existing.data:
            raise HTTPException(status_code=400, detail="Circular dependency detected")
    
    result = supabase.table("task_links").insert(link.dict()).execute()
    
    # If blocking, also create reverse link
    if link.link_type == "blocks":
        supabase.table("task_links").insert({
            "source_task_id": link.target_task_id,
            "target_task_id": link.source_task_id,
            "link_type": "blocked_by"
        }).execute()
    
    return result.data[0]

@app.delete("/task-links/{link_id}")
async def delete_task_link(link_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get link to find reverse
    link = supabase.table("task_links").select("*").eq("id", link_id).single().execute()
    
    if link.data:
        # Delete reverse link
        if link.data["link_type"] == "blocks":
            supabase.table("task_links").delete()\
                .eq("source_task_id", link.data["target_task_id"])\
                .eq("target_task_id", link.data["source_task_id"])\
                .eq("link_type", "blocked_by").execute()
    
    supabase.table("task_links").delete().eq("id", link_id).execute()
    return {"status": "deleted"}


# ============ DEADLINE REQUESTS ============

@app.get("/deadline-requests")
async def list_deadline_requests(status: Optional[str] = "pending"):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("deadline_requests").select(
        "*, task:tasks(id, title), requester:employees!deadline_requests_requester_id_fkey(name)"
    )
    
    if status:
        query = query.eq("status", status)
    
    result = query.order("created_at", desc=True).execute()
    return result.data

@app.post("/deadline-requests")
async def create_deadline_request(request: DeadlineRequestCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get current due date
    task = supabase.table("tasks").select("due_date, title").eq("id", request.task_id).single().execute()
    
    data = request.dict()
    data["old_due_date"] = task.data.get("due_date")
    
    result = supabase.table("deadline_requests").insert(data).execute()
    
    # TODO: Notify manager via Telegram
    
    return result.data[0]

@app.put("/deadline-requests/{request_id}")
async def review_deadline_request(request_id: str, review: DeadlineRequestReview):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get request
    req = supabase.table("deadline_requests").select("*").eq("id", request_id).single().execute()
    
    data = {
        "status": review.status,
        "reviewer_id": review.reviewer_id,
        "review_comment": review.review_comment,
        "reviewed_at": datetime.now().isoformat()
    }
    
    result = supabase.table("deadline_requests").update(data).eq("id", request_id).execute()
    
    # If approved, update task due date
    if review.status == "approved":
        supabase.table("tasks").update({
            "due_date": req.data["requested_date"]
        }).eq("id", req.data["task_id"]).execute()
        
        # Record in history
        supabase.table("task_history").insert({
            "task_id": req.data["task_id"],
            "field_name": "due_date",
            "old_value": req.data["old_due_date"],
            "new_value": req.data["requested_date"],
            "changed_by": review.reviewer_id
        }).execute()
    
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

@app.delete("/tags/{tag_id}")
async def delete_tag(tag_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Check if system tag
    tag = supabase.table("tags").select("is_system").eq("id", tag_id).single().execute()
    if tag.data.get("is_system"):
        raise HTTPException(status_code=400, detail="Cannot delete system tag")
    
    supabase.table("tags").delete().eq("id", tag_id).execute()
    return {"status": "deleted"}

@app.post("/tasks/{task_id}/tags/{tag_id}")
async def add_tag_to_task(task_id: str, tag_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("task_tags").insert({
        "task_id": task_id,
        "tag_id": tag_id
    }).execute()
    return result.data[0]

@app.delete("/tasks/{task_id}/tags/{tag_id}")
async def remove_tag_from_task(task_id: str, tag_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("task_tags").delete().eq("task_id", task_id).eq("tag_id", tag_id).execute()
    return {"status": "deleted"}


# ============ COMMENTS ============

@app.post("/task-comments")
async def create_comment(comment: TaskCommentCreate):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    result = supabase.table("task_comments").insert(comment.dict()).execute()
    return result.data[0]

@app.delete("/task-comments/{comment_id}")
async def delete_comment(comment_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    supabase.table("task_comments").delete().eq("id", comment_id).execute()
    return {"status": "deleted"}


# ============ KANBAN ============

@app.get("/kanban")
async def get_kanban(assignee_id: Optional[str] = None):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    query = supabase.table("tasks").select(
        "*, assignee:employees!tasks_assignee_id_fkey(id, name), tags:task_tags(tags(*))"
    ).is_("parent_id", "null")
    
    if assignee_id:
        query = query.or_(f"assignee_id.eq.{assignee_id},co_assignee_id.eq.{assignee_id}")
    
    result = query.execute()
    
    # Group by status
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
            # Flatten tags
            task["tags"] = [tt["tags"] for tt in task.get("tags", []) if tt.get("tags")]
            kanban[status].append(task)
    
    # Sort by priority within each column
    for status in kanban:
        kanban[status].sort(key=lambda x: (x.get("priority", 3), x.get("created_at", "")))
    
    return kanban

@app.put("/kanban/move")
async def move_task_kanban(task_id: str, new_status: str, position: int = 0):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get current status for history
    current = supabase.table("tasks").select("status").eq("id", task_id).single().execute()
    old_status = current.data.get("status")
    
    data = {"status": new_status}
    
    if new_status == "done" and old_status != "done":
        data["completed_at"] = datetime.now().isoformat()
    
    result = supabase.table("tasks").update(data).eq("id", task_id).execute()
    
    # Record history
    if old_status != new_status:
        supabase.table("task_history").insert({
            "task_id": task_id,
            "field_name": "status",
            "old_value": old_status,
            "new_value": new_status
        }).execute()
    
    return result.data[0]


# ============ TELEGRAM ============

@app.post("/telegram/link")
async def link_telegram(data: TelegramLinkRequest):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # This would be completed via Telegram bot interaction
    # For now, store the pending link
    result = supabase.table("telegram_users").upsert({
        "employee_id": data.employee_id,
        "telegram_username": data.telegram_username,
        "telegram_chat_id": 0,  # Will be set when user messages bot
        "notifications_enabled": True
    }).execute()
    
    return {"status": "pending", "message": f"Отправьте /start боту @one_on_one_ekf_bot"}

@app.post("/telegram/webhook")
async def telegram_webhook(data: dict):
    """Handle Telegram bot updates"""
    if not supabase or not TELEGRAM_BOT_TOKEN:
        return {"ok": True}
    
    message = data.get("message", {})
    chat_id = message.get("chat", {}).get("id")
    text = message.get("text", "")
    username = message.get("from", {}).get("username", "")
    
    if text == "/start":
        # Link account
        user = supabase.table("telegram_users").select("*, employee:employees(name)")\
            .eq("telegram_username", username).execute()
        
        if user.data:
            supabase.table("telegram_users").update({
                "telegram_chat_id": chat_id
            }).eq("telegram_username", username).execute()
            
            await send_telegram_message(
                chat_id,
                f"✅ Аккаунт привязан!\n\nВы будете получать уведомления о задачах."
            )
        else:
            await send_telegram_message(
                chat_id,
                "❌ Ваш username не найден в системе.\n\nПопросите руководителя добавить ваш Telegram."
            )
    
    elif text == "/tasks":
        # Show user's tasks
        user = supabase.table("telegram_users").select("employee_id")\
            .eq("telegram_chat_id", chat_id).execute()
        
        if user.data:
            tasks = supabase.table("tasks").select("title, status, due_date")\
                .eq("assignee_id", user.data[0]["employee_id"])\
                .neq("status", "done")\
                .order("due_date").limit(10).execute()
            
            if tasks.data:
                msg = "📋 <b>Ваши задачи:</b>\n\n"
                for t in tasks.data:
                    status_emoji = {"backlog": "📝", "todo": "📌", "in_progress": "🔄", "review": "👀"}.get(t["status"], "")
                    due = f" (до {t['due_date']})" if t.get("due_date") else ""
                    msg += f"{status_emoji} {t['title']}{due}\n"
                await send_telegram_message(chat_id, msg)
            else:
                await send_telegram_message(chat_id, "✨ У вас нет активных задач!")
    
    return {"ok": True}


# ============ HELPER FUNCTIONS ============

async def check_epic_completion(epic_id: str):
    """Check if all subtasks are done and suggest closing epic"""
    if not supabase:
        return
    
    subtasks = supabase.table("tasks").select("status").eq("parent_id", epic_id).execute()
    
    if subtasks.data:
        all_done = all(t["status"] == "done" for t in subtasks.data)
        
        if all_done:
            # Mark epic as done
            supabase.table("tasks").update({
                "status": "done",
                "completed_at": datetime.now().isoformat()
            }).eq("id", epic_id).execute()

async def notify_new_task(task_id: str, assignee_id: str, title: str):
    """Send Telegram notification about new task"""
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


# ============ ANALYTICS ============

@app.get("/analytics/employee/{employee_id}")
async def get_employee_analytics(employee_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    meetings = supabase.table("meetings")\
        .select("date, mood_score, analysis")\
        .eq("employee_id", employee_id)\
        .order("date")\
        .execute()
    
    agreements = supabase.table("agreements")\
        .select("status, meetings!inner(employee_id)")\
        .eq("meetings.employee_id", employee_id)\
        .execute()
    
    mood_history = [
        {"date": m["date"], "score": m["mood_score"]}
        for m in meetings.data if m.get("mood_score")
    ]
    
    agreement_stats = {
        "total": len(agreements.data),
        "completed": len([a for a in agreements.data if a["status"] == "completed"]),
        "pending": len([a for a in agreements.data if a["status"] == "pending"]),
        "overdue": len([a for a in agreements.data if a["status"] == "overdue"])
    }
    
    red_flags_history = []
    for m in meetings.data:
        if m.get("analysis") and m["analysis"].get("red_flags"):
            flags = m["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") != "low":
                red_flags_history.append({
                    "date": m["date"],
                    "flags": flags
                })
    
    # Task stats
    tasks = supabase.table("tasks").select("status")\
        .or_(f"assignee_id.eq.{employee_id},co_assignee_id.eq.{employee_id}").execute()
    
    task_stats = {
        "total": len(tasks.data),
        "done": len([t for t in tasks.data if t["status"] == "done"]),
        "in_progress": len([t for t in tasks.data if t["status"] == "in_progress"]),
        "overdue": 0  # TODO: calculate based on due_date
    }
    
    return {
        "mood_history": mood_history,
        "agreement_stats": agreement_stats,
        "task_stats": task_stats,
        "red_flags_history": red_flags_history,
        "total_meetings": len(meetings.data)
    }

@app.get("/analytics/dashboard")
async def get_dashboard():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    employees = supabase.table("employees").select("*").execute()
    
    recent_meetings = supabase.table("meetings")\
        .select("*, employees(name)")\
        .order("date", desc=True)\
        .limit(10)\
        .execute()
    
    pending_agreements = supabase.table("agreements")\
        .select("*, meetings(date, employees(name))")\
        .eq("status", "pending")\
        .order("deadline")\
        .limit(10)\
        .execute()
    
    # Task summary
    tasks = supabase.table("tasks").select("status, due_date").execute()
    
    today = date.today().isoformat()
    task_summary = {
        "total": len(tasks.data),
        "done": len([t for t in tasks.data if t["status"] == "done"]),
        "in_progress": len([t for t in tasks.data if t["status"] == "in_progress"]),
        "overdue": len([t for t in tasks.data if t.get("due_date") and t["due_date"] < today and t["status"] != "done"])
    }
    
    red_flags = []
    for meeting in recent_meetings.data:
        if meeting.get("analysis") and meeting["analysis"].get("red_flags"):
            flags = meeting["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") in ["medium", "high"]:
                red_flags.append({
                    "employee": meeting["employees"]["name"],
                    "date": meeting["date"],
                    "flags": flags
                })
    
    return {
        "employees": employees.data,
        "recent_meetings": recent_meetings.data,
        "pending_agreements": pending_agreements.data,
        "task_summary": task_summary,
        "red_flags": red_flags
    }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
