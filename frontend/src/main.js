import './style.css';

// PWA Service Worker & Install Prompt Registration
let deferredInstallPrompt = null;

if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/sw.js')
      .then((reg) => console.log('OpsFlow PWA SW registered:', reg.scope))
      .catch((err) => console.warn('SW registration failed:', err));
  });
}

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredInstallPrompt = e;
  const pwaBtn = document.getElementById('pwa-install-btn');
  if (pwaBtn) pwaBtn.style.display = 'inline-flex';
});

const API_URL = import.meta.env.VITE_API_URL || '';

// Theme Management (Light / Dark)
function initTheme() {
  const savedTheme = localStorage.getItem('opsflow_theme') || 'dark';
  document.documentElement.setAttribute('data-theme', savedTheme);
  updateThemeToggleIcon(savedTheme);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const next = current === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', next);
  localStorage.setItem('opsflow_theme', next);
  updateThemeToggleIcon(next);
}

function updateThemeToggleIcon(theme) {
  const btn = document.getElementById('theme-toggle-btn');
  if (!btn) return;
  btn.innerHTML = theme === 'dark' ? '🌞' : '🌙';
  btn.setAttribute('title', theme === 'dark' ? 'Switch to Light Theme' : 'Switch to Dark Theme');
}

function getAuthToken() {
  return localStorage.getItem('opsflow_token') || '';
}

