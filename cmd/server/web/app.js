const els = {
  status: document.querySelector("#status"),
  pageTitle: document.querySelector("#page-title"),
  sectionTitle: document.querySelector("#section-title"),
  navItems: document.querySelectorAll(".nav-item[data-view]"),
  views: document.querySelectorAll(".view"),
  dashboardSiteList: document.querySelector("#dashboard-site-list"),
  siteList: document.querySelector("#site-list"),
  inlineConfig: document.querySelector("#inline-config"),
  editor: document.querySelector("#editor"),
  form: document.querySelector("#site-form"),
  formTitle: document.querySelector("#form-title"),
  id: document.querySelector("#site-id"),
  address: document.querySelector("#address"),
  upstream: document.querySelector("#upstream"),
  root: document.querySelector("#root"),
  upstreamRow: document.querySelector("#upstream-row"),
  rootRow: document.querySelector("#root-row"),
  extra: document.querySelector("#extra"),
  enabled: document.querySelector("#enabled"),
  delete: document.querySelector("#delete"),
  configDialog: document.querySelector("#config-dialog"),
  configContent: document.querySelector("#config-content"),
  totalSites: document.querySelector("#total-sites"),
  activeSites: document.querySelector("#active-sites"),
  proxySites: document.querySelector("#proxy-sites"),
  staticSites: document.querySelector("#static-sites"),
  logSiteFilter: document.querySelector("#log-site-filter"),
  logList: document.querySelector("#log-list"),
  settingsForm: document.querySelector("#settings-form"),
  settingsAppName: document.querySelector("#settings-app-name"),
  settingsAuthEnabled: document.querySelector("#settings-auth-enabled"),
  settingsUsername: document.querySelector("#settings-username"),
  settingsPassword: document.querySelector("#settings-password"),
  settingsLogRetention: document.querySelector("#settings-log-retention"),
  settingsConfigPath: document.querySelector("#settings-config-path"),
};

const viewTitles = {
  dashboard: ["Dashboard", "Proxy Hosts Übersicht"],
  "proxy-hosts": ["Proxy Hosts", "Konfiguration der einzelnen Webseiten"],
  certificates: ["Certificates", "TLS-Zertifikate"],
  logs: ["Logs", "Logs der einzelnen Webseiten"],
  settings: ["Settings", "CaddyMGM Einstellungen"],
};

let sites = [];
let settings = null;

document.querySelector("#refresh").addEventListener("click", refreshCurrentView);
document.querySelector("#new-site").addEventListener("click", () => {
  showView("proxy-hosts");
  editSite();
});
document.querySelector("#cancel").addEventListener("click", () => (els.editor.hidden = true));
document.querySelector("#show-config").addEventListener("click", showConfig);
document.querySelector("#close-config").addEventListener("click", () => els.configDialog.close());
els.form.addEventListener("submit", saveSite);
els.delete.addEventListener("click", deleteSite);
els.logSiteFilter.addEventListener("change", loadLogs);
els.settingsForm.addEventListener("submit", saveSettings);
els.navItems.forEach((item) => item.addEventListener("click", () => showView(item.dataset.view)));
document.querySelectorAll("input[name='mode']").forEach((input) => {
  input.addEventListener("change", syncMode);
});

init();

async function init() {
  await Promise.all([loadSites(), loadSettings()]);
  await loadConfigPreview();
  await loadLogs();
}

async function refreshCurrentView() {
  await loadSites();
  const current = document.querySelector(".view.active")?.id.replace("view-", "");
  if (current === "proxy-hosts") await loadConfigPreview();
  if (current === "logs") await loadLogs();
  if (current === "settings") await loadSettings();
}

function showView(view) {
  els.navItems.forEach((item) => item.classList.toggle("active", item.dataset.view === view));
  els.views.forEach((panel) => panel.classList.toggle("active", panel.id === `view-${view}`));
  const [title, section] = viewTitles[view] || viewTitles.dashboard;
  els.pageTitle.textContent = title;
  els.sectionTitle.textContent = section;

  if (view === "proxy-hosts") loadConfigPreview();
  if (view === "logs") loadLogs();
  if (view === "settings") loadSettings();
}

async function loadSites() {
  setStatus("Lade Websites...");
  try {
    const data = await request("/api/sites");
    sites = data.sites || [];
    renderMetrics();
    renderHostLists();
    renderLogFilter();
    setStatus(`${sites.length} Host${sites.length === 1 ? "" : "s"} verwaltet`);
  } catch (err) {
    setStatus(err.message);
  }
}

async function loadSettings() {
  try {
    settings = await request("/api/settings");
    els.settingsAppName.value = settings.appName || "CaddyMGM";
    els.settingsAuthEnabled.checked = Boolean(settings.authEnabled);
    els.settingsUsername.value = settings.username || "admin";
    els.settingsPassword.value = "";
    els.settingsLogRetention.value = settings.logRetention || 100;
    els.settingsConfigPath.textContent = settings.configPath || "/config/Caddyfile";
  } catch (err) {
    setStatus(err.message);
  }
}

