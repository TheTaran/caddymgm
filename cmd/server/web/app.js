const els = {
  status: document.querySelector("#status"),
  pageTitle: document.querySelector("#page-title"),
  sectionTitle: document.querySelector("#section-title"),
  navItems: document.querySelectorAll(".nav-item[data-view]"),
  views: document.querySelectorAll(".view"),
  dashboardSiteList: document.querySelector("#dashboard-site-list"),
  siteList: document.querySelector("#site-list"),
  proxyGrid: document.querySelector("#proxy-hosts-grid"),
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
  logsEnabled: document.querySelector("#logs-enabled"),
  tlsEnabled: document.querySelector("#tls-enabled"),
  acmeIssuerRow: document.querySelector("#acme-issuer-row"),
  acmeIssuer: document.querySelector("#acme-issuer"),
  delete: document.querySelector("#delete"),
  totalSites: document.querySelector("#total-sites"),
  activeSites: document.querySelector("#active-sites"),
  proxySites: document.querySelector("#proxy-sites"),
  staticSites: document.querySelector("#static-sites"),
  logSiteFilter: document.querySelector("#log-site-filter"),
  logStreamLabel: document.querySelector("#log-stream-label"),
  logList: document.querySelector("#log-list"),
  settingsForm: document.querySelector("#settings-form"),
  settingsUsername: document.querySelector("#settings-username"),
  settingsPassword: document.querySelector("#settings-password"),
  settingsLogRetention: document.querySelector("#settings-log-retention"),
  settingsCaddyMode: document.querySelector("#settings-caddy-mode"),
  settingsCaddyAPIURL: document.querySelector("#settings-caddy-api-url"),
  certificateForm: document.querySelector("#certificate-form"),
  issuerNew: document.querySelector("#issuer-new"),
  issuerId: document.querySelector("#issuer-id"),
  issuerName: document.querySelector("#issuer-name"),
  issuerDirectory: document.querySelector("#issuer-directory"),
  issuerEmail: document.querySelector("#issuer-email"),
  issuerRootCA: document.querySelector("#issuer-root-ca"),
  issuerRootCAUpload: document.querySelector("#issuer-root-ca-upload"),
  issuerRootCAUploadButton: document.querySelector("#issuer-root-ca-upload-button"),
  issuerReset: document.querySelector("#issuer-reset"),
  issuerDelete: document.querySelector("#issuer-delete"),
  issuerList: document.querySelector("#issuer-list"),
  certificateList: document.querySelector("#certificate-list"),
  acmeDialog: document.querySelector("#acme-dialog"),
  acmeDialogClose: document.querySelector("#acme-dialog-close"),
  acmeDomain: document.querySelector("#acme-domain"),
  acmeAuthority: document.querySelector("#acme-authority"),
  acmeSteps: document.querySelector("#acme-steps"),
};

const viewTitles = {
  dashboard: ["Dashboard", "Web Hosts Overview"],
  "proxy-hosts": ["Web Hosts", "Website Configuration"],
  certificates: ["Certificates", "TLS Certificates"],
  logs: ["Logs", "Website Logs"],
  settings: ["Settings", "CaddyMGM Settings"],
};

let sites = [];
let settings = null;
let logPollTimer = null;

document.querySelector("#logout").addEventListener("click", logout);
document.querySelector("#new-site").addEventListener("click", () => {
  showView("proxy-hosts");
  editSite();
});
document.querySelector("#cancel").addEventListener("click", closeEditor);
els.form.addEventListener("submit", saveSite);
els.delete.addEventListener("click", deleteSite);
els.logSiteFilter.addEventListener("change", () => {
  loadLogs();
  syncLogPolling();
});
els.settingsForm.addEventListener("submit", saveSettings);
els.certificateForm.addEventListener("submit", saveIssuer);
els.issuerNew.addEventListener("click", () => editIssuer({}));
els.issuerReset.addEventListener("click", closeIssuerForm);
els.issuerDelete.addEventListener("click", deleteIssuer);
els.issuerRootCAUploadButton.addEventListener("click", uploadRootCA);
els.tlsEnabled.addEventListener("change", syncTLSMode);
els.acmeDialogClose.addEventListener("click", () => els.acmeDialog.close());
els.navItems.forEach((item) => item.addEventListener("click", () => showView(item.dataset.view)));
document.querySelectorAll("input[name='mode']").forEach((input) => {
  input.addEventListener("change", syncMode);
});

