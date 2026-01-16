'use client'

import { useState, useEffect, useCallback } from 'react'
import { Upload, FileAudio, Loader2, CheckCircle, AlertCircle, X, Users } from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
  position: string
}

interface Project {
  id: string
  name: string
  status: string
}

interface MeetingCategory {
  id: string
  code: string
  name: string
  description: string
}

interface TranscriptResult {
  whisper: string
  yandex: string
  merged: string
}

interface Analysis {
  meeting_id?: string
  summary: string
  // 1-on-1 specific
  employee_agenda?: string[]
  manager_agenda?: string[]
  development_notes?: string
  mood_score?: number
  mood_trend?: string
  comparison_with_previous?: string
  positive_signals?: string[]
  questions_to_ask?: string[]
  // Team meeting specific
  decisions?: string[]
  blockers?: string[]
  risks?: string[]
  open_questions?: string[]
  next_steps?: string[]
  project_health?: string
  health_reason?: string
  // Planning specific
  sprint_goal?: string
  committed_items?: Array<{task: string, responsible: string, estimate: string, priority: string}>
  capacity_concerns?: string[]
  dependencies?: string[]
  team_confidence?: number
  // Retro specific
  went_well?: string[]
  went_wrong?: string[]
  recurring_issues?: string[]
  resolved_issues?: string[]
  team_morale?: number
  morale_trend?: string
  patterns?: string[]
  // Interview specific
  candidate_strengths?: string[]
  candidate_weaknesses?: string[]
  technical_assessment?: {score: number, details: string}
  soft_skills_assessment?: {score: number, details: string}
  culture_fit?: {score: number, details: string}
  recommendation?: string
  recommendation_reason?: string
  suggested_next_steps?: string[]
  // Common
  agreements?: Array<{task: string, responsible: string, deadline: string | null}>
  action_items?: Array<{task: string, responsible: string, deadline: string | null}>
  key_points?: string[]
  red_flags?: {
    burnout_signs: boolean | string
    turnover_risk: string
    turnover_reason?: string
    team_conflicts: boolean | string
    concerns?: string[]
  }
  recommendations?: string[]
}

type ProcessingStatus = 'idle' | 'uploading' | 'transcribing' | 'analyzing' | 'done' | 'error'

