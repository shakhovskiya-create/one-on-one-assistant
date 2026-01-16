'use client'

import { useState, useEffect } from 'react'
import { Settings, Server, RefreshCw, CheckCircle, XCircle, Users, Calendar, Loader2, ChevronDown, ChevronUp } from 'lucide-react'

import { API_URL } from '@/lib/config'

interface ConnectorStatus {
  connected: boolean
  last_sync: string | null
  ad_status: 'connected' | 'disconnected' | 'unknown'
  exchange_status: 'connected' | 'disconnected' | 'unknown'
}

interface SyncStats {
  mode: string
  total_in_ad: number
  with_department: number
  without_department: number
  filtered_out: number
  new_users: number
  updated_users: number
  skipped_existing: number
  managers_updated: number
  imported: number
}

interface SyncResult {
  success: boolean
  message: string
  stats?: SyncStats
}

type SyncMode = 'full' | 'new_only' | 'changes'

export default function SettingsPage() {
  const [connectorStatus, setConnectorStatus] = useState<ConnectorStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState<string | null>(null)
  const [syncResults, setSyncResults] = useState<Record<string, SyncResult>>({})

  // AD Sync options
  const [syncMode, setSyncMode] = useState<SyncMode>('full')
  const [includePhotos, setIncludePhotos] = useState(true)
  const [requireDepartment, setRequireDepartment] = useState(true)
  const [showAdvanced, setShowAdvanced] = useState(false)

  useEffect(() => {
    checkConnectorStatus()
  }, [])

  const checkConnectorStatus = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${API_URL}/connector/status`)
      if (res.ok) {
        const data = await res.json()
        setConnectorStatus(data)
      } else {
        setConnectorStatus({
          connected: false,
          last_sync: null,
          ad_status: 'disconnected',
          exchange_status: 'disconnected'
        })
      }
    } catch (error) {
      console.error('Failed to check connector status:', error)
      setConnectorStatus({
        connected: false,
        last_sync: null,
        ad_status: 'disconnected',
        exchange_status: 'disconnected'
      })
    } finally {
      setLoading(false)
    }
  }

  const syncAD = async () => {
    setSyncing('ad')
    setSyncResults(prev => ({ ...prev, ad: { success: false, message: 'Синхронизация...' } }))
    try {
      const params = new URLSearchParams({
        mode: syncMode,
        include_photos: includePhotos.toString(),
        require_department: requireDepartment.toString()
      })

      const res = await fetch(`${API_URL}/ad/sync?${params}`, { method: 'POST' })
      const data = await res.json()
      if (res.ok) {
        const stats = data as SyncStats
        const message = `Импортировано ${stats.imported} (новых: ${stats.new_users}, обновлено: ${stats.updated_users})`
        setSyncResults(prev => ({
          ...prev,
          ad: { success: true, message, stats }
        }))
      } else {
        setSyncResults(prev => ({
          ...prev,
          ad: { success: false, message: data.detail || 'Ошибка синхронизации' }
        }))
      }
    } catch (error) {
      setSyncResults(prev => ({
        ...prev,
        ad: { success: false, message: 'Коннектор недоступен' }
      }))
    } finally {
      setSyncing(null)
    }
  }

  const syncCalendar = async () => {
    setSyncing('calendar')
    setSyncResults(prev => ({ ...prev, calendar: { success: false, message: 'Синхронизация...' } }))
    try {
      const res = await fetch(`${API_URL}/calendar/sync`, { method: 'POST' })
      const data = await res.json()
      if (res.ok) {
        setSyncResults(prev => ({
          ...prev,
          calendar: { success: true, message: `Синхронизировано ${data.synced || 0} из ${data.total || 0} встреч`, count: data.synced }
        }))
      } else {
        setSyncResults(prev => ({
          ...prev,
          calendar: { success: false, message: data.detail || 'Ошибка синхронизации' }
        }))
      }
    } catch (error) {
      setSyncResults(prev => ({
        ...prev,
        calendar: { success: false, message: 'Коннектор недоступен' }
      }))
    } finally {
      setSyncing(null)
    }
  }

  const StatusIcon = ({ status }: { status: 'connected' | 'disconnected' | 'unknown' }) => {
    if (status === 'connected') return <CheckCircle size={18} className="text-green-500" />
    if (status === 'disconnected') return <XCircle size={18} className="text-red-500" />
    return <div className="w-4 h-4 rounded-full bg-gray-300" />
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-orange"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Settings size={28} className="text-ekf-orange" />
        <h1 className="text-2xl font-bold text-ekf-dark">Настройки</h1>
      </div>

      {/* Connector Status */}
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="p-4 border-b">
          <h2 className="text-lg font-semibold text-ekf-dark flex items-center gap-2">
            <Server size={20} />
            On-Prem Коннектор
          </h2>
        </div>
        <div className="p-6">
          <div className="flex items-center gap-4 mb-6">
            <div className={`w-3 h-3 rounded-full ${connectorStatus?.connected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span className={`font-medium ${connectorStatus?.connected ? 'text-green-600' : 'text-red-600'}`}>
              {connectorStatus?.connected ? 'Подключен' : 'Не подключен'}
            </span>
            <button
              onClick={checkConnectorStatus}
              className="ml-auto text-ekf-gray hover:text-ekf-orange transition-colors"
            >
              <RefreshCw size={18} />
            </button>
          </div>

          {!connectorStatus?.connected && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
              <p className="text-yellow-800 text-sm">
                Для синхронизации с Active Directory и Exchange необходимо запустить on-prem коннектор
                внутри корпоративной сети. Подробная инструкция в README проекта.
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* AD Status */}
            <div className="border rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Users size={18} className="text-ekf-orange" />
                  <span className="font-medium">Active Directory</span>
                </div>
                <StatusIcon status={connectorStatus?.ad_status || 'unknown'} />
              </div>

              {/* Sync Mode */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-ekf-dark mb-2">Режим синхронизации</label>
                <select
                  value={syncMode}
                  onChange={(e) => setSyncMode(e.target.value as SyncMode)}
                  className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:border-ekf-orange focus:ring-1 focus:ring-ekf-orange outline-none"
                >
                  <option value="full">Полная синхронизация</option>
                  <option value="new_only">Только новые</option>
                  <option value="changes">Новые + изменения</option>
                </select>
              </div>

              {/* Advanced Options Toggle */}
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-1 text-sm text-ekf-gray mb-3 hover:text-ekf-orange"
              >
                {showAdvanced ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                Дополнительные опции
              </button>

              {showAdvanced && (
                <div className="space-y-3 mb-4 p-3 bg-gray-50 rounded-lg">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={requireDepartment}
                      onChange={(e) => setRequireDepartment(e.target.checked)}
                      className="rounded border-gray-300 text-ekf-orange focus:ring-ekf-orange"
                    />
                    <span className="text-sm">Только с департаментом</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={includePhotos}
                      onChange={(e) => setIncludePhotos(e.target.checked)}
                      className="rounded border-gray-300 text-ekf-orange focus:ring-ekf-orange"
                    />
                    <span className="text-sm">Загружать фото (медленнее)</span>
                  </label>
                </div>
              )}

              <button
                onClick={syncAD}
                disabled={syncing === 'ad' || !connectorStatus?.connected}
                className="w-full flex items-center justify-center gap-2 py-2 px-4 bg-ekf-orange text-white rounded-lg hover:bg-ekf-orange-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {syncing === 'ad' ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Синхронизация...
                  </>
                ) : (
                  <>
                    <RefreshCw size={16} />
                    Синхронизировать AD
                  </>
                )}
              </button>

              {/* Sync Results */}
              {syncResults.ad && (
                <div className={`mt-3 p-3 rounded-lg ${syncResults.ad.success ? 'bg-green-50' : 'bg-red-50'}`}>
                  <p className={`text-sm font-medium ${syncResults.ad.success ? 'text-green-700' : 'text-red-700'}`}>
                    {syncResults.ad.message}
                  </p>
                  {syncResults.ad.stats && (
                    <div className="mt-2 grid grid-cols-2 gap-2 text-xs text-gray-600">
                      <div>Всего в AD: {syncResults.ad.stats.total_in_ad}</div>
                      <div>С департаментом: {syncResults.ad.stats.with_department}</div>
                      <div>Новых: {syncResults.ad.stats.new_users}</div>
                      <div>Обновлено: {syncResults.ad.stats.updated_users}</div>
                      <div>Пропущено: {syncResults.ad.stats.skipped_existing}</div>
                      <div>Связей: {syncResults.ad.stats.managers_updated}</div>
                    </div>
                  )}
                </div>
              )}
            </div>

            {/* Exchange Status */}
            <div className="border rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Calendar size={18} className="text-ekf-orange" />
                  <span className="font-medium">Exchange / Outlook</span>
                </div>
                <StatusIcon status={connectorStatus?.exchange_status || 'unknown'} />
              </div>
              <p className="text-sm text-ekf-gray mb-4">
                Синхронизация календаря и встреч
              </p>
              <button
                onClick={syncCalendar}
                disabled={syncing === 'calendar' || !connectorStatus?.connected}
                className="w-full flex items-center justify-center gap-2 py-2 px-4 bg-ekf-orange text-white rounded-lg hover:bg-ekf-orange-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {syncing === 'calendar' ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Синхронизация...
                  </>
                ) : (
                  <>
                    <RefreshCw size={16} />
                    Синхронизировать календарь
                  </>
                )}
              </button>
              {syncResults.calendar && (
                <p className={`mt-2 text-sm ${syncResults.calendar.success ? 'text-green-600' : 'text-red-600'}`}>
                  {syncResults.calendar.message}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Info */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h2 className="text-lg font-semibold text-ekf-dark mb-4">Информация</h2>
        <div className="space-y-3 text-sm">
          <div className="flex justify-between">
            <span className="text-ekf-gray">Последняя синхронизация AD</span>
            <span className="text-ekf-dark">
              {connectorStatus?.last_sync
                ? new Date(connectorStatus.last_sync).toLocaleString('ru-RU')
                : 'Не выполнялась'}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-ekf-gray">Backend API</span>
            <span className="text-ekf-dark font-mono text-xs">{API_URL}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
