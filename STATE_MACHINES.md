# STATE_MACHINES.md
# EKF Hub — Базовые state machines

Этот файл фиксирует допустимые состояния и переходы.
Код обязан соответствовать этому файлу. Если в коде иначе — обновить этот файл (через diff).

---

## Task

Состояния (baseline):
- Backlog
- ToDo
- InProgress
- Review
- Done

Разрешённые переходы:
- Backlog → ToDo
- ToDo → InProgress
- InProgress → Review
- Review → Done
- Review → InProgress

Правила:
- Невалидные переходы отклоняются
- Побочные эффекты переходов (events/audit/log) должны быть явными

---

## Meeting

Состояния (baseline):
- Planned
- Conducted
- Processed

Переходы:
- Planned → Conducted
- Conducted → Processed

Правила:
- Нельзя Processed без Conducted
- Post-processing может создавать Agreements/Tasks (явно, не магией)

---

## Agreement

Состояния (baseline):
- Open
- Completed
- Overdue

Правила:
- Overdue определяется явным правилом (deadline + job), без скрытой "магии"
