const els = {
  status: document.querySelector("#status"),
  pageTitle: document.querySelector("#page-title"),
  pageTitleIcon: document.querySelector("#page-title-icon"),
  sectionTitle: document.querySelector("#section-title"),
  sectionTitleIcon: document.querySelector("#section-title-icon"),
  metrics: document.querySelector("#metrics"),
  profileAvatar: document.querySelector("#profile-avatar"),
  profileUsername: document.querySelector("#profile-username"),
  profileProvider: document.querySelector("#profile-provider"),
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
  comment: document.querySelector("#comment"),
  upstream: document.querySelector("#upstream"),
  skipTlsVerify: document.querySelector("#skip-tls-verify"),
  root: document.querySelector("#root"),
  upstreamRow: document.querySelector("#upstream-row"),
  skipTlsVerifyRow: document.querySelector("#skip-tls-verify-row"),
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
  serviceLogList: document.querySelector("#service-log-list"),
  serviceLogToggle: document.querySelector("#service-log-toggle"),
  siteLogToggle: document.querySelector("#site-log-toggle"),
  settingsForm: document.querySelector("#settings-form"),
  settingsUsername: document.querySelector("#settings-username"),
  settingsPassword: document.querySelector("#settings-password"),
  settingsPasswordConfirm: document.querySelector("#settings-password-confirm"),
  settingsOIDCEnabled: document.querySelector("#settings-oidc-enabled"),
  settingsOIDCIssuer: document.querySelector("#settings-oidc-issuer"),
  settingsOIDCClientID: document.querySelector("#settings-oidc-client-id"),
  settingsOIDCClientSecret: document.querySelector("#settings-oidc-client-secret"),
  settingsOIDCRedirect: document.querySelector("#settings-oidc-redirect"),
  settingsOIDCScopes: document.querySelector("#settings-oidc-scopes"),
  settingsWebHost: document.querySelector("#settings-web-host"),
  settingsWebTLSEnabled: document.querySelector("#settings-web-tls-enabled"),
  settingsWebACMERow: document.querySelector("#settings-web-acme-row"),
  settingsWebACME: document.querySelector("#settings-web-acme"),
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
  acmeLogList: document.querySelector("#acme-log-list"),
};

const viewTitles = {
  dashboard: ["Dashboard", "Web Hosts Overview"],
  "proxy-hosts": ["Web Hosts", "Website Configuration"],
  certificates: ["Certificates", "TLS Certificates"],
  logs: ["Logs", "Website Logs"],
  settings: ["Settings", "CaddyMGM Settings"],
};

const viewIcons = {
  dashboard: `
    <svg viewBox="0 0 24 24" fill="none">
      <path d="M4 13.5 12 5l8 8.5" />
      <path d="M6.5 11.5V20h11V11.5" />
      <path d="M10 20v-5h4v5" />
    </svg>
  `,
  "proxy-hosts": `
    <svg viewBox="0 0 24 24" fill="none">
      <rect x="4" y="6" width="16" height="5" rx="2.5" />
      <rect x="4" y="13" width="16" height="5" rx="2.5" />
      <path d="M8 8.5h.01M8 15.5h.01" />
      <path d="M13 8.5h4M13 15.5h4" />
    </svg>
  `,
  certificates: `
    <svg viewBox="0 0 24 24" fill="none">
      <path d="M12 3 6 5.5v5.5c0 4 2.56 7.2 6 10 3.44-2.8 6-6 6-10V5.5L12 3Z" />
      <path d="m9.5 12 1.8 1.8 3.7-4" />
    </svg>
  `,
  logs: `
    <svg viewBox="0 0 24 24" fill="none">
      <path d="M5 6.5h14" />
      <path d="M5 12h14" />
      <path d="M5 17.5h9" />
      <path d="M15 15.5 17.5 18 20 15.5" />
    </svg>
  `,
  settings: `
    <svg viewBox="0 0 24 24" fill="none">
      <path d="M12 8.5a3.5 3.5 0 1 0 0 7 3.5 3.5 0 0 0 0-7Z" />
      <path d="M19 12a7 7 0 0 0-.08-1l2-1.54-2-3.46-2.4.83a7.77 7.77 0 0 0-1.74-1L14.5 3h-5l-.28 2.83a7.77 7.77 0 0 0-1.74 1l-2.4-.83-2 3.46 2 1.54a7 7 0 0 0 0 2l-2 1.54 2 3.46 2.4-.83a7.77 7.77 0 0 0 1.74 1L9.5 21h5l.28-2.83a7.77 7.77 0 0 0 1.74-1l2.4.83 2-3.46-2-1.54c.05-.33.08-.67.08-1Z" />
    </svg>
  `,
};

