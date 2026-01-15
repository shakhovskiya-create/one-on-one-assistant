<script>
  import { onMount } from 'svelte'
  import { invoke } from '@tauri-apps/api/core'

  let activeTab = 'status'
  let status = {
    running: false,
    connected: false,
    ad_connected: false,
    exchange_connected: false,
    last_error: null,
    logs: []
  }

  let config = {
    backend_url: 'wss://one-on-one-back.up.railway.app/ws/connector',
    api_key: '',
    ad_server: 'ldap://172.20.0.33',
    ad_bind_user: '',
    ad_bind_password: '',
    ad_base_dn: 'DC=ekfgroup,DC=ru',
    ews_url: 'https://post.ekf.su/EWS/Exchange.asmx',
    ews_username: '',
    ews_password: ''
  }

  let saving = false
  let starting = false

  onMount(async () => {
    try {
      config = await invoke('get_config')
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
          <h4>Exchange</h4>
          <div class="value" class:connected={status.exchange_connected} class:disconnected={!status.exchange_connected}>
            {status.exchange_connected ? '● Подключен' : '○ Отключен'}
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
          <input type="password" bind:value={config.api_key} />
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
          <input type="password" bind:value={config.ad_bind_password} />
        </div>
        <div class="form-group">
          <label>Base DN</label>
          <input type="text" bind:value={config.ad_base_dn} placeholder="DC=domain,DC=com" />
        </div>
      </div>

      <div class="form-section">
        <h3>Exchange (EWS)</h3>
        <div class="form-group">
          <label>URL EWS</label>
          <input type="text" bind:value={config.ews_url} placeholder="https://mail.domain.com/EWS/Exchange.asmx" />
        </div>
        <div class="form-group">
          <label>Пользователь</label>
          <input type="text" bind:value={config.ews_username} />
        </div>
        <div class="form-group">
          <label>Пароль</label>
          <input type="password" bind:value={config.ews_password} />
        </div>
      </div>

      <button class="btn btn-primary btn-full" on:click={saveConfig} disabled={saving}>
        {saving ? 'Сохранение...' : 'Сохранить'}
      </button>
    </div>
  {/if}

  <!-- Logs Panel -->
  {#if activeTab === 'logs'}
    <div class="panel">
      <div class="logs">
        {#if status.logs.length === 0}
          <div class="log-entry">Нет логов</div>
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
</style>
