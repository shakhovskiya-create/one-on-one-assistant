#!/usr/bin/env bash
set -e

ROOT="$(pwd)"

# ---------- paths ----------
AI_ORG="$ROOT/ai_org"
HANDOFF_DIR="$AI_ORG/handoffs/active"
DELIV_DIR="$AI_ORG/deliverables/designer"
LOG_DIR="$AI_ORG/logs"

HANDOFF_FILE="$HANDOFF_DIR/$(date +%Y-%m-%d)__DESIGN__FIGMA_AI_PROCESS.md"
SPEC_FILE="$DELIV_DIR/figma_prompt_spec.md"
GAPS_FILE="$DELIV_DIR/FIGMA_GAPS.md"
AGENT_LOG="$LOG_DIR/AGENT_LOG.md"

# ---------- ensure dirs ----------
mkdir -p "$HANDOFF_DIR" "$DELIV_DIR" "$LOG_DIR"

# ---------- backup helper ----------
backup_if_exists () {
  if [ -f "$1" ]; then
    cp "$1" "$1.bak.$(date +%Y%m%d_%H%M%S)"
  fi
}

# ---------- handoff ----------
cat > "$HANDOFF_FILE" <<'EOF'
# HANDOFF — Процесс дизайна через Figma AI

## Контекст
В проекте вводится формализованный процесс работы с Figma AI
через подготовку канонического текстового задания и обязательную валидацию результата.

## Цель
- Исключить прямую работу с Figma AI без согласованного задания
- Сделать дизайн воспроизводимым и проверяемым
- Разделить ответственность между ролями

## Процесс
Designer → figma_prompt_spec.md → Product/Owner approval →
Figma Chat (Product/Owner) → Figma AI →
Designer (via MCP) → FIGMA_GAPS.md

## Инварианты
- figma_prompt_spec.md — единственный источник задания
- Дизайнер НЕ отправляет промпты в Figma Chat
- Все расхождения фиксируются в FIGMA_GAPS.md
EOF

# ---------- figma_prompt_spec ----------
backup_if_exists "$SPEC_FILE"
cat > "$SPEC_FILE" <<'EOF'
# FIGMA PROMPT SPEC

⚠️ ВАЖНО  
Этот файл предназначен для КОПИРОВАНИЯ В FIGMA CHAT.  
Любые изменения макетов без обновления этого файла считаются дефектом процесса.

## Цель
Подготовить однозначное и проверяемое задание для генерации UI в Figma AI.

## Контекст
- Продукт: IKF Hub
- Платформа: Web (desktop-first)
- Язык интерфейса: русский
- Стиль: корпоративный, нейтральный

## Навигация (фиксировано)
Top Bar:
Главная / Сотрудники / Задачи / Встречи / Почта / Сообщения / SD / Аналитика

Left Sidebar:
Только контекстный, без дублирования Top Bar.

## Экраны
(заполняется дизайнером)

## Компоненты
(заполняется дизайнером)

## Состояния
Empty / Loading / Error / Read-only

## Ограничения
- Не добавлять новые пункты глобальной навигации
- Не менять бизнес-логику
- Не изобретать новые UX-паттерны

## PROMPT (для Figma Chat)
<ВСТАВИТЬ ФИНАЛЬНЫЙ ПРОМПТ СЮДА>
EOF

# ---------- FIGMA_GAPS ----------
backup_if_exists "$GAPS_FILE"
cat > "$GAPS_FILE" <<'EOF'
# FIGMA GAPS — Расхождения Spec vs Result

Назначение:
Фиксация всех отклонений между figma_prompt_spec.md и результатом Figma AI.

## Цикл

### Метаданные
- Дата:
- Ссылка на Figma:
- Версия spec:

### GAPS
- [ ] GAP-001 | Severity: BLOCKER/MAJOR/MINOR
  - Требование из spec:
  - Фактический результат:
  - Как исправлять:

### Решения
(если принимались)

### Следующие шаги
- Designer:
- Product/Owner:
EOF

# ---------- AGENT LOG ----------
touch "$AGENT_LOG"
cat >> "$AGENT_LOG" <<EOF

---
timestamp: $(date -Iseconds)
cycle_id: design-figma-process-bootstrap
read:
  - designer.md
  - CLAUDE.md
created_or_updated:
  - $HANDOFF_FILE
  - $SPEC_FILE
  - $GAPS_FILE
notes:
  - Инициализирован формализованный процесс дизайна через Figma AI
---
EOF

echo "OK: handoff, designer deliverables и AGENT_LOG обновлены."