async function saveSettings(event) {
  event.preventDefault();
  const payload = {
    appName: els.settingsAppName.value,
    authEnabled: els.settingsAuthEnabled.checked,
    username: els.settingsUsername.value,
    password: els.settingsPassword.value,
    logRetention: Number(els.settingsLogRetention.value || 100),
  };
  try {
    settings = await request("/api/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    els.settingsPassword.value = "";
    setStatus("Settings gespeichert");
  } catch (err) {
    setStatus(err.message);
  }
}

function renderMetrics() {
  els.totalSites.textContent = sites.length;
  els.activeSites.textContent = sites.filter((site) => site.enabled).length;
  els.proxySites.textContent = sites.filter((site) => site.mode === "proxy").length;
  els.staticSites.textContent = sites.filter((site) => site.mode === "static").length;
}

function renderHostLists() {
  renderSiteList(els.dashboardSiteList, false);
  renderSiteList(els.siteList, true);
}

function renderSiteList(container, editable) {
  container.innerHTML = "";
  if (!sites.length) {
    const empty = document.createElement("div");
    empty.className = "site-row empty";
    empty.textContent = "Noch keine Proxy Hosts angelegt.";
    container.append(empty);
    return;
  }

  for (const site of sites) {
    const row = document.createElement("div");
    row.className = "site-row";
    row.innerHTML = `
      <strong></strong>
      <span></span>
      <span class="target"></span>
      <span class="badge"></span>
      <button class="secondary" type="button"></button>
    `;
    row.children[0].textContent = site.address;
    row.children[1].textContent = site.mode === "static" ? "Static" : "Proxy";
    row.children[2].textContent = site.mode === "static" ? site.root : site.upstream;
    row.children[3].textContent = site.enabled ? "Aktiv" : "Inaktiv";
    row.children[3].classList.toggle("off", !site.enabled);
    row.children[4].textContent = editable ? "Bearbeiten" : "Öffnen";
    row.children[4].addEventListener("click", () => {
      showView("proxy-hosts");
      editSite(site);
    });
    container.append(row);
  }
}

function renderLogFilter() {
  const current = els.logSiteFilter.value;
  els.logSiteFilter.innerHTML = `<option value="">Alle Hosts</option>`;
  for (const site of sites) {
    const option = document.createElement("option");
    option.value = site.id;
    option.textContent = site.address;
    els.logSiteFilter.append(option);
  }
  els.logSiteFilter.value = current;
}

async function loadLogs() {
  const query = els.logSiteFilter.value ? `?siteId=${encodeURIComponent(els.logSiteFilter.value)}` : "";
  try {
    const data = await request(`/api/logs${query}`);
    renderLogs(data.logs || []);
  } catch (err) {
    setStatus(err.message);
  }
}

function renderLogs(logs) {
  els.logList.innerHTML = "";
  if (!logs.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "Noch keine Logs vorhanden.";
    els.logList.append(empty);
    return;
  }

  for (const entry of logs) {
    const row = document.createElement("div");
    row.className = "log-row";
    row.innerHTML = `
      <time></time>
      <strong></strong>
      <span></span>
    `;
    row.children[0].textContent = new Date(entry.time).toLocaleString();
    row.children[1].textContent = entry.site || "CaddyMGM";
    row.children[2].textContent = `${entry.action}: ${entry.message}`;
    els.logList.append(row);
  }
}

function editSite(site = null) {
  els.editor.hidden = false;
  els.form.reset();
  els.id.value = site?.id || "";
  els.formTitle.textContent = site ? "Website bearbeiten" : "Website anlegen";
  els.address.value = site?.address || "";
  els.upstream.value = site?.upstream || "";
  els.root.value = site?.root || "";
  els.extra.value = site?.extraDirectives || "";
  els.enabled.checked = site?.enabled ?? true;
  const mode = site?.mode || "proxy";
  document.querySelector(`input[name='mode'][value='${mode}']`).checked = true;
  els.delete.hidden = !site;
  syncMode();
  els.address.focus();
}

function syncMode() {
  const mode = getMode();
  els.upstreamRow.hidden = mode !== "proxy";
  els.rootRow.hidden = mode !== "static";
  els.upstream.required = mode === "proxy";
  els.root.required = mode === "static";
}

async function saveSite(event) {
  event.preventDefault();
  const payload = {
    address: els.address.value,
    mode: getMode(),
    upstream: els.upstream.value,
    root: els.root.value,
    extraDirectives: els.extra.value,
    enabled: els.enabled.checked,
  };
  const id = els.id.value;
  try {
    await request(id ? `/api/sites/${id}` : "/api/sites", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload),
    });
    els.editor.hidden = true;
    await loadSites();
    await loadConfigPreview();
    await loadLogs();
  } catch (err) {
    setStatus(err.message);
  }
}

async function deleteSite() {
  const id = els.id.value;
  if (!id || !confirm("Website wirklich löschen?")) return;
  try {
    await request(`/api/sites/${id}`, { method: "DELETE" });
    els.editor.hidden = true;
    await loadSites();
    await loadConfigPreview();
    await loadLogs();
  } catch (err) {
    setStatus(err.message);
  }
}

async function loadConfigPreview() {
  try {
    const response = await fetch("/api/config");
    if (!response.ok) throw new Error("Caddyfile konnte nicht geladen werden");
    els.inlineConfig.textContent = await response.text();
  } catch (err) {
    els.inlineConfig.textContent = err.message;
  }
}

async function showConfig() {
  const response = await fetch("/api/config");
  els.configContent.textContent = await response.text();
  els.configDialog.showModal();
}

async function request(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (response.status === 204) return null;
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || "Request failed");
  return data;
}

function getMode() {
  return document.querySelector("input[name='mode']:checked").value;
}

function setStatus(message) {
  els.status.textContent = message;
}