init();

async function init() {
  await Promise.all([loadSites(), loadSettings()]);
  renderLogs([]);
}

function showView(view) {
  els.navItems.forEach((item) => item.classList.toggle("active", item.dataset.view === view));
  els.views.forEach((panel) => panel.classList.toggle("active", panel.id === `view-${view}`));
  const [title, section] = viewTitles[view] || viewTitles.dashboard;
  els.pageTitle.textContent = title;
  els.sectionTitle.textContent = section;

  if (view === "proxy-hosts") closeEditor();
  if (view === "logs") {
    loadLogs();
    syncLogPolling();
  } else {
    stopLogPolling();
  }
  if (view === "settings") loadSettings();
  if (view === "certificates") renderCertificatesView();
}

async function loadSites() {
  setStatus("Loading websites...");
  try {
    const data = await request("/api/sites");
    sites = data.sites || [];
    renderMetrics();
    renderHostLists();
    renderLogFilter();
    renderCertificatesView();
    setStatus(`${sites.length} host${sites.length === 1 ? "" : "s"} managed`);
  } catch (err) {
    setStatus(err.message);
  }
}

async function loadSettings() {
  try {
    settings = await request("/api/settings");
    els.settingsUsername.value = settings.username || "admin";
    els.settingsPassword.value = "";
    els.settingsLogRetention.value = settings.logRetention || 100;
    els.settingsCaddyMode.value = settings.caddyMode || "file";
    els.settingsCaddyAPIURL.value = settings.caddyApiUrl || "";
    renderCertificatesView();
    renderIssuerOptions();
  } catch (err) {
    setStatus(err.message);
  }
}