function getAuthHeaders() {
  const headers = { 'Content-Type': 'application/json' };
  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

// Router
const routes = {
  '/': renderDashboard,
  '/incidents': renderIncidents,
  '/services': renderServices,
  '/ai': renderAIAssistant,
  '/l3-workbench': renderL3Workbench,
  '/knowledge-docs': renderKnowledgeDocs,
};

function navigate(path) {
  window.history.pushState({}, '', path);
  render();
}

function render() {
  const path = window.location.pathname;
  const route = routes[path] || renderDashboard;
  const content = document.getElementById('content');
  if (content) {
    content.innerHTML = '';
    route(content);
  }
  document.querySelectorAll('.nav-link').forEach((link) => {
    link.classList.toggle('active', link.dataset.path === path);
  });
}

window.addEventListener('popstate', render);

// Layout & Initialization
function createApp() {
  initTheme();

  const app = document.getElementById('app');
  app.innerHTML = `
    <aside class="sidebar" id="sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <div class="logo-icon">O</div>
          <span class="logo-text">OpsFlow</span>
        </div>
        <button class="sidebar-toggle" id="sidebar-toggle" aria-label="Toggle sidebar">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path d="M3 5h14M3 10h14M3 15h14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
        </button>
      </div>
      <nav class="nav">
        <a class="nav-link active" data-path="/" id="nav-dashboard">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><rect x="2" y="2" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="11" y="2" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="2" y="11" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="11" y="11" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/></svg>
          <span>Dashboard</span>
        </a>
        <a class="nav-link" data-path="/incidents" id="nav-incidents">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><path d="M10 2L18 17H2L10 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/><path d="M10 8v4M10 14v.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          <span>Incidents</span>
        </a>
        <a class="nav-link" data-path="/services" id="nav-services">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><rect x="3" y="3" width="14" height="5" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="12" width="14" height="5" rx="1.5" stroke="currentColor" stroke-width="1.5"/><circle cx="6" cy="5.5" r="1" fill="currentColor"/><circle cx="6" cy="14.5" r="1" fill="currentColor"/></svg>
          <span>Service Catalog</span>
        </a>
        <a class="nav-link" data-path="/ai" id="nav-ai">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><circle cx="10" cy="10" r="7.5" stroke="currentColor" stroke-width="1.5"/><path d="M7 10a3 3 0 016 0M10 7v0M8 13h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          <span>AI & Approvals</span>
        </a>
        <a class="nav-link" data-path="/l3-workbench" id="nav-l3-workbench">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><path d="M7 4h6M4 8h12M6 12h8M8 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="4" cy="4" r="1" fill="currentColor"/><circle cx="16" cy="16" r="1" fill="currentColor"/></svg>
          <span>L3 Engineering</span>
        </a>
        <a class="nav-link" data-path="/knowledge-docs" id="nav-knowledge-docs">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none"><path d="M4 3h12a1 1 0 011 1v12a1 1 0 01-1 1H4a1 1 0 01-1-1V4a1 1 0 011-1z" stroke="currentColor" stroke-width="1.5"/><path d="M7 7h6M7 11h6M7 15h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          <span>Docs & Playbooks</span>
        </a>
      </nav>
      <div class="sidebar-footer">
        <div class="env-badge">KNOWLEDGE HUB</div>
      </div>
    </aside>
    <main class="main">
      <header class="topbar">
        <button class="mobile-menu" id="mobile-menu" aria-label="Open menu">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none"><path d="M3 6h18M3 12h18M3 18h18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        </button>
        <div class="search-container">
          <svg class="search-icon" width="16" height="16" viewBox="0 0 20 20" fill="none"><circle cx="8.5" cy="8.5" r="5.5" stroke="currentColor" stroke-width="1.5"/><path d="M17.5 17.5L12.5 12.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          <input type="text" class="search-input" id="global-search" placeholder="Search incidents & services (Ctrl+K)..." />
          <div class="search-results-overlay" id="search-overlay"></div>
        </div>
        <div class="topbar-right">
          <button class="theme-toggle-btn" id="theme-toggle-btn" aria-label="Toggle theme">🌞</button>
          <button class="btn-pwa-install" id="pwa-install-btn">
            <svg width="16" height="16" viewBox="0 0 20 20" fill="none"><path d="M10 2v10M6 8l4 4 4-4M4 16h12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            Install App
          </button>
          <div class="status-indicator" id="api-status">
            <span class="status-dot"></span>
            <span class="status-text">Connecting...</span>
          </div>
        </div>
      </header>
      <div class="content" id="content"></div>
    </main>

    <!-- Modal Container -->
    <div class="modal-overlay" id="modal-overlay">
      <div class="modal" id="modal-container"></div>
    </div>
  `;

  // Attach event listeners
  document.querySelectorAll('.nav-link').forEach((link) => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      navigate(link.dataset.path);
      document.getElementById('sidebar').classList.remove('open');
    });
  });

  document.getElementById('sidebar-toggle').addEventListener('click', () => {
    document.getElementById('sidebar').classList.toggle('collapsed');
  });

  document.getElementById('mobile-menu').addEventListener('click', () => {
    document.getElementById('sidebar').classList.toggle('open');
  });

  document.getElementById('theme-toggle-btn').addEventListener('click', toggleTheme);

  const pwaBtn = document.getElementById('pwa-install-btn');
  pwaBtn.addEventListener('click', async () => {
    if (deferredInstallPrompt) {
      deferredInstallPrompt.prompt();
      const choice = await deferredInstallPrompt.userChoice;
      if (choice.outcome === 'accepted') {
        pwaBtn.style.display = 'none';
      }
      deferredInstallPrompt = null;
    }
  });

  const searchInput = document.getElementById('global-search');
  searchInput.addEventListener('input', debounce((e) => handleSearch(e.target.value), 300));

  updateThemeToggleIcon(localStorage.getItem('opsflow_theme') || 'dark');
  render();
  checkApiHealth();
}

function debounce(func, delay) {
  let timer;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => func.apply(this, args), delay);
  };
}

async function handleSearch(query) {
  const overlay = document.getElementById('search-overlay');
  if (!query || query.trim().length < 2) {
    overlay.classList.remove('open');
    return;
  }

  try {
    const res = await fetch(`${API_URL}/api/v1/search?q=${encodeURIComponent(query)}`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) return;
    const json = await res.json();
    const data = json.data || {};

    let html = '';
    if (data.incidents && data.incidents.length > 0) {
      html += `<div style="padding: 8px 12px; font-size: 0.75rem; color: var(--text-muted); font-weight: 600;">INCIDENTS</div>`;
      data.incidents.forEach((inc) => {
        html += `<div class="search-item" onclick="navigate('/incidents')">
          <strong>${inc.title || 'Incident'}</strong>
          <span style="font-size: 0.8rem; color: var(--text-secondary); display: block;">${inc.incident_key || ''} • ${inc.severity || ''}</span>
        </div>`;
      });
    }

    if (data.services && data.services.length > 0) {
      html += `<div style="padding: 8px 12px; font-size: 0.75rem; color: var(--text-muted); font-weight: 600;">SERVICES</div>`;
      data.services.forEach((s) => {
        html += `<div class="search-item" onclick="navigate('/services')">
          <strong>${s.name || 'Service'}</strong>
          <span style="font-size: 0.8rem; color: var(--text-secondary); display: block;">${s.description || ''}</span>
        </div>`;
      });
    }

    if (!html) {
      html = `<div style="padding: 16px; color: var(--text-muted); text-align: center; font-size: 0.85rem;">No matching results</div>`;
    }

    overlay.innerHTML = html;
    overlay.classList.add('open');
  } catch (err) {
    overlay.classList.remove('open');
  }
}

