import os
import json
import tempfile
from datetime import datetime, date
from typing import Optional, List
from fastapi import FastAPI, UploadFile, File, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from openai import OpenAI
from anthropic import Anthropic
from supabase import create_client, Client
import uvicorn

app = FastAPI(title="1-on-1 Assistant API", version="1.0.0")

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

# Models
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

# Script data
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


# API Endpoints

@app.get("/")
async def root():
    return {"status": "ok", "service": "1-on-1 Assistant API"}

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "openai": openai_client is not None,
        "anthropic": anthropic_client is not None,
        "supabase": supabase is not None
    }

@app.get("/script")
async def get_script():
    return MEETING_SCRIPT

# Employees
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

# Meetings
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

# Transcription
@app.post("/transcribe")
async def transcribe_audio(file: UploadFile = File(...)):
    if not openai_client:
        raise HTTPException(status_code=500, detail="OpenAI not configured")
    
    # Save uploaded file temporarily
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

# Analysis
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
    
    # Parse JSON from response
    try:
        # Try to extract JSON if wrapped in markdown
        if "```json" in response_text:
            response_text = response_text.split("```json")[1].split("```")[0]
        elif "```" in response_text:
            response_text = response_text.split("```")[1].split("```")[0]
        
        analysis = json.loads(response_text.strip())
        return analysis
    except json.JSONDecodeError as e:
        raise HTTPException(status_code=500, detail=f"Failed to parse analysis: {str(e)}")

# Full processing pipeline
@app.post("/process-meeting")
async def process_meeting(
    employee_id: str,
    meeting_date: str,
    file: UploadFile = File(...)
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
        
        # Save agreements
        for agreement in analysis.get("agreements", []):
            agreement_data = {
                "meeting_id": meeting_id,
                "task": agreement["task"],
                "responsible": agreement["responsible"],
                "deadline": agreement.get("deadline"),
                "status": "pending"
            }
            supabase.table("agreements").insert(agreement_data).execute()
        
        analysis["meeting_id"] = meeting_id
    
    return {
        "transcript": transcript,
        "analysis": analysis
    }

# Agreements
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
    
    # Filter by employee if needed
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

# Analytics
@app.get("/analytics/employee/{employee_id}")
async def get_employee_analytics(employee_id: str):
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get all meetings
    meetings = supabase.table("meetings")\
        .select("date, mood_score, analysis")\
        .eq("employee_id", employee_id)\
        .order("date")\
        .execute()
    
    # Get agreements stats
    agreements = supabase.table("agreements")\
        .select("status, meetings!inner(employee_id)")\
        .eq("meetings.employee_id", employee_id)\
        .execute()
    
    # Calculate stats
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
    
    # Extract red flags history
    red_flags_history = []
    for m in meetings.data:
        if m.get("analysis") and m["analysis"].get("red_flags"):
            flags = m["analysis"]["red_flags"]
            if flags.get("burnout_signs") or flags.get("turnover_risk") != "low":
                red_flags_history.append({
                    "date": m["date"],
                    "flags": flags
                })
    
    return {
        "mood_history": mood_history,
        "agreement_stats": agreement_stats,
        "red_flags_history": red_flags_history,
        "total_meetings": len(meetings.data)
    }

@app.get("/analytics/dashboard")
async def get_dashboard():
    if not supabase:
        raise HTTPException(status_code=500, detail="Database not configured")
    
    # Get all employees with latest meeting
    employees = supabase.table("employees").select("*").execute()
    
    # Get recent meetings
    recent_meetings = supabase.table("meetings")\
        .select("*, employees(name)")\
        .order("date", desc=True)\
        .limit(10)\
        .execute()
    
    # Get pending agreements
    pending_agreements = supabase.table("agreements")\
        .select("*, meetings(date, employees(name))")\
        .eq("status", "pending")\
        .order("deadline")\
        .limit(10)\
        .execute()
    
    # Find red flags
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
        "red_flags": red_flags
    }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