let sites = [];
const hostSort = {
  dashboard: { key: "address", direction: "asc" },
  "web-hosts": { key: "address", direction: "asc" },
};
const hostCollator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });
let settings = null;
let logPollTimer = null;
let acmeDialogTimer = null;
let acmeDialogContext = null;
let serviceLogsExpanded = false;
let siteLogsExpanded = false;
let editingSite = null;
const LOG_PREVIEW_LIMIT = 10;

document.querySelector("#logout").addEventListener("click", logout);
document.querySelector("#new-site").addEventListener("click", () => {
  showView("proxy-hosts");
  editSite();
});
document.querySelector("#cancel").addEventListener("click", closeEditor);
els.form.addEventListener("submit", saveSite);
els.delete.addEventListener("click", deleteSite);
els.logSiteFilter.addEventListener("change", () => {
  siteLogsExpanded = false;
  loadLogs();
  syncLogPolling();
});
els.serviceLogToggle.addEventListener("click", toggleServiceLogsExpanded);
els.siteLogToggle.addEventListener("click", toggleSiteLogsExpanded);
els.settingsForm.addEventListener("submit", saveSettings);
els.settingsPassword.addEventListener("input", syncSettingsPasswordConfirmation);
els.settingsPasswordConfirm.addEventListener("input", syncSettingsPasswordConfirmation);
els.settingsWebTLSEnabled.addEventListener("change", syncSettingsWebTLS);
els.certificateForm.addEventListener("submit", saveIssuer);
els.issuerNew.addEventListener("click", () => editIssuer({}));
els.issuerReset.addEventListener("click", closeIssuerForm);
els.issuerDelete.addEventListener("click", deleteIssuer);
els.issuerRootCAUploadButton.addEventListener("click", uploadRootCA);
els.tlsEnabled.addEventListener("change", syncTLSMode);
els.upstream.addEventListener("input", syncUpstreamTLS);
els.acmeDialogClose.addEventListener("click", closeACMEStatus);
els.acmeDialog.addEventListener("close", stopACMEStatusPolling);
els.acmeDialog.addEventListener("cancel", stopACMEStatusPolling);
els.navItems.forEach((item) => item.addEventListener("click", () => showView(item.dataset.view)));
document.querySelectorAll("[data-sort-table]").forEach((header) => {
  header.addEventListener("click", (event) => {
    const button = event.target.closest(".sort-button");
    if (!button) return;
    const table = header.dataset.sortTable;
    const sort = hostSort[table];
    if (sort.key === button.dataset.sort) {
      sort.direction = sort.direction === "asc" ? "desc" : "asc";
    } else {
      sort.key = button.dataset.sort;
      sort.direction = "asc";
    }
    renderHostLists();
  });
});
document.querySelectorAll("input[name='mode']").forEach((input) => {
  input.addEventListener("change", syncMode);
});

init();

async function init() {
  await Promise.all([loadSites(), loadSettings(), loadProfile()]);
  renderLogs([]);
  renderServiceLogs([]);
}