// ──────────────────────────────────────────────
// Pages
// ──────────────────────────────────────────────

function renderDashboard(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Operations Dashboard</h1>
        <p class="page-subtitle">Real-time system health and active operational workflow</p>
      </div>
      <div style="display: flex; gap: 10px;">
        <button class="btn btn-primary btn-sm" onclick="openDeclareIncidentModal()">+ Declare Incident</button>
        <button class="btn btn-secondary btn-sm" onclick="openRegisterServiceModal()">+ Register Service</button>
      </div>
    </div>
    <div class="card-grid">
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-primary)">
          <svg viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="18" height="7" rx="2" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="14" width="18" height="7" rx="2" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value" id="val-services">5</span>
          <span class="stat-label">Active Microservices</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-rose)">
          <svg viewBox="0 0 24 24" fill="none"><path d="M12 2L22 20H2L12 2Z" stroke="currentColor" stroke-width="1.5"/><path d="M12 9v5M12 16v.5" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value" id="val-incidents">1</span>
          <span class="stat-label">Open Incidents</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-amber)">
          <svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.5"/><path d="M12 7v5l3 3" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value" id="val-approvals">1</span>
          <span class="stat-label">Pending AI Approvals</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-emerald)">
          <svg viewBox="0 0 24 24" fill="none"><path d="M22 12h-4l-3 9L9 3l-3 9H2" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value" id="val-health">99.9%</span>
          <span class="stat-label">System Uptime</span>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="card-header">
        <h2>Live System Activity</h2>
        <span class="badge badge-medium">CONNECTED</span>
      </div>
      <div class="card-body" id="recent-activity-list">
        <div style="display: flex; flex-direction: column; gap: 12px;">
          <div style="display: flex; justify-content: space-between; align-items: center; padding-bottom: 8px; border-bottom: 1px solid var(--border-glass);">
            <div>
              <strong>Payment Database Timeout</strong>
              <span style="display: block; font-size: 0.8rem; color: var(--text-secondary);">Severity: CRITICAL • State: INVESTIGATING</span>
            </div>
            <span class="badge badge-critical">CRITICAL</span>
          </div>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <strong>Human Approval Pending: restart_service</strong>
              <span style="display: block; font-size: 0.8rem; color: var(--text-secondary);">Requested by AI Model Router for payment-service</span>
            </div>
            <span class="badge badge-pending">PENDING</span>
          </div>
        </div>
      </div>
    </div>
  `;
}

function renderIncidents(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Incident Lifecycle</h1>
        <p class="page-subtitle">Manage, investigate, and resolve system incidents</p>
      </div>
      <button class="btn btn-primary" onclick="openDeclareIncidentModal()">+ Declare Incident</button>
    </div>
    <div class="card">
      <div class="card-header">
        <h2>Active Incidents</h2>
        <span class="badge badge-medium">STATE MACHINE ENABLED</span>
      </div>
      <div class="card-body">
        <table class="data-table">
          <thead>
            <tr>
              <th>Incident Key</th>
              <th>Title</th>
              <th>Severity</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>INC-2026-001</code></td>
              <td><strong>Payment DB Latency Timeout</strong></td>
              <td><span class="badge badge-critical">CRITICAL</span></td>
              <td><span class="badge badge-pending">INVESTIGATING</span></td>
              <td>
                <button class="btn btn-secondary btn-sm" onclick="alert('Transitioning incident state to MITIGATING')">Mitigate</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  `;
}