export default function UploadPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [categories, setCategories] = useState<MeetingCategory[]>([])

  const [selectedCategory, setSelectedCategory] = useState('one_on_one')
  const [selectedEmployee, setSelectedEmployee] = useState('')
  const [selectedProject, setSelectedProject] = useState('')
  const [selectedParticipants, setSelectedParticipants] = useState<string[]>([])
  const [meetingTitle, setMeetingTitle] = useState('')
  const [meetingDate, setMeetingDate] = useState(new Date().toISOString().split('T')[0])

  const [file, setFile] = useState<File | null>(null)
  const [status, setStatus] = useState<ProcessingStatus>('idle')
  const [progress, setProgress] = useState(0)
  const [transcripts, setTranscripts] = useState<TranscriptResult | null>(null)
  const [analysis, setAnalysis] = useState<Analysis | null>(null)
  const [error, setError] = useState('')
  const [dragActive, setDragActive] = useState(false)
  const [showTranscriptDetails, setShowTranscriptDetails] = useState(false)

  useEffect(() => {
    fetchEmployees()
    fetchProjects()
    fetchCategories()
  }, [])

  const fetchEmployees = async () => {
    try {
      const res = await fetch(`${API_URL}/employees`)
      if (res.ok) {
        const data = await res.json()
        setEmployees(data)
      }
    } catch (error) {
      console.error('Failed to fetch employees:', error)
    }
  }

  const fetchProjects = async () => {
    try {
      const res = await fetch(`${API_URL}/projects?status=active`)
      if (res.ok) {
        const data = await res.json()
        setProjects(data)
      }
    } catch (error) {
      console.error('Failed to fetch projects:', error)
    }
  }

  const fetchCategories = async () => {
    try {
      const res = await fetch(`${API_URL}/meeting-categories`)
      if (res.ok) {
        const data = await res.json()
        setCategories(data)
      }
    } catch (error) {
      console.error('Failed to fetch categories:', error)
    }
  }

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      setFile(e.dataTransfer.files[0])
    }
  }, [])

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0])
    }
  }

  const toggleParticipant = (id: string) => {
    setSelectedParticipants(prev =>
      prev.includes(id)
        ? prev.filter(p => p !== id)
        : [...prev, id]
    )
  }

  const isOneOnOne = selectedCategory === 'one_on_one'
  const needsProject = ['team_meeting', 'planning', 'retro', 'kickoff', 'status', 'demo'].includes(selectedCategory)
  const needsParticipants = ['team_meeting', 'planning', 'retro', 'kickoff', 'status', 'demo'].includes(selectedCategory)

  const canSubmit = () => {
    if (!file) return false
    if (isOneOnOne && !selectedEmployee) return false
    if (needsProject && !selectedProject) return false
    return true
  }

  const processFile = async () => {
    if (!canSubmit()) {
      setError('Заполните обязательные поля')
      return
    }

    setError('')
    setStatus('uploading')
    setProgress(10)

    try {
      const formData = new FormData()
      formData.append('file', file!)
      formData.append('category_code', selectedCategory)
      formData.append('meeting_date', meetingDate)

      if (meetingTitle) {
        formData.append('title', meetingTitle)
      }

      if (isOneOnOne && selectedEmployee) {
        formData.append('employee_id', selectedEmployee)
      }

      if (needsProject && selectedProject) {
        formData.append('project_id', selectedProject)
      }

      if (needsParticipants && selectedParticipants.length > 0) {
        formData.append('participant_ids', JSON.stringify(selectedParticipants))
      }

      setStatus('transcribing')
      setProgress(30)

      const res = await fetch(`${API_URL}/process-meeting`, {
        method: 'POST',
        body: formData,
      })

      setProgress(70)

      if (!res.ok) {
        const errorData = await res.json().catch(() => ({}))
        throw new Error(errorData.detail || 'Processing failed')
      }

      setStatus('analyzing')
      setProgress(90)

      const data = await res.json()
      setTranscripts(data.transcript)
      setAnalysis(data.analysis)

      setStatus('done')
      setProgress(100)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка')
      setStatus('error')
    }
  }

  const reset = () => {
    setFile(null)
    setStatus('idle')
    setProgress(0)
    setTranscripts(null)
    setAnalysis(null)
    setError('')
    setShowTranscriptDetails(false)
  }

  const getStatusText = () => {
    switch (status) {
      case 'uploading':
        return 'Загрузка файла...'
      case 'transcribing':
        return 'Транскрибирование (Whisper + Yandex)...'
      case 'analyzing':
        return 'AI-анализ встречи...'
      case 'done':
        return 'Готово!'
      case 'error':
        return 'Ошибка'
      default:
        return ''
    }
  }

  const getCategoryName = (code: string) => {
    return categories.find(c => c.code === code)?.name || code
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Загрузить запись встречи</h1>

      {status === 'idle' && (
        <>
          {/* Category Selection */}
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              Тип встречи
            </label>
            <div className="grid grid-cols-4 gap-2">
              {categories.map((cat) => (
                <button
                  key={cat.code}
                  onClick={() => {
                    setSelectedCategory(cat.code)
                    setSelectedEmployee('')
                    setSelectedProject('')
                    setSelectedParticipants([])
                  }}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    selectedCategory === cat.code
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {cat.name}
                </button>
              ))}
            </div>
            {categories.find(c => c.code === selectedCategory)?.description && (
              <p className="text-sm text-gray-500 mt-2">
                {categories.find(c => c.code === selectedCategory)?.description}
              </p>
            )}
          </div>

          {/* Dynamic Settings based on category */}
          <div className="bg-white rounded-lg shadow-sm border p-6 space-y-4">
            <div className="grid grid-cols-2 gap-4">
              {/* Title - always shown */}
              <div className="col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Название встречи (опционально)
                </label>
                <input
                  type="text"
                  value={meetingTitle}
                  onChange={(e) => setMeetingTitle(e.target.value)}
                  placeholder={`${getCategoryName(selectedCategory)} - ${meetingDate}`}
                  className="w-full border rounded-lg px-3 py-2"
                />
              </div>

              {/* Employee - for 1-on-1 */}
              {isOneOnOne && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Сотрудник *
                  </label>
                  <select
                    value={selectedEmployee}
                    onChange={(e) => setSelectedEmployee(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2"
                  >
                    <option value="">Выберите сотрудника</option>
                    {employees.map((emp) => (
                      <option key={emp.id} value={emp.id}>
                        {emp.name} - {emp.position}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Project - for team meetings */}
              {needsProject && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Проект *
                  </label>
                  <select
                    value={selectedProject}
                    onChange={(e) => setSelectedProject(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2"
                  >
                    <option value="">Выберите проект</option>
                    {projects.map((proj) => (
                      <option key={proj.id} value={proj.id}>
                        {proj.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Date */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Дата встречи
                </label>
                <input
                  type="date"
                  value={meetingDate}
                  onChange={(e) => setMeetingDate(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2"
                />
              </div>
            </div>

            {/* Participants - for team meetings */}
            {needsParticipants && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Участники
                </label>
                <div className="flex flex-wrap gap-2">
                  {employees.map((emp) => (
                    <button
                      key={emp.id}
                      onClick={() => toggleParticipant(emp.id)}
                      className={`flex items-center gap-1 px-3 py-1 rounded-full text-sm transition-colors ${
                        selectedParticipants.includes(emp.id)
                          ? 'bg-blue-100 text-blue-800 border border-blue-300'
                          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                      }`}
                    >
                      <Users size={14} />
                      {emp.name}
                    </button>
                  ))}
                </div>
                {selectedParticipants.length > 0 && (
                  <p className="text-sm text-gray-500 mt-2">
                    Выбрано: {selectedParticipants.length}
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Drop Zone */}
          <div
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
            className={`bg-white rounded-lg shadow-sm border-2 border-dashed p-12 text-center transition-colors ${
              dragActive
                ? 'border-blue-500 bg-blue-50'
                : file
                ? 'border-green-500 bg-green-50'
                : 'border-gray-300'
            }`}
          >
            {file ? (
              <div className="space-y-4">
                <FileAudio size={48} className="mx-auto text-green-600" />
                <div>
                  <p className="font-medium text-gray-900">{file.name}</p>
                  <p className="text-sm text-gray-500">
                    {(file.size / 1024 / 1024).toFixed(2)} MB
                  </p>
                </div>
                <button
                  onClick={() => setFile(null)}
                  className="text-sm text-red-600 hover:underline"
                >
                  Удалить
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <Upload size={48} className="mx-auto text-gray-400" />
                <div>
                  <p className="font-medium text-gray-900">
                    Перетащите файл сюда или{' '}
                    <label className="text-blue-600 hover:underline cursor-pointer">
                      выберите
                      <input
                        type="file"
                        accept="audio/*,video/*"
                        onChange={handleFileChange}
                        className="hidden"
                      />
                    </label>
                  </p>
                  <p className="text-sm text-gray-500 mt-1">
                    Поддерживаются: MP3, WAV, MP4, WebM, M4A, OGG
                  </p>
                </div>
              </div>
            )}
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-center gap-3">
              <AlertCircle className="text-red-600" size={20} />
              <span className="text-red-700">{error}</span>
            </div>
          )}

          <button
            onClick={processFile}
            disabled={!canSubmit()}
            className="w-full py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium"
          >
            Обработать запись
          </button>
        </>
      )}

      {(status === 'uploading' || status === 'transcribing' || status === 'analyzing') && (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
          <Loader2 size={48} className="mx-auto text-blue-600 animate-spin" />
          <div>
            <p className="font-medium text-gray-900">{getStatusText()}</p>
            <p className="text-sm text-gray-500 mt-1">Это может занять несколько минут</p>
          </div>
          <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-600 transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {status === 'done' && analysis && (
        <div className="space-y-6">
          <div className="bg-green-50 border border-green-200 rounded-lg p-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <CheckCircle className="text-green-600" size={24} />
              <span className="text-green-700 font-medium">Анализ завершен!</span>
            </div>
            <button onClick={reset} className="text-gray-600 hover:text-gray-800">
              <X size={20} />
            </button>
          </div>

          {/* Summary */}
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-lg font-semibold mb-4">Резюме встречи</h2>
            <p className="text-gray-700">{analysis.summary}</p>

            {/* Mood/Health indicators */}
            <div className="mt-4 flex items-center gap-4 flex-wrap">
              {analysis.mood_score && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">Настроение:</span>
                  <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      analysis.mood_score >= 7
                        ? 'bg-green-100 text-green-800'
                        : analysis.mood_score >= 5
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {analysis.mood_score}/10
                    {analysis.mood_trend && (
                      <span className="ml-1">
                        ({analysis.mood_trend === 'improving' ? '↑' : analysis.mood_trend === 'declining' ? '↓' : '→'})
                      </span>
                    )}
                  </span>
                </div>
              )}

              {analysis.project_health && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">Здоровье проекта:</span>
                  <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      analysis.project_health === 'green'
                        ? 'bg-green-100 text-green-800'
                        : analysis.project_health === 'yellow'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {analysis.project_health === 'green' ? 'Хорошо' :
                     analysis.project_health === 'yellow' ? 'Есть риски' : 'Проблемы'}
                  </span>
                </div>
              )}

              {analysis.team_confidence && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">Уверенность команды:</span>
                  <span className="px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
                    {analysis.team_confidence}/10
                  </span>
                </div>
              )}

              {analysis.team_morale && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">Моральный дух:</span>
                  <span className="px-3 py-1 rounded-full text-sm font-medium bg-purple-100 text-purple-800">
                    {analysis.team_morale}/10
                    {analysis.morale_trend && (
                      <span className="ml-1">
                        ({analysis.morale_trend === 'improving' ? '↑' : analysis.morale_trend === 'declining' ? '↓' : '→'})
                      </span>
                    )}
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* 1-on-1 specific: Agendas */}
          {(analysis.employee_agenda || analysis.manager_agenda) && (
            <div className="grid grid-cols-2 gap-6">
              {analysis.employee_agenda && (
                <div className="bg-white rounded-lg shadow-sm border p-6">
                  <h3 className="font-semibold mb-3">Повестка сотрудника</h3>
                  <ul className="space-y-2">
                    {analysis.employee_agenda.map((item, i) => (
                      <li key={i} className="text-gray-700 flex items-start gap-2">
                        <span className="text-blue-600">•</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {analysis.manager_agenda && (
                <div className="bg-white rounded-lg shadow-sm border p-6">
                  <h3 className="font-semibold mb-3">Повестка руководителя</h3>
                  <ul className="space-y-2">
                    {analysis.manager_agenda.map((item, i) => (
                      <li key={i} className="text-gray-700 flex items-start gap-2">
                        <span className="text-blue-600">•</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Retro specific: What went well/wrong */}
          {(analysis.went_well || analysis.went_wrong) && (
            <div className="grid grid-cols-2 gap-6">
              {analysis.went_well && (
                <div className="bg-green-50 rounded-lg border border-green-200 p-6">
                  <h3 className="font-semibold mb-3 text-green-800">Что было хорошо</h3>
                  <ul className="space-y-2">
                    {analysis.went_well.map((item, i) => (
                      <li key={i} className="text-green-700 flex items-start gap-2">
                        <span>✓</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {analysis.went_wrong && (
                <div className="bg-red-50 rounded-lg border border-red-200 p-6">
                  <h3 className="font-semibold mb-3 text-red-800">Что пошло не так</h3>
                  <ul className="space-y-2">
                    {analysis.went_wrong.map((item, i) => (
                      <li key={i} className="text-red-700 flex items-start gap-2">
                        <span>✗</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Decisions */}
          {analysis.decisions && analysis.decisions.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-3">Принятые решения</h3>
              <ul className="space-y-2">
                {analysis.decisions.map((item, i) => (
                  <li key={i} className="text-gray-700 flex items-start gap-2">
                    <span className="text-green-600">✓</span>
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Sprint goal */}
          {analysis.sprint_goal && (
            <div className="bg-blue-50 rounded-lg border border-blue-200 p-6">
              <h3 className="font-semibold mb-2 text-blue-800">Цель спринта</h3>
              <p className="text-blue-700">{analysis.sprint_goal}</p>
            </div>
          )}

          {/* Agreements / Action Items */}
          {((analysis.agreements && analysis.agreements.length > 0) ||
            (analysis.action_items && analysis.action_items.length > 0)) && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">
                {analysis.agreements ? 'Договоренности' : 'Задачи'}
              </h3>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-3">Задача</th>
                      <th className="text-left py-2 px-3">Ответственный</th>
                      <th className="text-left py-2 px-3">Срок</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(analysis.agreements || analysis.action_items || []).map((item, i) => (
                      <tr key={i} className="border-b">
                        <td className="py-2 px-3">{item.task}</td>
                        <td className="py-2 px-3">{item.responsible}</td>
                        <td className="py-2 px-3">{item.deadline || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Blockers & Risks */}
          {(analysis.blockers || analysis.risks) && (
            <div className="grid grid-cols-2 gap-6">
              {analysis.blockers && analysis.blockers.length > 0 && (
                <div className="bg-red-50 rounded-lg border border-red-200 p-6">
                  <h3 className="font-semibold mb-3 text-red-800">Блокеры</h3>
                  <ul className="space-y-2">
                    {analysis.blockers.map((item, i) => (
                      <li key={i} className="text-red-700 flex items-start gap-2">
                        <span>⚠</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {analysis.risks && analysis.risks.length > 0 && (
                <div className="bg-yellow-50 rounded-lg border border-yellow-200 p-6">
                  <h3 className="font-semibold mb-3 text-yellow-800">Риски</h3>
                  <ul className="space-y-2">
                    {analysis.risks.map((item, i) => (
                      <li key={i} className="text-yellow-700 flex items-start gap-2">
                        <span>⚡</span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Development Notes - 1on1 */}
          {analysis.development_notes && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-3">Развитие сотрудника</h3>
              <p className="text-gray-700">{analysis.development_notes}</p>
            </div>
          )}

          {/* Comparison with previous */}
          {analysis.comparison_with_previous && (
            <div className="bg-blue-50 rounded-lg border border-blue-200 p-6">
              <h3 className="font-semibold mb-2 text-blue-800">Сравнение с прошлой встречей</h3>
              <p className="text-blue-700">{analysis.comparison_with_previous}</p>
            </div>
          )}

          {/* Red Flags - 1on1 */}
          {analysis.red_flags && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Красные флаги</h3>
              <div className="grid grid-cols-3 gap-4">
                <div
                  className={`p-4 rounded-lg ${
                    analysis.red_flags.burnout_signs
                      ? 'bg-red-50 border border-red-200'
                      : 'bg-green-50 border border-green-200'
                  }`}
                >
                  <p className="font-medium text-sm">Признаки выгорания</p>
                  <p
                    className={`text-sm mt-1 ${
                      analysis.red_flags.burnout_signs ? 'text-red-700' : 'text-green-700'
                    }`}
                  >
                    {analysis.red_flags.burnout_signs
                      ? typeof analysis.red_flags.burnout_signs === 'string'
                        ? analysis.red_flags.burnout_signs
                        : 'Да'
                      : 'Нет'}
                  </p>
                </div>
                <div
                  className={`p-4 rounded-lg ${
                    analysis.red_flags.turnover_risk === 'high'
                      ? 'bg-red-50 border border-red-200'
                      : analysis.red_flags.turnover_risk === 'medium'
                      ? 'bg-yellow-50 border border-yellow-200'
                      : 'bg-green-50 border border-green-200'
                  }`}
                >
                  <p className="font-medium text-sm">Риск ухода</p>
                  <p className="text-sm mt-1">
                    {analysis.red_flags.turnover_risk === 'high'
                      ? 'Высокий'
                      : analysis.red_flags.turnover_risk === 'medium'
                      ? 'Средний'
                      : 'Низкий'}
                  </p>
                  {analysis.red_flags.turnover_reason && (
                    <p className="text-xs mt-1 text-gray-500">{analysis.red_flags.turnover_reason}</p>
                  )}
                </div>
                <div
                  className={`p-4 rounded-lg ${
                    analysis.red_flags.team_conflicts
                      ? 'bg-red-50 border border-red-200'
                      : 'bg-green-50 border border-green-200'
                  }`}
                >
                  <p className="font-medium text-sm">Конфликты в команде</p>
                  <p
                    className={`text-sm mt-1 ${
                      analysis.red_flags.team_conflicts ? 'text-red-700' : 'text-green-700'
                    }`}
                  >
                    {analysis.red_flags.team_conflicts
                      ? typeof analysis.red_flags.team_conflicts === 'string'
                        ? analysis.red_flags.team_conflicts
                        : 'Да'
                      : 'Нет'}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Interview: Candidate Assessment */}
          {analysis.recommendation && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Оценка кандидата</h3>

              <div className="grid grid-cols-3 gap-4 mb-4">
                {analysis.technical_assessment && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <p className="font-medium text-sm">Технические навыки</p>
                    <p className="text-2xl font-bold text-blue-600">{analysis.technical_assessment.score}/10</p>
                    <p className="text-xs text-gray-500 mt-1">{analysis.technical_assessment.details}</p>
                  </div>
                )}
                {analysis.soft_skills_assessment && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <p className="font-medium text-sm">Soft Skills</p>
                    <p className="text-2xl font-bold text-purple-600">{analysis.soft_skills_assessment.score}/10</p>
                    <p className="text-xs text-gray-500 mt-1">{analysis.soft_skills_assessment.details}</p>
                  </div>
                )}
                {analysis.culture_fit && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <p className="font-medium text-sm">Culture Fit</p>
                    <p className="text-2xl font-bold text-green-600">{analysis.culture_fit.score}/10</p>
                    <p className="text-xs text-gray-500 mt-1">{analysis.culture_fit.details}</p>
                  </div>
                )}
              </div>

              <div className={`p-4 rounded-lg ${
                analysis.recommendation === 'hire' ? 'bg-green-100' :
                analysis.recommendation === 'no_hire' ? 'bg-red-100' : 'bg-yellow-100'
              }`}>
                <p className="font-semibold">
                  Рекомендация: {
                    analysis.recommendation === 'hire' ? 'Нанять' :
                    analysis.recommendation === 'no_hire' ? 'Отказать' : 'Требуется обсуждение'
                  }
                </p>
                <p className="text-sm mt-1">{analysis.recommendation_reason}</p>
              </div>
            </div>
          )}

          {/* Recommendations */}
          {analysis.recommendations && analysis.recommendations.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-3">Рекомендации</h3>
              <ul className="space-y-2">
                {analysis.recommendations.map((rec, i) => (
                  <li key={i} className="text-gray-700 flex items-start gap-2">
                    <span className="text-blue-600">•</span>
                    {rec}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Questions for next meeting */}
          {analysis.questions_to_ask && analysis.questions_to_ask.length > 0 && (
            <div className="bg-purple-50 rounded-lg border border-purple-200 p-6">
              <h3 className="font-semibold mb-3 text-purple-800">Вопросы на следующую встречу</h3>
              <ul className="space-y-2">
                {analysis.questions_to_ask.map((q, i) => (
                  <li key={i} className="text-purple-700 flex items-start gap-2">
                    <span>?</span>
                    {q}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Transcript */}
          <details className="bg-white rounded-lg shadow-sm border">
            <summary className="p-6 cursor-pointer font-semibold flex items-center justify-between">
              <span>Транскрипт встречи</span>
              {transcripts && (
                <button
                  onClick={(e) => {
                    e.preventDefault()
                    setShowTranscriptDetails(!showTranscriptDetails)
                  }}
                  className="text-sm text-blue-600 hover:underline"
                >
                  {showTranscriptDetails ? 'Скрыть детали' : 'Показать Whisper/Yandex'}
                </button>
              )}
            </summary>
            <div className="px-6 pb-6 space-y-4">
              {showTranscriptDetails && transcripts && (
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h4 className="font-medium text-sm text-gray-600 mb-2">Whisper (OpenAI)</h4>
                    <pre className="whitespace-pre-wrap text-xs text-gray-600 bg-blue-50 p-3 rounded-lg max-h-48 overflow-y-auto">
                      {transcripts.whisper || 'Нет данных'}
                    </pre>
                  </div>
                  <div>
                    <h4 className="font-medium text-sm text-gray-600 mb-2">Yandex SpeechKit</h4>
                    <pre className="whitespace-pre-wrap text-xs text-gray-600 bg-green-50 p-3 rounded-lg max-h-48 overflow-y-auto">
                      {transcripts.yandex || 'Нет данных'}
                    </pre>
                  </div>
                </div>
              )}
              <div>
                <h4 className="font-medium text-sm text-gray-600 mb-2">
                  {showTranscriptDetails ? 'Объединённый транскрипт (Claude)' : ''}
                </h4>
                <pre className="whitespace-pre-wrap text-sm text-gray-700 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto">
                  {transcripts?.merged || ''}
                </pre>
              </div>
            </div>
          </details>

          <button
            onClick={reset}
            className="w-full py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 font-medium"
          >
            Загрузить еще одну запись
          </button>
        </div>
      )}

      {status === 'error' && (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
          <AlertCircle size={48} className="mx-auto text-red-600" />
          <div>
            <p className="font-medium text-gray-900">Произошла ошибка</p>
            <p className="text-sm text-gray-500 mt-1">{error}</p>
          </div>
          <button
            onClick={reset}
            className="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
          >
            Попробовать снова
          </button>
        </div>
      )}
    </div>
  )
}