async function loadProfile() {
  try {
    const me = await request("/api/auth/me");
    const username = String(me.username || "admin");
    const provider = String(me.provider || "local");
    els.profileUsername.textContent = username;
    els.profileProvider.textContent = provider === "oidc" ? "OIDC Login" : "Local Login";
    els.profileAvatar.textContent = initialsForUser(username);
  } catch (_err) {
    els.profileUsername.textContent = "admin@local";
    els.profileProvider.textContent = "Caddy Management";
    els.profileAvatar.textContent = "CM";
  }
}

function showView(view) {
  els.navItems.forEach((item) => item.classList.toggle("active", item.dataset.view === view));
  els.views.forEach((panel) => panel.classList.toggle("active", panel.id === `view-${view}`));
  els.metrics.hidden = view !== "dashboard";
  const [title, section] = viewTitles[view] || viewTitles.dashboard;
  els.pageTitle.textContent = title;
  els.sectionTitle.textContent = section;
  const icon = viewIcons[view] || viewIcons.dashboard;
  els.pageTitleIcon.innerHTML = icon;
  els.sectionTitleIcon.innerHTML = icon;

  if (view === "proxy-hosts") closeEditor();
  if (view === "logs") {
    loadServiceLogs();
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
    els.settingsPasswordConfirm.value = "";
    els.settingsOIDCEnabled.checked = !!settings.oidc?.enabled;
    els.settingsOIDCIssuer.value = settings.oidc?.issuerUrl || "";
    els.settingsOIDCClientID.value = settings.oidc?.clientId || "";
    els.settingsOIDCClientSecret.value = "";
    els.settingsOIDCRedirect.value = settings.oidc?.redirectUrl || "";
    els.settingsOIDCScopes.value = settings.oidc?.scopes || "openid profile email";
    els.settingsWebHost.value = settings.webInterface?.host || "";
    els.settingsWebTLSEnabled.checked = !!settings.webInterface?.tlsEnabled;
    els.settingsLogRetention.value = settings.logRetention || 100;
    els.settingsCaddyMode.value = settings.caddyMode || "file";
    els.settingsCaddyAPIURL.value = settings.caddyApiUrl || "";
    renderCertificatesView();
    renderIssuerOptions();
    els.settingsWebACME.value = settings.webInterface?.acmeIssuerId || "";
    syncSettingsPasswordConfirmation();
    syncSettingsWebTLS();
  } catch (err) {
    setStatus(err.message);
  }
}