function renderServices(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Service Catalog</h1>
        <p class="page-subtitle">Microservice registry, dependencies, and environments</p>
      </div>
      <button class="btn btn-primary" onclick="openRegisterServiceModal()">+ Register Service</button>
    </div>
    <div class="card-grid">
      <div class="card" style="margin-bottom:0;">
        <div class="card-header">
          <h2>auth-service</h2>
          <span class="badge badge-approved">HEALTHY</span>
        </div>
        <div class="card-body">
          <p style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 12px;">Authentication & Role-Based Access Control</p>
          <div style="font-size: 0.8rem; color: var(--text-muted);">Owner: Core Engineering</div>
        </div>
      </div>
      <div class="card" style="margin-bottom:0;">
        <div class="card-header">
          <h2>payment-service</h2>
          <span class="badge badge-critical">DEGRADED</span>
        </div>
        <div class="card-body">
          <p style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 12px;">Core Banking Payment Processor</p>
          <div style="font-size: 0.8rem; color: var(--text-muted);">Owner: Payments Squad</div>
        </div>
      </div>
    </div>
  `;
}

function renderAIAssistant(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>AI Assistant & Human Approvals</h1>
        <p class="page-subtitle">Operational copilot with human-in-the-loop safety guardrails</p>
      </div>
      <div style="display:flex; gap:10px; align-items:center;">
        <span style="font-size:0.85rem; color:var(--text-secondary);">Model:</span>
        <select class="form-control" style="width: auto; padding: 4px 10px;" id="llm-model-select">
          <option value="mock">Mock LLM (Local)</option>
          <option value="ollama">Ollama (Local Llama3)</option>
          <option value="cloud">Cloud (OpenAI / Gemini)</option>
        </select>
      </div>
    </div>

    <div class="card" style="margin-bottom: 24px;">
      <div class="card-header">
        <h2>Human-in-the-Loop Approval Queue</h2>
        <span class="badge badge-pending" id="pending-count-badge">1 PENDING</span>
      </div>
      <div class="card-body" id="approval-queue-body">
        <div class="approval-card">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <strong>Action Requested: <code>restart_service</code></strong>
            <span class="badge badge-pending">PENDING APPROVAL</span>
          </div>
          <p style="font-size: 0.85rem; color: var(--text-secondary);">Target Service: <code>payment-service</code> (Environment: production)</p>
          <div style="display: flex; gap: 10px; margin-top: 6px;">
            <button class="btn btn-success btn-sm" onclick="handleApproveAction()">Approve Execution</button>
            <button class="btn btn-danger btn-sm" onclick="handleRejectAction()">Reject Action</button>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h2>AI Console Chat</h2>
      </div>
      <div class="chat-container">
        <div class="chat-messages" id="chat-messages">
          <div class="chat-bubble system">OpsFlow AI Assistant Ready. Human approval workflow active.</div>
          <div class="chat-bubble user">Please check payment-service and restart it if connection pool is full.</div>
          <div class="chat-bubble assistant">Checking service status... Generated human approval request for sensitive action 'restart_service'.</div>
        </div>
        <div class="chat-input-bar">
          <input type="text" class="form-control" id="chat-input" placeholder="Type prompt (e.g. 'What is the status of payment-service?')..." />
          <button class="btn btn-primary" onclick="sendChatMessage()">Send</button>
        </div>
      </div>
    </div>
  `;
}

