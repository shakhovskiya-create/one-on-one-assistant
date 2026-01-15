'use client'

import { useState, useEffect } from 'react'
import { Play, Pause, RotateCcw, Check, ChevronRight, Clock, User } from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Question {
  text: string
  checked: boolean
  notes: string
}

interface Section {
  id: string
  title: string
  duration: number
  questions: Question[]
  expanded: boolean
}

interface Employee {
  id: string
  name: string
  position: string
}

export default function ScriptPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedEmployee, setSelectedEmployee] = useState<string>('')
  const [sections, setSections] = useState<Section[]>([])
  const [currentSection, setCurrentSection] = useState(0)
  const [timer, setTimer] = useState(0)
  const [isRunning, setIsRunning] = useState(false)
  const [totalTime, setTotalTime] = useState(0)

  useEffect(() => {
    fetchScript()
    fetchEmployees()
  }, [])

  useEffect(() => {
    let interval: NodeJS.Timeout
    if (isRunning) {
      interval = setInterval(() => {
        setTimer((t) => t + 1)
        setTotalTime((t) => t + 1)
      }, 1000)
    }
    return () => clearInterval(interval)
  }, [isRunning])

  const fetchScript = async () => {
    try {
      const res = await fetch(`${API_URL}/script`)
      if (res.ok) {
        const data = await res.json()
        setSections(
          data.sections.map((s: any, i: number) => ({
            ...s,
            questions: s.questions.map((q: string) => ({
              text: q,
              checked: false,
              notes: '',
            })),
            expanded: i === 0,
          }))
        )
      }
    } catch (error) {
      // Use default script if API fails
      setSections(getDefaultScript())
    }
  }

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

  const getDefaultScript = (): Section[] => [
    {
      id: 'checkin',
      title: 'Чекин',
      duration: 5,
      expanded: true,
      questions: [
        { text: 'Как ты? Что нового?', checked: false, notes: '' },
        { text: 'Как прошла неделя?', checked: false, notes: '' },
        { text: 'Что занимает голову прямо сейчас?', checked: false, notes: '' },
        { text: 'Как настроение команды?', checked: false, notes: '' },
      ],
    },
    {
      id: 'employee_agenda',
      title: 'Повестка сотрудника',
      duration: 20,
      expanded: false,
      questions: [
        { text: 'С чем пришел? Что хочешь обсудить?', checked: false, notes: '' },
        { text: 'Где нужна помощь или ресурс?', checked: false, notes: '' },
        { text: 'Что буксует и почему?', checked: false, notes: '' },
        { text: 'Что мешает команде работать эффективнее?', checked: false, notes: '' },
        { text: 'Какие решения ожидаются?', checked: false, notes: '' },
      ],
    },
    {
      id: 'manager_agenda',
      title: 'Повестка руководителя',
      duration: 15,
      expanded: false,
      questions: [
        { text: 'Статус по ключевым проектам', checked: false, notes: '' },
        { text: 'Изменения в приоритетах', checked: false, notes: '' },
        { text: 'Ожидания и сроки', checked: false, notes: '' },
        { text: 'Обратная связь от смежных подразделений', checked: false, notes: '' },
      ],
    },
    {
      id: 'development',
      title: 'Развитие сотрудника',
      duration: 10,
      expanded: false,
      questions: [
        { text: 'Как оцениваешь свою работу за последние 2 недели?', checked: false, notes: '' },
        { text: 'Что получилось хорошо?', checked: false, notes: '' },
        { text: 'Что бы сделал иначе?', checked: false, notes: '' },
        { text: 'Чему хочешь научиться?', checked: false, notes: '' },
        { text: 'Какая поддержка нужна для роста?', checked: false, notes: '' },
      ],
    },
    {
      id: 'feedback',
      title: 'Обратная связь руководителю',
      duration: 5,
      expanded: false,
      questions: [
        { text: 'Что я мог бы делать иначе?', checked: false, notes: '' },
        { text: 'Достаточно ли контекста и информации ты получаешь?', checked: false, notes: '' },
        { text: 'Есть что-то, что хотел сказать, но не решался?', checked: false, notes: '' },
      ],
    },
    {
      id: 'agreements',
      title: 'Договоренности',
      duration: 5,
      expanded: false,
      questions: [
        { text: 'Фиксируем договоренности и сроки', checked: false, notes: '' },
      ],
    },
  ]

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  const toggleSection = (index: number) => {
    setSections((prev) =>
      prev.map((s, i) => ({
        ...s,
        expanded: i === index ? !s.expanded : s.expanded,
      }))
    )
    setCurrentSection(index)
    setTimer(0)
  }

  const toggleQuestion = (sectionIndex: number, questionIndex: number) => {
    setSections((prev) =>
      prev.map((s, si) =>
        si === sectionIndex
          ? {
              ...s,
              questions: s.questions.map((q, qi) =>
                qi === questionIndex ? { ...q, checked: !q.checked } : q
              ),
            }
          : s
      )
    )
  }

  const updateNotes = (sectionIndex: number, questionIndex: number, notes: string) => {
    setSections((prev) =>
      prev.map((s, si) =>
        si === sectionIndex
          ? {
              ...s,
              questions: s.questions.map((q, qi) =>
                qi === questionIndex ? { ...q, notes } : q
              ),
            }
          : s
      )
    )
  }

  const nextSection = () => {
    if (currentSection < sections.length - 1) {
      setSections((prev) =>
        prev.map((s, i) => ({
          ...s,
          expanded: i === currentSection + 1,
        }))
      )
      setCurrentSection((prev) => prev + 1)
      setTimer(0)
    }
  }

  const resetMeeting = () => {
    setSections(getDefaultScript())
    setCurrentSection(0)
    setTimer(0)
    setTotalTime(0)
    setIsRunning(false)
  }

  const getSectionProgress = (section: Section) => {
    const checked = section.questions.filter((q) => q.checked).length
    return Math.round((checked / section.questions.length) * 100)
  }

  const currentSectionData = sections[currentSection]
  const sectionTimeLimit = currentSectionData?.duration * 60
  const isOverTime = timer > sectionTimeLimit

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Скрипт встречи</h1>
        <div className="flex items-center gap-4">
          <select
            value={selectedEmployee}
            onChange={(e) => setSelectedEmployee(e.target.value)}
            className="border rounded-lg px-3 py-2"
          >
            <option value="">Выберите сотрудника</option>
            {employees.map((emp) => (
              <option key={emp.id} value={emp.id}>
                {emp.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Timer Panel */}
      <div className="bg-white rounded-lg shadow-sm border p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div className="text-center">
              <p className="text-sm text-gray-500">Секция</p>
              <p className={`text-3xl font-mono font-bold ${isOverTime ? 'text-red-600' : 'text-gray-900'}`}>
                {formatTime(timer)}
              </p>
              <p className="text-xs text-gray-400">
                из {currentSectionData?.duration || 0} мин
              </p>
            </div>
            <div className="text-center border-l pl-6">
              <p className="text-sm text-gray-500">Всего</p>
              <p className="text-3xl font-mono font-bold text-gray-900">
                {formatTime(totalTime)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsRunning(!isRunning)}
              className={`p-3 rounded-full ${
                isRunning
                  ? 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200'
                  : 'bg-green-100 text-green-700 hover:bg-green-200'
              }`}
            >
              {isRunning ? <Pause size={24} /> : <Play size={24} />}
            </button>
            <button
              onClick={resetMeeting}
              className="p-3 rounded-full bg-gray-100 text-gray-700 hover:bg-gray-200"
            >
              <RotateCcw size={24} />
            </button>
          </div>
        </div>
      </div>

      {/* Sections */}
      <div className="space-y-4">
        {sections.map((section, sectionIndex) => (
          <div
            key={section.id}
            className={`bg-white rounded-lg shadow-sm border overflow-hidden ${
              sectionIndex === currentSection ? 'ring-2 ring-blue-500' : ''
            }`}
          >
            <button
              onClick={() => toggleSection(sectionIndex)}
              className="w-full p-4 flex items-center justify-between hover:bg-gray-50"
            >
              <div className="flex items-center gap-3">
                <ChevronRight
                  className={`transform transition-transform ${
                    section.expanded ? 'rotate-90' : ''
                  }`}
                  size={20}
                />
                <span className="font-semibold">{section.title}</span>
                <span className="text-sm text-gray-500">({section.duration} мин)</span>
              </div>
              <div className="flex items-center gap-3">
                <div className="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-green-500 transition-all"
                    style={{ width: `${getSectionProgress(section)}%` }}
                  />
                </div>
                <span className="text-sm text-gray-500">
                  {getSectionProgress(section)}%
                </span>
              </div>
            </button>

            {section.expanded && (
              <div className="border-t p-4 space-y-3">
                {section.questions.map((question, questionIndex) => (
                  <div key={questionIndex} className="space-y-2">
                    <label className="flex items-start gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={question.checked}
                        onChange={() => toggleQuestion(sectionIndex, questionIndex)}
                        className="mt-1 h-5 w-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      />
                      <span
                        className={`flex-1 ${
                          question.checked ? 'text-gray-400 line-through' : 'text-gray-700'
                        }`}
                      >
                        {question.text}
                      </span>
                    </label>
                    {section.expanded && (
                      <textarea
                        value={question.notes}
                        onChange={(e) =>
                          updateNotes(sectionIndex, questionIndex, e.target.value)
                        }
                        placeholder="Заметки..."
                        className="w-full ml-8 p-2 text-sm border rounded resize-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                        rows={2}
                      />
                    )}
                  </div>
                ))}

                {sectionIndex < sections.length - 1 && (
                  <button
                    onClick={nextSection}
                    className="mt-4 w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center justify-center gap-2"
                  >
                    Следующая секция
                    <ChevronRight size={20} />
                  </button>
                )}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Progress Overview */}
      <div className="bg-white rounded-lg shadow-sm border p-4">
        <h3 className="font-semibold mb-3">Прогресс встречи</h3>
        <div className="flex gap-2">
          {sections.map((section, index) => (
            <div
              key={section.id}
              className={`flex-1 h-2 rounded-full ${
                getSectionProgress(section) === 100
                  ? 'bg-green-500'
                  : index === currentSection
                  ? 'bg-blue-500'
                  : 'bg-gray-200'
              }`}
            />
          ))}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-500">
          {sections.map((section) => (
            <span key={section.id}>{section.title.split(' ')[0]}</span>
          ))}
        </div>
      </div>
    </div>
  )
}