async function saveSettings(event) {
  event.preventDefault();
  if (!syncSettingsPasswordConfirmation()) {
    setStatus("The new passwords do not match");
    els.settingsPasswordConfirm.focus({ preventScroll: true });
    return;
  }
  const payload = {
    appName: settings?.appName || "CaddyMGM",
    authEnabled: settings?.authEnabled ?? true,
    username: els.settingsUsername.value,
    password: els.settingsPassword.value,
    oidc: {
      enabled: els.settingsOIDCEnabled.checked,
      issuerUrl: els.settingsOIDCIssuer.value,
      clientId: els.settingsOIDCClientID.value,
      clientSecret: els.settingsOIDCClientSecret.value,
      redirectUrl: els.settingsOIDCRedirect.value,
      scopes: els.settingsOIDCScopes.value,
    },
    webInterface: {
      host: els.settingsWebHost.value,
      tlsEnabled: els.settingsWebTLSEnabled.checked,
      acmeIssuerId: els.settingsWebTLSEnabled.checked ? els.settingsWebACME.value : "",
    },
    logRetention: Number(els.settingsLogRetention.value || 100),
    acmeIssuers: settings?.acmeIssuers || [],
  };
  try {
    settings = await request("/api/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    els.settingsPassword.value = "";
    els.settingsPasswordConfirm.value = "";
    syncSettingsPasswordConfirmation();
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
    oidc: settings?.oidc || {},
    webInterface: settings?.webInterface || {},
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
      <button class="secondary" type="button">Renew</button>
    `;
    row.children[0].textContent = site.address;
    row.children[1].textContent = certificateIssuerName(site);
    row.children[2].textContent = formatCertificateExpiry(site.certificateExpiresAt);
    row.children[3].textContent = site.enabled ? "Active" : "Disabled";
    row.children[3].classList.toggle("off", !site.enabled);
    row.children[4].disabled = !site.enabled;
    row.children[4].title = site.enabled ? `Force renew for ${site.address}` : "Enable the web host before forcing renewal";
    row.children[4].addEventListener("click", () => renewCertificate(site));
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
  const settingsCurrent = els.settingsWebACME.value;
  els.acmeIssuer.innerHTML = "";
  els.settingsWebACME.innerHTML = "";
  for (const issuer of settings?.acmeIssuers || []) {
    const option = document.createElement("option");
    option.value = issuer.id;
    option.textContent = issuer.name;
    els.acmeIssuer.append(option);
    els.settingsWebACME.append(option.cloneNode(true));
  }
  els.acmeIssuer.value = current;
  els.settingsWebACME.value = settingsCurrent;
  syncSettingsWebTLS();
}

function syncSettingsWebTLS() {
  const enabled = els.settingsWebTLSEnabled.checked;
  els.settingsWebACMERow.hidden = !enabled;
  els.settingsWebACME.required = enabled;
  els.settingsWebACME.disabled = !enabled;
  els.settingsWebHost.placeholder = "mgm.example.com";
  if (!enabled) {
    els.settingsWebACME.value = "";
  } else if (!els.settingsWebACME.value && els.settingsWebACME.options.length > 0) {
    els.settingsWebACME.value = els.settingsWebACME.options[0].value;
  }
}

function syncSettingsPasswordConfirmation() {
  const password = els.settingsPassword.value;
  const confirmation = els.settingsPasswordConfirm.value;
  const mismatch = password !== confirmation;
  const message = mismatch ? "The new passwords do not match" : "";
  els.settingsPasswordConfirm.setCustomValidity(message);
  if (confirmation) {
    els.settingsPasswordConfirm.reportValidity();
  }
  return !mismatch;
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
  updateHostSortHeaders();
  renderSiteList(els.dashboardSiteList, false, "dashboard");
  renderSiteList(els.siteList, true, "web-hosts");
}

function updateHostSortHeaders() {
  document.querySelectorAll("[data-sort-table]").forEach((header) => {
    const sort = hostSort[header.dataset.sortTable];
    header.querySelectorAll(".sort-button").forEach((button) => {
      const active = button.dataset.sort === sort.key;
      button.dataset.direction = active ? sort.direction : "";
      button.setAttribute("aria-label", `${button.textContent}: ${active ? (sort.direction === "asc" ? "sorted ascending" : "sorted descending") : "not sorted"}`);
      button.parentElement.setAttribute("aria-sort", active ? (sort.direction === "asc" ? "ascending" : "descending") : "none");
    });
  });
}

function hostSortValue(site, key) {
  switch (key) {
    case "protocol": return protocolForSite(site);
    case "mode": return site.mode === "static" ? "Static" : "Proxy";
    case "target": return site.mode === "static" ? site.root : site.upstream;
    case "upstreamTls": return site.mode === "proxy" ? (site.skipTlsVerify ? "Skipped" : "Verified") : "-";
    case "status": return site.enabled ? "Active" : "Inactive";
    case "comment": return site.comment || "";
    default: return site.address || "";
  }
}

function sortedSitesFor(table) {
  const sort = hostSort[table];
  if (!sort.key) return sites;
  const direction = sort.direction === "desc" ? -1 : 1;
  return sites.map((site, index) => ({ site, index })).sort((left, right) => {
    const result = hostCollator.compare(hostSortValue(left.site, sort.key), hostSortValue(right.site, sort.key));
    return result === 0 ? left.index - right.index : result * direction;
  }).map(({ site }) => site);
}

function renderSiteList(container, editable, table) {
  container.innerHTML = "";
  if (!sites.length) {
    const empty = document.createElement("div");
    empty.className = "site-row empty";
    empty.textContent = "No web hosts configured yet.";
    container.append(empty);
    return;
  }

  for (const site of sortedSitesFor(table)) {
    const row = document.createElement("div");
    row.className = `site-row ${editable ? "site-row-editable" : "site-row-dashboard"}`;
    row.innerHTML = editable
      ? `
          <strong></strong>
          <span class="badge"></span>
          <span></span>
          <span class="target"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="target"></span>
          <div class="site-actions"></div>
        `
      : `
          <strong></strong>
          <span class="badge"></span>
          <span></span>
          <span class="target"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="target"></span>
        `;
    row.children[0].textContent = site.address;
    const protocol = protocolForSite(site);
    row.children[1].textContent = protocol;
    row.children[1].classList.toggle("secure", protocol === "https");
    row.children[1].classList.toggle("warn", protocol === "http");
    row.children[2].textContent = site.mode === "static" ? "Static" : "Proxy";
    row.children[3].textContent = site.mode === "static" ? site.root : site.upstream;
    row.children[4].textContent = site.mode === "proxy" ? (site.skipTlsVerify ? "Skipped" : "Verified") : "-";
    row.children[4].classList.toggle("secure", site.mode === "proxy" && !site.skipTlsVerify);
    row.children[4].classList.toggle("warn", site.mode === "proxy" && !!site.skipTlsVerify);
    row.children[4].classList.toggle("off", site.mode !== "proxy");
    row.children[5].textContent = site.enabled ? "Active" : "Inactive";
    row.children[5].classList.toggle("off", !site.enabled);
    row.children[6].textContent = site.comment || "-";
    row.children[6].title = site.comment || "";
    if (editable) {
      const actions = row.children[7];
      actions.append(createSiteActionButton("Edit", "secondary", () => {
        showView("proxy-hosts");
        editSite(site);
      }));
      actions.append(
        createSiteActionButton(site.enabled ? "Deactivate" : "Activate", "secondary", () => toggleSiteEnabled(site)),
      );
    }
    container.append(row);
  }
}

function createSiteActionButton(label, className, onClick) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = className;
  button.textContent = label;
  button.addEventListener("click", onClick);
  return button;
}

async function toggleSiteEnabled(site) {
  if (!site?.id) return;
  const nextEnabled = !site.enabled;
  try {
    await request(`/api/sites/${site.id}`, {
      method: "PUT",
      body: JSON.stringify({
        address: site.address,
        comment: site.comment || "",
        mode: site.mode,
        upstream: site.upstream || "",
        skipTlsVerify: !!site.skipTlsVerify,
        root: site.root || "",
        extraDirectives: site.extraDirectives || "",
        logsEnabled: !!site.logsEnabled,
        tlsMode: site.tlsMode || "off",
        acmeIssuerId: site.tlsMode === "acme" ? site.acmeIssuerId || "" : "",
        enabled: nextEnabled,
      }),
    });
    setStatus(`Website ${nextEnabled ? "activated" : "deactivated"}`);
    await loadSites();
    await loadLogs();
  } catch (err) {
    setStatus(err.message);
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

async function loadServiceLogs() {
  try {
    const data = await request("/api/logs?source=caddy-service");
    renderServiceLogs(data.logs || [], data.available !== false);
    renderACMEActivityLogs(data.logs || [], data.available !== false);
  } catch (err) {
    setStatus(err.message);
  }
}

function renderLogs(logs) {
  els.logList.innerHTML = "";
  els.siteLogToggle.hidden = true;
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

  const visibleLogs = siteLogsExpanded ? logs : logs.slice(0, LOG_PREVIEW_LIMIT);
  for (const entry of visibleLogs) {
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
  syncLogToggle(els.siteLogToggle, logs.length, siteLogsExpanded);
}

function renderServiceLogs(logs, available = true) {
  els.serviceLogList.innerHTML = "";
  els.serviceLogToggle.hidden = true;
  if (!available) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No Caddy service log available in the current mode.";
    els.serviceLogList.append(empty);
    return;
  }
  if (!logs.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No Caddy service log lines yet.";
    els.serviceLogList.append(empty);
    return;
  }

  const visibleLogs = serviceLogsExpanded ? logs : logs.slice(0, LOG_PREVIEW_LIMIT);
  for (const entry of visibleLogs) {
    const row = document.createElement("div");
    row.className = "log-row service-log-row";
    row.innerHTML = `
      <time></time>
      <span class="log-method"></span>
      <span class="log-path"></span>
      <span class="log-status badge"></span>
    `;
    row.children[0].textContent = new Date(entry.time).toLocaleTimeString();
    row.children[0].title = new Date(entry.time).toLocaleString();
    row.children[1].textContent = entry.action || "service";
    row.children[2].textContent = entry.message || "-";
    row.children[2].title = entry.message || "";
    row.children[3].textContent = entry.status || "INFO";
    row.children[3].classList.toggle("off", String(entry.status || "").toUpperCase() === "ERROR");
    els.serviceLogList.append(row);
  }
  syncLogToggle(els.serviceLogToggle, logs.length, serviceLogsExpanded);
}

function syncLogToggle(button, total, expanded) {
  if (total <= LOG_PREVIEW_LIMIT) {
    button.hidden = true;
    return;
  }
  button.hidden = false;
  button.textContent = expanded ? "Show less" : `Show ${total - LOG_PREVIEW_LIMIT} more`;
}

function toggleServiceLogsExpanded() {
  serviceLogsExpanded = !serviceLogsExpanded;
  loadServiceLogs();
}

function toggleSiteLogsExpanded() {
  siteLogsExpanded = !siteLogsExpanded;
  loadLogs();
}

function syncLogPolling() {
  stopLogPolling();
  const logsViewActive = document.querySelector("#view-logs").classList.contains("active");
  if (!logsViewActive) return;
  logPollTimer = window.setInterval(() => {
    loadServiceLogs();
    if (els.logSiteFilter.value) {
      loadLogs();
    }
  }, 2500);
}

function stopLogPolling() {
  if (!logPollTimer) return;
  window.clearInterval(logPollTimer);
  logPollTimer = null;
}

function editSite(site = null) {
  editingSite = site ? { ...site } : null;
  els.editor.hidden = false;
  els.proxyGrid.classList.add("editor-open");
  els.form.reset();
  els.id.value = site?.id || "";
  els.formTitle.textContent = site ? "Edit Website" : "Create Website";
  els.address.value = site?.address || "";
  els.comment.value = site?.comment || "";
  els.upstream.value = site?.upstream || "";
  els.skipTlsVerify.checked = !!site?.skipTlsVerify;
  els.root.value = site?.root || "";
  els.extra.value = site?.extraDirectives || "";
  els.enabled.checked = site?.enabled ?? true;
  els.logsEnabled.checked = site?.logsEnabled ?? true;
  els.tlsEnabled.checked = !!site && site?.tlsMode !== "off";
  renderIssuerOptions();
  els.acmeIssuer.value = site?.acmeIssuerId || "";
  document.querySelectorAll("input[name='mode']").forEach((input) => {
    input.checked = false;
  });
  if (site?.mode) {
    const selected = document.querySelector(`input[name='mode'][value='${site.mode}']`);
    if (selected) selected.checked = true;
  }
  els.delete.hidden = !site;
  syncMode();
  syncTLSMode();
  els.editor.scrollIntoView({ behavior: "smooth", block: "start" });
  els.address.focus({ preventScroll: true });
}

function closeEditor() {
  editingSite = null;
  els.editor.hidden = true;
  els.proxyGrid.classList.remove("editor-open");
}

function syncMode() {
  const mode = getMode();
  setFieldVisible(els.upstreamRow, mode === "proxy");
  setFieldVisible(els.skipTlsVerifyRow, mode === "proxy");
  setFieldVisible(els.rootRow, mode === "static");
  els.upstream.required = mode === "proxy";
  els.root.required = mode === "static";
  els.upstream.disabled = mode !== "proxy";
  els.skipTlsVerify.disabled = mode !== "proxy";
  els.root.disabled = mode !== "static";
  if (mode === "proxy") {
    els.root.value = "";
  } else if (mode === "static") {
    els.upstream.value = "";
    els.skipTlsVerify.checked = false;
  } else {
    els.upstream.value = "";
    els.skipTlsVerify.checked = false;
    els.root.value = "";
  }
  syncUpstreamTLS();
}

function syncUpstreamTLS() {
  const usesPlainHTTP = /^http:\/\//i.test(els.upstream.value.trim());
  if (usesPlainHTTP) els.skipTlsVerify.checked = false;
  els.skipTlsVerify.disabled = getMode() !== "proxy" || usesPlainHTTP;
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
  const mode = getMode();
  if (!mode) {
    setStatus("Select a website type first");
    return;
  }
  const payload = {
    address: els.address.value,
    comment: els.comment.value,
    mode,
    upstream: els.upstream.value,
    skipTlsVerify: els.skipTlsVerify.checked,
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
    const shouldStartACMEFlow = shouldOpenACMEStatus(editingSite, payload, saved);
    closeEditor();
    await loadSites();
    await loadLogs();
    if (shouldStartACMEFlow) {
      showACMEStatus(saved || payload);
    }
  } catch (err) {
    setStatus(err.message);
  }
}

function shouldOpenACMEStatus(previousSite, payload, saved) {
  if (payload.tlsMode !== "acme") return false;
  if (saved?.certificateExpiresAt) return false;
  if (!previousSite) return true;

  const previousAddress = normalizeAddress(previousSite.address);
  const nextAddress = normalizeAddress(payload.address);
  const sameAddress = previousAddress === nextAddress;
  const sameIssuer = String(previousSite.acmeIssuerId || "") === String(payload.acmeIssuerId || "");
  const previousTLSActive = previousSite.tlsMode === "acme";
  const previousCertificateExists = !!previousSite.certificateExpiresAt;

  if (previousTLSActive && sameAddress && sameIssuer && previousCertificateExists) {
    return false;
  }
  return true;
}

function normalizeAddress(value) {
  return String(value || "")
    .trim()
    .replace(/^https?:\/\//i, "")
    .toLowerCase();
}

function showACMEStatus(site) {
  const issuer = (settings?.acmeIssuers || []).find((item) => item.id === site.acmeIssuerId);
  els.acmeDomain.textContent = site.address;
  els.acmeAuthority.textContent = issuer ? issuer.name : "Selected ACME Authority";
  acmeDialogContext = {
    domain: String(site.address || "").toLowerCase(),
    startedAt: Date.now() - 5000,
  };
  els.acmeDialog.showModal();
  renderACMEActivityLogs([], true);
  stopACMEStatusPolling();
  loadServiceLogs();
  acmeDialogTimer = window.setInterval(loadServiceLogs, 2000);
}

function closeACMEStatus() {
  stopACMEStatusPolling();
  acmeDialogContext = null;
  els.acmeDialog.close();
}

function stopACMEStatusPolling() {
  if (!acmeDialogTimer) return;
  window.clearInterval(acmeDialogTimer);
  acmeDialogTimer = null;
}

function renderACMEActivityLogs(logs, available = true) {
  if (!els.acmeDialog.open) return;
  els.acmeLogList.innerHTML = "";
  if (!available) {
    appendACMEEmptyState("No Caddy service log available in the current mode.");
    return;
  }
  const filtered = filterACMELogs(logs);
  if (!filtered.length) {
    appendACMEEmptyState("Waiting for matching Caddy log lines for this host.");
    return;
  }
  for (const entry of filtered.slice(0, 12)) {
    const row = document.createElement("div");
    row.className = "log-row service-log-row";
    row.innerHTML = `
      <time></time>
      <span class="log-method"></span>
      <span class="log-path"></span>
      <span class="log-status badge"></span>
    `;
    row.children[0].textContent = new Date(entry.time).toLocaleTimeString();
    row.children[0].title = new Date(entry.time).toLocaleString();
    row.children[1].textContent = entry.action || "service";
    row.children[2].textContent = entry.message || "-";
    row.children[2].title = entry.message || "";
    row.children[3].textContent = entry.status || "INFO";
    row.children[3].classList.toggle("off", String(entry.status || "").toUpperCase() === "ERROR");
    els.acmeLogList.append(row);
  }
}

function appendACMEEmptyState(message) {
  const empty = document.createElement("div");
  empty.className = "empty-state";
  empty.textContent = message;
  els.acmeLogList.append(empty);
}

function filterACMELogs(logs) {
  if (!acmeDialogContext) return [];
  const { domain, startedAt } = acmeDialogContext;
  return logs.filter((entry) => {
    const time = Date.parse(entry.time || "");
    const recent = Number.isNaN(time) ? true : time >= startedAt;
    if (!recent) return false;
    const haystack = `${entry.action || ""} ${entry.message || ""} ${entry.status || ""}`.toLowerCase();
    if (domain && haystack.includes(domain)) return true;
    return [
      "load complete",
      "received request",
      "automatic tls certificate management",
      "obtaining certificate",
      "trying to solve challenge",
      "authorization finalized",
      "validations succeeded",
      "certificate obtained successfully",
      "server running",
      "served key authentication",
      "releasing lock",
      "lock acquired",
    ].some((needle) => haystack.includes(needle));
  });
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

async function renewCertificate(site) {
  if (!site?.id || !confirm(`Force renew certificate for ${site.address}?`)) return;
  try {
    showACMEStatus(site);
    await request(`/api/certificates/renew/${site.id}`, { method: "POST" });
    setStatus(`Certificate renew started for ${site.address}`);
    await loadSites();
  } catch (err) {
    closeACMEStatus();
    setStatus(err.message);
  }
}

async function request(url, options = {}) {
  const headers = new Headers(options.headers || {});
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (needsCSRFFromMethod(options.method || "GET")) {
    headers.set("X-CSRF-Token", getCSRFToken());
  }
  const response = await fetch(url, {
    ...options,
    headers,
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
  const response = await fetch(url, {
    method: "POST",
    body,
    headers: { "X-CSRF-Token": getCSRFToken() },
  });
  if (response.status === 401) {
    window.location.assign("/login.html");
    throw new Error("authentication required");
  }
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || "Upload failed");
  return data;
}

async function logout() {
  await fetch("/api/auth/logout", {
    method: "POST",
    headers: { "X-CSRF-Token": getCSRFToken() },
  });
  window.location.assign("/login.html");
}

function getCSRFToken() {
  const match = document.cookie.match(/(?:^|; )caddymgm_csrf=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : "";
}

function needsCSRFFromMethod(method) {
  const normalized = String(method || "GET").toUpperCase();
  return !["GET", "HEAD", "OPTIONS"].includes(normalized);
}

function initialsForUser(value) {
  const cleaned = String(value || "")
    .trim()
    .replace(/[^a-zA-Z0-9@._ -]/g, "");
  if (!cleaned) return "CM";
  const parts = cleaned.split(/[@._ -]+/).filter(Boolean);
  if (parts.length === 0) return cleaned.slice(0, 2).toUpperCase();
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return `${parts[0][0] || ""}${parts[1][0] || ""}`.toUpperCase();
}

function getMode() {
  return document.querySelector("input[name='mode']:checked")?.value || "";
}

function protocolForSite(site) {
  return site?.tlsMode && site.tlsMode !== "off" ? "https" : "http";
}

function setStatus(message) {
  const text = String(message || "");
  els.status.textContent = text.length > 96 ? `${text.slice(0, 93)}...` : text;
  els.status.title = text;
}