function renderL3Workbench(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>L3 Engineering Workbench</h1>
        <p class="page-subtitle">Deep diagnostics, stacktrace RCA analysis, DB pool inspection, and playbooks</p>
      </div>
      <div style="display:flex; gap:10px;">
        <button class="btn btn-primary btn-sm" onclick="runDBDiagnostics()">Run DB Diagnostics</button>
        <button class="btn btn-secondary btn-sm" onclick="generateIncidentRCA()">Generate RCA Report</button>
      </div>
    </div>

    <div class="card-grid">
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-primary)">
          <svg viewBox="0 0 24 24" fill="none"><path d="M12 2v20M2 12h20" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value">98/100</span>
          <span class="stat-label">PostgreSQL Pool Conns</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-amber)">
          <svg viewBox="0 0 24 24" fill="none"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value">4.2 ms</span>
          <span class="stat-label">P99 Query Latency</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-icon" style="--accent: var(--accent-emerald)">
          <svg viewBox="0 0 24 24" fill="none"><path d="M5 13l4 4L19 7" stroke="currentColor" stroke-width="1.5"/></svg>
        </div>
        <div class="stat-card-body">
          <span class="stat-value">0</span>
          <span class="stat-label">Deadlocks Detected</span>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 24px;">
      <div class="card-header">
        <h2>Stacktrace & Error Log RCA Analyzer</h2>
        <span class="badge badge-medium">AI DIAGNOSTIC ENGINE</span>
      </div>
      <div class="card-body">
        <div class="form-group">
          <label>Paste Raw Stacktrace or Panic Log</label>
          <textarea class="form-control" id="stacktrace-input" rows="4" style="font-family: var(--font-mono); font-size: 0.85rem;" placeholder="e.g. panic: runtime error: invalid memory address or nil pointer dereference at github.com/opsflow/service..."></textarea>
        </div>
        <button class="btn btn-primary" onclick="analyzeStacktraceInput()">Analyze Stacktrace with AI</button>
        <div id="stacktrace-result" style="margin-top: 16px; display: none;"></div>
      </div>
    </div>
  `;
}

// ──────────────────────────────────────────────
// Docs & Playbook Workspace View (Confluence/Jira Style)
// ──────────────────────────────────────────────

const preloadedDocs = {
  rca_postmortem: `# Root Cause Analysis (RCA) — Incident Post-Mortem

## Incident Overview
- **Incident Key**: INC-2026-001
- **Severity**: CRITICAL (P1)
- **Impacted Service**: \`payment-service\`
- **Start Time**: 2026-08-09T17:45:00Z
- **Resolution Time**: 2026-08-09T18:15:00Z

## Executive Summary
Payment processing API experienced HTTP 504 gateway timeouts due to PostgreSQL connection pool exhaustion.

## Root Cause
Unindexed \`SELECT ... FOR UPDATE\` query in the Outbox processor coupled with a 3x traffic spike caused query execution locks to pile up, exhausting the max 100 \`pgx\` pool connections.

## Remediation Steps
1. Executed emergency index addition on \`outbox_events(status, created_at)\`.
2. Increased \`pgx\` connection pool \`max_conns\` from 100 to 250.
3. Enabled Redis read-through caching in Service Registry.

## Action Items
- [x] Add DB pool monitoring alerts in Prometheus.
- [ ] Implement query timeout circuit breaker at API Gateway level.`,

  sop_db_pool: `# Standard Operating Procedure (SOP) — DB Connection Pool Exhaustion

## Trigger Criteria
- Prometheus Alert: \`PostgreSQLPoolUsage > 90%\`
- API Gateway returns HTTP 504 Gateway Timeout for \`payment-service\` or \`incident-service\`.

## Diagnostic Steps
1. Navigate to **L3 Engineering Workbench** -> Click **Run DB Diagnostics**.
2. Run PostgreSQL lock query:
\`\`\`sql
SELECT pid, query, state, age(clock_timestamp(), query_start) 
FROM pg_stat_activity 
WHERE state != 'idle' 
ORDER BY age DESC;
\`\`\`

## Recovery Actions
1. **Drain Idle Connections**: Trigger L3 Playbook \`DB Pool Connection Drain\`.
2. **Scale DB Pool Thresholds**: Update \`DB_MAX_CONNS=250\` in K8s ConfigMap.
3. **Flush Redis Cache**: Execute sensitive action \`flush_redis_cache\` if cache stale.`,

  oncall_matrix: `# On-Call Escalation Matrix & Basic Knowledge Base

## Escalation Tiers
| Tier | Role | Primary Contact | SLA Response |
|------|------|-----------------|--------------|
| Tier 1 | L1 Operations Desk | \`ops-desk@opsflow.local\` | 15 mins |
| Tier 2 | L2 DevOps Squad | \`devops-oncall@opsflow.local\` | 30 mins |
| Tier 3 | L3 Core Infrastructure Engineer | \`l3-eng@opsflow.local\` | Immediate |

## Architecture Overview
- **API Gateway**: Port 8080
- **Auth Service**: Port 8081 (JWT + Bcrypt)
- **Incident Service**: Port 8082 (Outbox Pattern)
- **Service Registry**: Port 8083 (Redis Cached)
- **AI Gateway**: Port 8084 (Model Router & Tool Guardrails)
- **Notification Worker**: Port 8085 (RabbitMQ AMQP)`
};

