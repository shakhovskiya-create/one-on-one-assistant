<script>
  import { onMount } from 'svelte'
  import { invoke } from '@tauri-apps/api/core'

  let activeTab = 'status'
  let status = {
    running: false,
    connected: false,
    ad_connected: false,
    last_error: null,
    logs: []
  }

  let config = {
    backend_url: 'wss://one-on-one-back.up.railway.app/ws/connector',
    api_key: '',
    ad_server: 'ldap://172.20.0.33',
    ad_bind_user: '',
    ad_bind_password: '',
    ad_base_dn: 'DC=ekfgroup,DC=ru'
  }

  let originalConfig = null
  let saving = false
  let starting = false

  // Password visibility states
  let showApiKey = false
  let showAdPassword = false

  // Check if config has changed
  $: hasChanges = originalConfig && JSON.stringify(config) !== JSON.stringify(originalConfig)

  onMount(async () => {
    try {
      config = await invoke('get_config')
      originalConfig = JSON.parse(JSON.stringify(config))
      status = await invoke('get_status')
    } catch (e) {
      console.error('Failed to load config:', e)
    }

    // Poll status
    setInterval(async () => {
      try {
        status = await invoke('get_status')
      } catch (e) {
        console.error('Failed to get status:', e)
      }
    }, 2000)
  })

  async function saveConfig() {
    saving = true
    try {
      await invoke('save_config', { config })
      originalConfig = JSON.parse(JSON.stringify(config))
      alert('Настройки сохранены')
    } catch (e) {
      alert('Ошибка сохранения: ' + e)
    } finally {
      saving = false
    }
  }

  async function toggleConnector() {
    starting = true
    try {
      if (status.running) {
        await invoke('stop_connector')
      } else {
        await invoke('start_connector')
      }
      status = await invoke('get_status')
    } catch (e) {
      alert('Ошибка: ' + e)
    } finally {
      starting = false
    }
  }

  async function clearLogs() {
    await invoke('clear_logs')
    status = await invoke('get_status')
  }
</script>

<div class="app">
  <!-- Header -->
  <header class="header">
    <div class="logo">EKF</div>
    <h1>Connector</h1>
  </header>

  <!-- Status Bar -->
  <div class="status">
    <div class="status-dot" class:connected={status.connected}></div>
    <span class="status-text">
      {status.connected ? 'Подключен' : status.running ? 'Подключение...' : 'Отключен'}
    </span>
  </div>

  <!-- Tabs -->
  <div class="tabs">
    <button
      class="tab"
      class:active={activeTab === 'status'}
      on:click={() => activeTab = 'status'}
    >
      Статус
    </button>
    <button
      class="tab"
      class:active={activeTab === 'settings'}
      on:click={() => activeTab = 'settings'}
    >
      Настройки
    </button>
    <button
      class="tab"
      class:active={activeTab === 'logs'}
      on:click={() => activeTab = 'logs'}
    >
      Логи
    </button>
  </div>

  <!-- Status Panel -->
  {#if activeTab === 'status'}
    <div class="panel">
      <div class="status-cards">
        <div class="status-card">
          <h4>Backend</h4>
          <div class="value" class:connected={status.connected} class:disconnected={!status.connected}>
            {status.connected ? '● Подключен' : '○ Отключен'}
          </div>
        </div>
        <div class="status-card">
          <h4>Active Directory</h4>
          <div class="value" class:connected={status.ad_connected} class:disconnected={!status.ad_connected}>
            {status.ad_connected ? '● Подключен' : '○ Отключен'}
          </div>
        </div>
        <div class="status-card">
          <h4>Статус</h4>
          <div class="value">
            {status.running ? 'Запущен' : 'Остановлен'}
          </div>
        </div>
      </div>

      {#if status.last_error}
        <div class="error-box">
          <strong>Ошибка:</strong> {status.last_error}
        </div>
      {/if}

      <div class="actions">
        <button
          class="btn btn-full"
          class:btn-primary={!status.running}
          class:btn-danger={status.running}
          on:click={toggleConnector}
          disabled={starting}
        >
          {starting ? 'Подождите...' : status.running ? 'Остановить' : 'Запустить'}
        </button>
      </div>
    </div>
  {/if}

  <!-- Settings Panel -->
  {#if activeTab === 'settings'}
    <div class="panel">
      <div class="form-section">
        <h3>Backend</h3>
        <div class="form-group">
          <label>URL бэкенда</label>
          <input type="text" bind:value={config.backend_url} placeholder="wss://..." />
        </div>
        <div class="form-group">
          <label>API Key</label>
          <div class="password-input">
            {#if showApiKey}
              <input type="text" bind:value={config.api_key} />
            {:else}
              <input type="password" bind:value={config.api_key} />
            {/if}
            <button type="button" class="eye-btn" on:click={() => showApiKey = !showApiKey}>
              {showApiKey ? '🙈' : '👁'}
            </button>
          </div>
        </div>
      </div>

      <div class="form-section">
        <h3>Active Directory</h3>
        <div class="form-group">
          <label>Сервер LDAP</label>
          <input type="text" bind:value={config.ad_server} placeholder="ldap://..." />
        </div>
        <div class="form-group">
          <label>Пользователь</label>
          <input type="text" bind:value={config.ad_bind_user} placeholder="DOMAIN\username" />
        </div>
        <div class="form-group">
          <label>Пароль</label>
          <div class="password-input">
            {#if showAdPassword}
              <input type="text" bind:value={config.ad_bind_password} />
            {:else}
              <input type="password" bind:value={config.ad_bind_password} />
            {/if}
            <button type="button" class="eye-btn" on:click={() => showAdPassword = !showAdPassword}>
              {showAdPassword ? '🙈' : '👁'}
            </button>
          </div>
        </div>
        <div class="form-group">
          <label>Base DN</label>
          <input type="text" bind:value={config.ad_base_dn} placeholder="DC=domain,DC=com" />
        </div>
      </div>

      <button
        class="btn btn-full"
        class:btn-primary={hasChanges}
        class:btn-disabled={!hasChanges}
        on:click={saveConfig}
        disabled={saving || !hasChanges}
      >
        {saving ? 'Сохранение...' : hasChanges ? 'Сохранить' : 'Сохранено'}
      </button>
    </div>
  {/if}

  <!-- Logs Panel -->
  {#if activeTab === 'logs'}
    <div class="panel">
      <div class="logs">
        {#if status.logs.length === 0}
          <div class="log-entry log-empty">Нет логов</div>
        {:else}
          {#each status.logs as log}
            <div class="log-entry">{log}</div>
          {/each}
        {/if}
      </div>
      <div class="actions">
        <button class="btn btn-secondary btn-full" on:click={clearLogs}>
          Очистить логи
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .error-box {
    background: #FEF2F2;
    border: 1px solid #FECACA;
    border-radius: 6px;
    padding: 12px;
    margin-bottom: 16px;
    font-size: 12px;
    color: #DC2626;
  }

  .password-input {
    display: flex;
    gap: 8px;
  }

  .password-input input {
    flex: 1;
  }

  .eye-btn {
    background: #f3f4f6;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    padding: 8px 12px;
    cursor: pointer;
    font-size: 14px;
  }

  .eye-btn:hover {
    background: #e5e7eb;
  }

  .btn-disabled {
    background: #e5e7eb;
    color: #9ca3af;
    cursor: not-allowed;
  }

  .log-empty {
    color: #9ca3af;
    font-style: italic;
  }
</style>