async function saveSettings(event) {
  event.preventDefault();
  const payload = {
    appName: settings?.appName || "CaddyMGM",
    authEnabled: settings?.authEnabled ?? true,
    username: els.settingsUsername.value,
    password: els.settingsPassword.value,
    logRetention: Number(els.settingsLogRetention.value || 100),
    acmeIssuers: settings?.acmeIssuers || [],
  };
  try {
    settings = await request("/api/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    els.settingsPassword.value = "";
    setStatus("Settings saved");
    renderCertificatesView();
    renderIssuerOptions();
  } catch (err) {
    setStatus(err.message);
  }
}

async function saveIssuer(event) {
  event.preventDefault();
  const issuer = {
    id: els.issuerId.value,
    name: els.issuerName.value,
    directoryUrl: els.issuerDirectory.value,
    email: els.issuerEmail.value,
    rootCaFile: els.issuerRootCA.value,
  };
  const issuers = [...(settings?.acmeIssuers || [])];
  const index = issuers.findIndex((item) => item.id === issuer.id && issuer.id);
  if (index >= 0) {
    issuers[index] = issuer;
  } else {
    issuers.push(issuer);
  }
  await saveIssuers(issuers, "Authority saved");
  closeIssuerForm();
}

async function deleteIssuer() {
  const id = els.issuerId.value;
  const issuer = (settings?.acmeIssuers || []).find((item) => item.id === id);
  if (issuer?.builtIn) {
    setStatus("Built-in authorities cannot be deleted");
    return;
  }
  if (!id || !confirm("Delete this ACME authority?")) return;
  const issuers = (settings?.acmeIssuers || []).filter((issuer) => issuer.id !== id);
  await saveIssuers(issuers, "Authority deleted");
  closeIssuerForm();
}

async function uploadRootCA() {
  const file = els.issuerRootCAUpload.files?.[0];
  if (!file) {
    setStatus("Select a Root CA file first");
    return;
  }
  const body = new FormData();
  body.append("certificate", file);
  try {
    const data = await uploadRequest("/api/certificates/root-ca", body);
    els.issuerRootCA.value = data.rootCaFile || "";
    els.issuerRootCAUpload.value = "";
    setStatus("Root CA uploaded");
  } catch (err) {
    setStatus(err.message);
  }
}

async function saveIssuers(issuers, message) {
  const payload = {
    appName: settings?.appName || "CaddyMGM",
    authEnabled: settings?.authEnabled ?? true,
    username: settings?.username || els.settingsUsername.value || "admin",
    password: "",
    logRetention: settings?.logRetention || Number(els.settingsLogRetention.value || 100),
    acmeIssuers: issuers,
  };
  try {
    settings = await request("/api/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    renderCertificatesView();
    renderIssuerOptions();
    setStatus(message);
  } catch (err) {
    setStatus(err.message);
  }
}

function renderIssuers() {
  els.issuerList.innerHTML = "";
  const issuers = settings?.acmeIssuers || [];
  if (!issuers.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No ACME authorities configured.";
    els.issuerList.append(empty);
    return;
  }
  for (const issuer of issuers) {
    const row = document.createElement("div");
    row.className = "issuer-row";
    row.innerHTML = `
      <strong></strong>
      <span></span>
      <span class="badge"></span>
      <button class="secondary" type="button">Edit</button>
    `;
    row.children[0].textContent = issuer.name;
    row.children[1].textContent = issuer.directoryUrl;
    row.children[2].textContent = issuer.builtIn ? "Built-in" : "Custom";
    row.children[3].addEventListener("click", () => editIssuer(issuer));
    els.issuerList.append(row);
  }
}

function renderCertificatesView() {
  renderIssuers();
  renderIssuedCertificates();
}

function renderIssuedCertificates() {
  els.certificateList.innerHTML = "";
  const tlsSites = sites.filter((site) => site.tlsMode && site.tlsMode !== "off");
  if (!tlsSites.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No issued certificates known yet.";
    els.certificateList.append(empty);
    return;
  }
  for (const site of tlsSites) {
    const row = document.createElement("div");
    row.className = "certificate-row";
    row.innerHTML = `
      <strong></strong>
      <span></span>
      <span></span>
      <span class="badge"></span>
    `;
    row.children[0].textContent = site.address;
    row.children[1].textContent = certificateIssuerName(site);
    row.children[2].textContent = formatCertificateExpiry(site.certificateExpiresAt);
    row.children[3].textContent = site.enabled ? "Active" : "Disabled";
    row.children[3].classList.toggle("off", !site.enabled);
    els.certificateList.append(row);
  }
}

function formatCertificateExpiry(value) {
  if (!value) return "Unknown";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Unknown";
  return date.toLocaleDateString();
}

function certificateIssuerName(site) {
  const issuer = (settings?.acmeIssuers || []).find((item) => item.id === site.acmeIssuerId);
  return issuer?.name || "ACME Authority";
}

function renderIssuerOptions() {
  const current = els.acmeIssuer.value;
  els.acmeIssuer.innerHTML = "";
  for (const issuer of settings?.acmeIssuers || []) {
    const option = document.createElement("option");
    option.value = issuer.id;
    option.textContent = issuer.name;
    els.acmeIssuer.append(option);
  }
  els.acmeIssuer.value = current;
}

function editIssuer(issuer = null) {
  els.certificateForm.hidden = false;
  els.issuerId.value = issuer?.id || "";
  els.issuerName.value = issuer?.name || "";
  els.issuerDirectory.value = issuer?.directoryUrl || "";
  els.issuerEmail.value = issuer?.email || "";
  els.issuerRootCA.value = issuer?.rootCaFile || "";
  els.issuerRootCAUpload.value = "";
  els.issuerDelete.hidden = !issuer?.id || issuer.builtIn;
  els.issuerName.readOnly = !!issuer?.builtIn;
  els.issuerDirectory.readOnly = !!issuer?.builtIn;
  els.issuerRootCA.readOnly = !!issuer?.builtIn;
  els.issuerName.focus({ preventScroll: true });
}

function closeIssuerForm() {
  els.certificateForm.hidden = true;
  els.certificateForm.reset();
  els.issuerId.value = "";
  els.issuerName.readOnly = false;
  els.issuerDirectory.readOnly = false;
  els.issuerRootCA.readOnly = false;
  els.issuerRootCAUpload.value = "";
  els.issuerDelete.hidden = true;
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
    empty.textContent = "No web hosts configured yet.";
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
    row.children[3].textContent = site.enabled ? "Active" : "Inactive";
    row.children[3].classList.toggle("off", !site.enabled);
    row.children[4].textContent = editable ? "Edit" : "Open";
    row.children[4].addEventListener("click", () => {
      showView("proxy-hosts");
      editSite(site);
    });
    container.append(row);
  }
}

function renderLogFilter() {
  const current = els.logSiteFilter.value;
  els.logSiteFilter.innerHTML = `<option value="">Select Web Host</option>`;
  for (const site of sites) {
    const option = document.createElement("option");
    option.value = site.id;
    option.textContent = site.address;
    els.logSiteFilter.append(option);
  }
  els.logSiteFilter.value = current;
  syncLogPolling();
}

async function loadLogs() {
  const siteID = els.logSiteFilter.value;
  const selectedSite = sites.find((site) => site.id === siteID);
  els.logStreamLabel.hidden = !selectedSite;
  els.logStreamLabel.querySelector("strong").textContent = selectedSite?.address || "";
  if (!siteID) {
    renderLogs([]);
    return;
  }
  const query = `?siteId=${encodeURIComponent(siteID)}`;
  try {
    const data = await request(`/api/logs${query}`);
    renderLogs(data.logs || []);
  } catch (err) {
    setStatus(err.message);
  }
}

function renderLogs(logs) {
  els.logList.innerHTML = "";
  if (!els.logSiteFilter.value) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "Select a web host to view realtime logs.";
    els.logList.append(empty);
    return;
  }
  if (!logs.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No log lines for this web host yet.";
    els.logList.append(empty);
    return;
  }

  for (const entry of logs) {
    const row = document.createElement("div");
    row.className = "log-row";
    row.innerHTML = `
      <time></time>
      <span class="log-method"></span>
      <span class="log-path"></span>
      <span class="log-status badge"></span>
    `;
    row.children[0].textContent = new Date(entry.time).toLocaleTimeString();
    row.children[0].title = new Date(entry.time).toLocaleString();
    row.children[1].textContent = entry.method || "-";
    row.children[2].textContent = entry.path || entry.message || "-";
    row.children[2].title = entry.path || entry.message || "";
    row.children[3].textContent = entry.status || "-";
    row.children[3].classList.toggle("off", !String(entry.status || "").startsWith("2"));
    els.logList.append(row);
  }
}