let currentDocKey = 'rca_postmortem';
let activeViewMode = 'edit'; // 'edit' | 'preview' | 'split'

function renderKnowledgeDocs(container) {
  container.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Docs & Playbook Workspace</h1>
        <p class="page-subtitle">Jira & Confluence style documentation, SOP playbooks, and RCA templates</p>
      </div>
      <div style="display:flex; gap:10px;">
        <button class="btn btn-primary btn-sm" onclick="saveCurrentDoc()">Save Document</button>
        <button class="btn btn-secondary btn-sm" onclick="exportMarkdownDoc()">Export Markdown</button>
      </div>
    </div>

    <div class="docs-workspace">
      <!-- Left Pane: Template & Document Library -->
      <div class="docs-sidebar-panel">
        <h3 style="font-size: 0.9rem; color: var(--text-muted); font-weight: 700; text-transform: uppercase;">TEMPLATES & DOCS</h3>
        <div class="docs-tree-list">
          <div class="docs-tree-item ${currentDocKey === 'rca_postmortem' ? 'active' : ''}" onclick="selectDocTemplate('rca_postmortem')">
            <span>📄 Incident RCA Post-Mortem</span>
          </div>
          <div class="docs-tree-item ${currentDocKey === 'sop_db_pool' ? 'active' : ''}" onclick="selectDocTemplate('sop_db_pool')">
            <span>📋 SOP: DB Pool Exhaustion</span>
          </div>
          <div class="docs-tree-item ${currentDocKey === 'oncall_matrix' ? 'active' : ''}" onclick="selectDocTemplate('oncall_matrix')">
            <span>🧠 On-Call Escalation & Architecture</span>
          </div>
        </div>
        <button class="btn btn-secondary btn-sm" style="width: 100%;" onclick="createNewDoc()">+ New Custom Doc</button>
      </div>

      <!-- Right Pane: Rich Markdown Text Editor & Live Preview -->
      <div class="editor-panel">
        <div class="editor-toolbar-bar">
          <div class="editor-btn-group">
            <button class="editor-tool-btn" onclick="insertFormatting('# ')">H1</button>
            <button class="editor-tool-btn" onclick="insertFormatting('## ')">H2</button>
            <button class="editor-tool-btn" onclick="insertFormatting('**', '**')">B</button>
            <button class="editor-tool-btn" onclick="insertFormatting('*', '*')">I</button>
            <button class="editor-tool-btn" onclick="insertFormatting('\`', '\`')">Code</button>
            <button class="editor-tool-btn" onclick="insertFormatting('- ')">List</button>
          </div>
          <div class="editor-btn-group">
            <button class="editor-tool-btn ${activeViewMode === 'edit' ? 'active' : ''}" onclick="switchEditorView('edit')">Edit</button>
            <button class="editor-tool-btn ${activeViewMode === 'preview' ? 'active' : ''}" onclick="switchEditorView('preview')">Preview</button>
          </div>
        </div>

        <div class="docs-content-area" id="docs-content-area">
          <textarea class="docs-textarea-input" id="docs-editor" oninput="handleEditorInput()">${preloadedDocs[currentDocKey]}</textarea>
          <div class="docs-markdown-preview" id="docs-preview" style="display: none;"></div>
        </div>
      </div>
    </div>
  `;

  renderMarkdownPreview();
}

window.selectDocTemplate = function (key) {
  currentDocKey = key;
  const editor = document.getElementById('docs-editor');
  if (editor && preloadedDocs[key]) {
    editor.value = preloadedDocs[key];
    renderMarkdownPreview();
  }
  document.querySelectorAll('.docs-tree-item').forEach((item) => {
    item.classList.remove('active');
  });
  event.currentTarget.classList.add('active');
};

window.createNewDoc = function () {
  const editor = document.getElementById('docs-editor');
  if (editor) {
    editor.value = `# New Operational Document\n\nWrite document content here using Markdown...`;
    renderMarkdownPreview();
  }
};

