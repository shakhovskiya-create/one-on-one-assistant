'use client'

import { useState, useEffect, useCallback } from 'react'
import { Upload, FileAudio, Loader2, CheckCircle, AlertCircle, X } from 'lucide-react'

const API_URL = process.env.API_URL || 'http://localhost:8000'

interface Employee {
  id: string
  name: string
  position: string
}

interface Analysis {
  meeting_id?: string
  summary: string
  employee_agenda: string[]
  manager_agenda: string[]
  agreements: Array<{
    task: string
    responsible: string
    deadline: string | null
  }>
  development_notes: string
  red_flags: {
    burnout_signs: boolean | string
    turnover_risk: string
    team_conflicts: boolean | string
  }
  mood_score: number
  mood_change?: number
  recommendations: string[]
}

type ProcessingStatus = 'idle' | 'uploading' | 'transcribing' | 'analyzing' | 'done' | 'error'

export default function UploadPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedEmployee, setSelectedEmployee] = useState('')
  const [meetingDate, setMeetingDate] = useState(new Date().toISOString().split('T')[0])
  const [file, setFile] = useState<File | null>(null)
  const [status, setStatus] = useState<ProcessingStatus>('idle')
  const [progress, setProgress] = useState(0)
  const [transcript, setTranscript] = useState('')
  const [analysis, setAnalysis] = useState<Analysis | null>(null)
  const [error, setError] = useState('')
  const [dragActive, setDragActive] = useState(false)

  useEffect(() => {
    fetchEmployees()
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

  const processFile = async () => {
    if (!file || !selectedEmployee) {
      setError('Выберите сотрудника и файл')
      return
    }

    setError('')
    setStatus('uploading')
    setProgress(10)

    try {
      const formData = new FormData()
      formData.append('file', file)

      setStatus('transcribing')
      setProgress(30)

      // Full pipeline
      const res = await fetch(
        `${API_URL}/process-meeting?employee_id=${selectedEmployee}&meeting_date=${meetingDate}`,
        {
          method: 'POST',
          body: formData,
        }
      )

      setProgress(70)

      if (!res.ok) {
        throw new Error('Processing failed')
      }

      setStatus('analyzing')
      setProgress(90)

      const data = await res.json()
      setTranscript(data.transcript)
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
    setTranscript('')
    setAnalysis(null)
    setError('')
  }

  const getStatusText = () => {
    switch (status) {
      case 'uploading':
        return 'Загрузка файла...'
      case 'transcribing':
        return 'Транскрибирование аудио...'
      case 'analyzing':
        return 'Анализ встречи...'
      case 'done':
        return 'Готово!'
      case 'error':
        return 'Ошибка'
      default:
        return ''
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Загрузить запись встречи</h1>

      {status === 'idle' && (
        <>
          {/* Settings */}
          <div className="bg-white rounded-lg shadow-sm border p-6 space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Сотрудник
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
                    Поддерживаются: MP3, WAV, MP4, WebM, M4A
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
            disabled={!file || !selectedEmployee}
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
            <div className="mt-4 flex items-center gap-4">
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
                  {analysis.mood_change !== undefined && (
                    <span className="ml-1">
                      ({analysis.mood_change >= 0 ? '+' : ''}
                      {analysis.mood_change})
                    </span>
                  )}
                </span>
              </div>
            </div>
          </div>

          {/* Agendas */}
          <div className="grid grid-cols-2 gap-6">
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
          </div>

          {/* Agreements */}
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h3 className="font-semibold mb-4">Договоренности</h3>
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
                  {analysis.agreements.map((agreement, i) => (
                    <tr key={i} className="border-b">
                      <td className="py-2 px-3">{agreement.task}</td>
                      <td className="py-2 px-3">{agreement.responsible}</td>
                      <td className="py-2 px-3">{agreement.deadline || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Development */}
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h3 className="font-semibold mb-3">Развитие сотрудника</h3>
            <p className="text-gray-700">{analysis.development_notes}</p>
          </div>

          {/* Red Flags */}
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

          {/* Recommendations */}
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

          {/* Transcript */}
          <details className="bg-white rounded-lg shadow-sm border">
            <summary className="p-6 cursor-pointer font-semibold">
              Транскрипт встречи
            </summary>
            <div className="px-6 pb-6">
              <pre className="whitespace-pre-wrap text-sm text-gray-700 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto">
                {transcript}
              </pre>
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