function syncLogPolling() {
  stopLogPolling();
  const logsViewActive = document.querySelector("#view-logs").classList.contains("active");
  if (!logsViewActive || !els.logSiteFilter.value) return;
  logPollTimer = window.setInterval(loadLogs, 2500);
}

function stopLogPolling() {
  if (!logPollTimer) return;
  window.clearInterval(logPollTimer);
  logPollTimer = null;
}

function editSite(site = null) {
  els.editor.hidden = false;
  els.proxyGrid.classList.add("editor-open");
  els.form.reset();
  els.id.value = site?.id || "";
  els.formTitle.textContent = site ? "Edit Website" : "Create Website";
  els.address.value = site?.address || "";
  els.upstream.value = site?.upstream || "";
  els.root.value = site?.root || "";
  els.extra.value = site?.extraDirectives || "";
  els.enabled.checked = site?.enabled ?? true;
  els.logsEnabled.checked = site?.logsEnabled ?? true;
  els.tlsEnabled.checked = !!site && site?.tlsMode !== "off";
  renderIssuerOptions();
  els.acmeIssuer.value = site?.acmeIssuerId || "";
  const mode = site?.mode || "proxy";
  document.querySelector(`input[name='mode'][value='${mode}']`).checked = true;
  els.delete.hidden = !site;
  syncMode();
  syncTLSMode();
  els.editor.scrollIntoView({ behavior: "smooth", block: "start" });
  els.address.focus({ preventScroll: true });
}

