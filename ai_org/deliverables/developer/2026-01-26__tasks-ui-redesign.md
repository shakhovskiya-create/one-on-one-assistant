# Tasks UI Redesign

**Дата:** 2026-01-26
**Статус:** ✅ DONE

## Summary

Редизайн Kanban карточек задач для соответствия макету 01-tasks.html.

## Changes

### File: `frontend/src/routes/tasks/+page.svelte`

#### 1. Task Interface Extension
```typescript
interface Task {
  // ... existing fields
  task_type?: 'feature' | 'bug' | 'tech_debt' | 'improvement' | 'task';
}
```

#### 2. New Constants
```typescript
const taskTypeLabels = {
  feature: { label: 'Фича', color: 'bg-blue-100 text-blue-700', icon: '✨' },
  bug: { label: 'Баг', color: 'bg-red-100 text-red-700', icon: '🐛' },
  tech_debt: { label: 'Техдолг', color: 'bg-purple-100 text-purple-700', icon: '🔧' },
  improvement: { label: 'Улучшение', color: 'bg-green-100 text-green-700', icon: '📈' },
  task: { label: 'Задача', color: 'bg-gray-100 text-gray-700', icon: '📋' }
};

const priorityBorderColors = {
  1: 'border-l-red-500',
  2: 'border-l-orange-500',
  3: 'border-l-yellow-400',
  4: 'border-l-blue-300',
  5: 'border-l-gray-300'
};
```

#### 3. Helper Function
```typescript
function getEmployeePhoto(id: string): string | null {
  const emp = employees.find(e => e.id === id);
  return emp?.photo_base64 || null;
}
```

#### 4. Kanban Card Redesign
- `rounded-lg` - закруглённые углы
- `hover:-translate-y-0.5 hover:shadow-lg` - hover эффекты
- `border-l-[3px]` + `priorityBorderColors` - цветная граница
- Task ID format: `EKF-{id.substring(0,4).toUpperCase()}`
- Employee photo with `ring-2 ring-white shadow-sm`

#### 5. Modal Form Update
- Grid changed from 4 to 5 columns
- Added task_type selector with emoji icons

## Design System

- EKF Red: #E53935
- Rounded corners: `rounded-lg`
- Card shadow: `shadow-sm` → `hover:shadow-lg`
- Priority colors: red/orange/yellow/blue/gray
- Type badges: colored pills with icons

## Testing

- [x] Cards display with rounded corners
- [x] Type badges show correctly
- [x] Photos render from base64
- [x] Hover effects work
- [x] Modal saves task_type