window.switchEditorView = function (mode) {
  activeViewMode = mode;
  const editor = document.getElementById('docs-editor');
  const preview = document.getElementById('docs-preview');
  if (!editor || !preview) return;

  if (mode === 'preview') {
    editor.style.display = 'none';
    preview.style.display = 'block';
    renderMarkdownPreview();
  } else {
    editor.style.display = 'block';
    preview.style.display = 'none';
  }
};

window.handleEditorInput = function () {
  renderMarkdownPreview();
};

function renderMarkdownPreview() {
  const editor = document.getElementById('docs-editor');
  const preview = document.getElementById('docs-preview');
  if (!editor || !preview) return;

  let raw = editor.value;
  // Simple markdown renderer
  let html = raw
    .replace(/^# (.*$)/gim, '<h1>$1</h1>')
    .replace(/^## (.*$)/gim, '<h2>$1</h2>')
    .replace(/^### (.*$)/gim, '<h3>$1</h3>')
    .replace(/\*\*(.*)\*\*/gim, '<strong>$1</strong>')
    .replace(/\*(.*)\*/gim, '<em>$1</em>')
    .replace(/`(.*)`/gim, '<code>$1</code>')
    .replace(/\n/g, '<br/>');

  preview.innerHTML = html;
}

window.insertFormatting = function (prefix, suffix = '') {
  const editor = document.getElementById('docs-editor');
  if (!editor) return;
  const start = editor.selectionStart;
  const end = editor.selectionEnd;
  const text = editor.value;
  const sel = text.substring(start, end);
  const replacement = prefix + sel + suffix;
  editor.value = text.substring(0, start) + replacement + text.substring(end);
  editor.focus();
  renderMarkdownPreview();
};

window.saveCurrentDoc = function () {
  const editor = document.getElementById('docs-editor');
  if (editor) {
    localStorage.setItem(`opsflow_doc_${currentDocKey}`, editor.value);
    alert('Document saved to local storage successfully!');
  }
};

window.exportMarkdownDoc = function () {
  const editor = document.getElementById('docs-editor');
  if (editor) {
    navigator.clipboard.writeText(editor.value);
    alert('Markdown copied to clipboard!');
  }
};

// ──────────────────────────────────────────────
// L3 Helper Action Handlers
// ──────────────────────────────────────────────

window.analyzeStacktraceInput = function () {
  const input = document.getElementById('stacktrace-input');
  const resDiv = document.getElementById('stacktrace-result');
  if (!input || !resDiv) return;

  const val = input.value.trim();
  if (!val) {
    alert('Please paste a valid stacktrace or error log first.');
    return;
  }

  resDiv.style.display = 'block';
  resDiv.innerHTML = `
    <div style="padding: 16px; border: 1px solid var(--border-glass-strong); border-radius: var(--radius-md); background: rgba(99, 102, 241, 0.1);">
      <h3 style="font-size: 1rem; color: var(--accent-primary); margin-bottom: 8px;">🔍 AI Diagnostic Analysis</h3>
      <p style="font-size: 0.85rem; line-height: 1.5; color: var(--text-primary);">
        <strong>Root Cause Hypothesis:</strong> PostgreSQL connection pool exhaustion (98/100 active connections).<br/>
        <strong>Affected Component:</strong> <code>services/incident/internal/adapters/postgres/incident_repo.go:124</code><br/>
        <strong>Recommended L3 Remediation:</strong> Scale pgx pool max_conns to 200 or execute Redis read cache playbook.
      </p>
    </div>
  `;
};

window.runDBDiagnostics = function () {
  alert('DB Diagnostics Complete:\nActive Pool Connections: 98/100\nIdle: 2\nSlow Query: SELECT * FROM incidents FOR UPDATE\nStatus: WARNING_HIGH_LOAD');
};

window.generateIncidentRCA = function () {
  alert('RCA Report Generated:\n\n# Root Cause Analysis (RCA) — INC-2026-001\n\n- Incident: Payment Database Latency Timeout\n- Root Cause: Unindexed SELECT FOR UPDATE query\n- Preventative Action: Added composite index on outbox_events(status, created_at).');
};

// ──────────────────────────────────────────────
// Action Helpers
// ──────────────────────────────────────────────

window.openDeclareIncidentModal = function () {
  const overlay = document.getElementById('modal-overlay');
  const container = document.getElementById('modal-container');
  container.innerHTML = `
    <div class="modal-header">
      <h2>Declare New Incident</h2>
      <button class="sidebar-toggle" onclick="closeModal()">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label>Title</label>
        <input type="text" class="form-control" id="inc-title" placeholder="e.g. Payment Timeout" />
      </div>
      <div class="form-group">
        <label>Severity</label>
        <select class="form-control" id="inc-severity">
          <option value="CRITICAL">CRITICAL</option>
          <option value="HIGH">HIGH</option>
          <option value="MEDIUM">MEDIUM</option>
          <option value="LOW">LOW</option>
        </select>
      </div>
      <div class="form-group">
        <label>Description</label>
        <textarea class="form-control" id="inc-desc" rows="3" placeholder="Provide operational impact summary..."></textarea>
      </div>
      <div style="display:flex; justify-content:flex-end; gap:10px;">
        <button class="btn btn-secondary" onclick="closeModal()">Cancel</button>
        <button class="btn btn-primary" onclick="submitIncident()">Declare Incident</button>
      </div>
    </div>
  `;
  overlay.classList.add('open');
};

window.openRegisterServiceModal = function () {
  const overlay = document.getElementById('modal-overlay');
  const container = document.getElementById('modal-container');
  container.innerHTML = `
    <div class="modal-header">
      <h2>Register Service</h2>
      <button class="sidebar-toggle" onclick="closeModal()">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label>Service Name</label>
        <input type="text" class="form-control" id="svc-name" placeholder="e.g. auth-service" />
      </div>
      <div class="form-group">
        <label>Owner Team</label>
        <input type="text" class="form-control" id="svc-owner" placeholder="e.g. Core Engineering" />
      </div>
      <div style="display:flex; justify-content:flex-end; gap:10px;">
        <button class="btn btn-secondary" onclick="closeModal()">Cancel</button>
        <button class="btn btn-primary" onclick="closeModal(); alert('Service registered successfully!');">Register</button>
      </div>
    </div>
  `;
  overlay.classList.add('open');
};

window.closeModal = function () {
  document.getElementById('modal-overlay').classList.remove('open');
};

window.submitIncident = function () {
  alert('Incident declared and written to Outbox events!');
  closeModal();
  navigate('/incidents');
};

window.handleApproveAction = function () {
  alert('Action APPROVED! Operational tool mutation executed successfully.');
  const queue = document.getElementById('approval-queue-body');
  if (queue) {
    queue.innerHTML = `<div style="padding: 16px; color: var(--accent-emerald); font-weight: 500;">✓ Action APPROVED and executed safely. Audit trail recorded.</div>`;
  }
};

window.handleRejectAction = function () {
  alert('Action REJECTED by Operator.');
  const queue = document.getElementById('approval-queue-body');
  if (queue) {
    queue.innerHTML = `<div style="padding: 16px; color: var(--accent-rose); font-weight: 500;">✕ Action REJECTED. No operational mutation performed.</div>`;
  }
};

window.sendChatMessage = function () {
  const input = document.getElementById('chat-input');
  const messages = document.getElementById('chat-messages');
  if (!input || !messages || !input.value.trim()) return;

  const val = input.value.trim();
  messages.innerHTML += `<div class="chat-bubble user">${val}</div>`;
  input.value = '';

  setTimeout(() => {
    messages.innerHTML += `<div class="chat-bubble assistant">AI Response: Processed request '${val}'. Operational safety guardrails active.</div>`;
    messages.scrollTop = messages.scrollHeight;
  }, 500);
};

async function checkApiHealth() {
  const statusEl = document.getElementById('api-status');
  if (!statusEl) return;

  try {
    const res = await fetch(`${API_URL}/health`, { signal: AbortSignal.timeout(3000) });
    if (res.ok) {
      statusEl.querySelector('.status-dot').className = 'status-dot online';
      statusEl.querySelector('.status-text').textContent = 'API Online';
    } else {
      throw new Error('Not ok');
    }
  } catch {
    statusEl.querySelector('.status-dot').className = 'status-dot offline';
    statusEl.querySelector('.status-text').textContent = 'API Standalone (PWA)';
  }
}

document.addEventListener('DOMContentLoaded', createApp);