function closeEditor() {
  els.editor.hidden = true;
  els.proxyGrid.classList.remove("editor-open");
}

function syncMode() {
  const mode = getMode();
  setFieldVisible(els.upstreamRow, mode === "proxy");
  setFieldVisible(els.rootRow, mode === "static");
  els.upstream.required = mode === "proxy";
  els.root.required = mode === "static";
  els.upstream.disabled = mode !== "proxy";
  els.root.disabled = mode !== "static";
  if (mode === "proxy") {
    els.root.value = "";
  } else {
    els.upstream.value = "";
  }
}

function setFieldVisible(row, visible) {
  row.hidden = !visible;
  row.classList.toggle("field-hidden", !visible);
  row.setAttribute("aria-hidden", visible ? "false" : "true");
}

function syncTLSMode() {
  const enabled = els.tlsEnabled.checked;
  els.acmeIssuerRow.hidden = !enabled;
  els.acmeIssuer.required = enabled;
  els.acmeIssuer.disabled = !enabled;
  if (!enabled) {
    els.acmeIssuer.value = "";
  } else if (!els.acmeIssuer.value && els.acmeIssuer.options.length > 0) {
    els.acmeIssuer.value = els.acmeIssuer.options[0].value;
  }
}

async function saveSite(event) {
  event.preventDefault();
  const payload = {
    address: els.address.value,
    mode: getMode(),
    upstream: els.upstream.value,
    root: els.root.value,
    extraDirectives: els.extra.value,
    logsEnabled: els.logsEnabled.checked,
    tlsMode: els.tlsEnabled.checked ? "acme" : "off",
    acmeIssuerId: els.tlsEnabled.checked ? els.acmeIssuer.value : "",
    enabled: els.enabled.checked,
  };
  const id = els.id.value;
  try {
    const saved = await request(id ? `/api/sites/${id}` : "/api/sites", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload),
    });
    closeEditor();
    await loadSites();
    await loadLogs();
    if (payload.tlsMode === "acme") {
      showACMEStatus(saved || payload);
    }
  } catch (err) {
    setStatus(err.message);
  }
}

function showACMEStatus(site) {
  const issuer = (settings?.acmeIssuers || []).find((item) => item.id === site.acmeIssuerId);
  els.acmeDomain.textContent = site.address;
  els.acmeAuthority.textContent = issuer ? issuer.name : "Selected ACME Authority";
  els.acmeSteps.innerHTML = "";
  [
    `Saved web host ${site.address}.`,
    "Generated Caddy TLS configuration with the selected ACME authority.",
    "Loaded the updated Caddyfile through the Caddy Admin API.",
    "Caddy now performs the ACME order and certificate validation in the background.",
  ].forEach((step) => {
    const item = document.createElement("li");
    item.textContent = step;
    els.acmeSteps.append(item);
  });
  els.acmeDialog.showModal();
}

async function deleteSite() {
  const id = els.id.value;
  if (!id || !confirm("Delete this website?")) return;
  try {
    await request(`/api/sites/${id}`, { method: "DELETE" });
    closeEditor();
    await loadSites();
    await loadLogs();
  } catch (err) {
    setStatus(err.message);
  }
}

async function request(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (response.status === 401) {
    window.location.assign("/login.html");
    throw new Error("authentication required");
  }
  if (response.status === 204) return null;
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || "Request failed");
  return data;
}

async function uploadRequest(url, body) {
  const response = await fetch(url, { method: "POST", body });
  if (response.status === 401) {
    window.location.assign("/login.html");
    throw new Error("authentication required");
  }
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || "Upload failed");
  return data;
}

async function logout() {
  await fetch("/api/auth/logout", { method: "POST" });
  window.location.assign("/login.html");
}

function getMode() {
  return document.querySelector("input[name='mode']:checked").value;
}

function setStatus(message) {
  const text = String(message || "");
  els.status.textContent = text.length > 96 ? `${text.slice(0, 93)}...` : text;
  els.status.title = text;
}
