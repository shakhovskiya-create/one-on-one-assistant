# ============ TASKS MODELS ============

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
    link_type: str  # blocks, blocked_by, relates_to

class DeadlineRequestCreate(BaseModel):
    task_id: str
    requester_id: str
    requested_date: str
    reason: str

class DeadlineRequestReview(BaseModel):
    status: str  # approved, rejected
    reviewer_id: str
    review_comment: Optional[str] = None

class TagCreate(BaseModel):
    name: str
    color: str = "gray"

class TaskCommentCreate(BaseModel):
    task_id: str
    author_id: str
    content: str
