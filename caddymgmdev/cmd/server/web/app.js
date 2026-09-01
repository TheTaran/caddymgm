const els = {
  status: document.querySelector("#status"),
  saveConfirmation: document.querySelector("#save-confirmation"),
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
  securityOverviewSummary: document.querySelector("#security-overview-summary"),
  securityRequests: document.querySelector("#security-requests"),
  securityManagedBlocks: document.querySelector("#security-managed-blocks"),
  securityGeoBlocks: document.querySelector("#security-geo-blocks"),
  securityManualBlocks: document.querySelector("#security-manual-blocks"),
  securityExternalBlocks: document.querySelector("#security-external-blocks"),
  securityUnclassified: document.querySelector("#security-unclassified"),
  security4xx: document.querySelector("#security-4xx"),
  security5xx: document.querySelector("#security-5xx"),
  securityOverviewPeriod: document.querySelector("#security-overview-period"),
  securityEvents: document.querySelector("#security-events"),
  securityRuleCounts: document.querySelector("#security-rule-counts"),
  geoMap: document.querySelector("#geo-map"),
  geoMapSummary: document.querySelector("#geo-map-summary"),
  geoTopIPList: document.querySelector("#geo-top-ip-list"),
  geoTopIPSummary: document.querySelector("#geo-top-ip-summary"),
  geoIPScope: document.querySelector("#geo-ip-scope"),
  geoIPHost: document.querySelector("#geo-ip-host"),
  geoIPLimit: document.querySelector("#geo-ip-limit"),
  geoMapDetails: document.querySelector("#geo-map-details"),
  siteList: document.querySelector("#site-list"),
  hostFilterProtocol: document.querySelector("#host-filter-protocol"),
  hostFilterCertificateProvider: document.querySelector("#host-filter-certificate-provider"),
  hostFilterVisibility: document.querySelector("#host-filter-visibility"),
  hostFilterMode: document.querySelector("#host-filter-mode"),
  hostFilterUpstreamTLS: document.querySelector("#host-filter-upstream-tls"),
  hostFilterStatus: document.querySelector("#host-filter-status"),
  hostFilterAuth: document.querySelector("#host-filter-auth"),
  hostFilterComment: document.querySelector("#host-filter-comment"),
  hostFilterReset: document.querySelector("#host-filter-reset"),
  hostFilterSummary: document.querySelector("#host-filter-summary"),
  proxyGrid: document.querySelector("#proxy-hosts-grid"),
  webProtectionForm: document.querySelector("#web-protection-form"),
  webProtectionEnabled: document.querySelector("#web-protection-enabled"),
  webProtectionCountryMode: document.querySelector("#web-protection-country-mode"),
  webProtectionModeBlock: document.querySelector("#web-protection-mode-block"),
  webProtectionModeAllow: document.querySelector("#web-protection-mode-allow"),
  webProtectionCountriesButton: document.querySelector("#web-protection-countries-button"),
  webProtectionCountriesSummary: document.querySelector("#web-protection-countries-summary"),
  countryPickerDialog: document.querySelector("#country-picker-dialog"),
  countryPickerSearch: document.querySelector("#country-picker-search"),
  countryPickerList: document.querySelector("#country-picker-list"),
  countryPickerApply: document.querySelector("#country-picker-apply"),
  countryPickerModeHint: document.querySelector("#country-picker-mode-hint"),
  ipListsPanel: document.querySelector("#ip-lists-panel"),
  manualIPListList: document.querySelector("#manual-ip-list-list"),
  manualIPListEmpty: document.querySelector("#manual-ip-list-empty"),
  manualIPListAdd: document.querySelector("#manual-ip-list-add"),
  manualIPListDialog: document.querySelector("#manual-ip-list-dialog"),
  manualIPListForm: document.querySelector("#manual-ip-list-form"),
  manualIPListIndex: document.querySelector("#manual-ip-list-index"),
  manualIPListName: document.querySelector("#manual-ip-list-name"),
  manualIPListMode: document.querySelector("#manual-ip-list-mode"),
  manualIPListReference: document.querySelector("#manual-ip-list-reference"),
  manualIPListEntries: document.querySelector("#manual-ip-list-entries"),
  manualIPListClose: document.querySelector("#manual-ip-list-close"),
  manualIPListCancel: document.querySelector("#manual-ip-list-cancel"),
  assignIPDialog: document.querySelector("#assign-ip-dialog"),
  assignIPAddress: document.querySelector("#assign-ip-address"),
  assignIPValue: document.querySelector("#assign-ip-value"),
  assignIPList: document.querySelector("#assign-ip-list"),
  assignIPForm: document.querySelector("#assign-ip-form"),
  assignIPClose: document.querySelector("#assign-ip-close"),
  assignIPCancel: document.querySelector("#assign-ip-cancel"),
  externalBlocklistsForm: document.querySelector("#external-blocklists-form"),
  externalBlocklistList: document.querySelector("#external-blocklist-list"),
  externalBlocklistEmpty: document.querySelector("#external-blocklist-empty"),
  externalBlocklistAdd: document.querySelector("#external-blocklist-add"),
  externalBlocklistTotal: document.querySelector("#external-blocklist-total"),
  editor: document.querySelector("#editor"),
  form: document.querySelector("#site-form"),
  formTitle: document.querySelector("#form-title"),
  editorHostAddress: document.querySelector("#editor-host-address"),
  editorHostProtocol: document.querySelector("#editor-host-protocol"),
  editorHostMode: document.querySelector("#editor-host-mode"),
  editorHostStatus: document.querySelector("#editor-host-status"),
  id: document.querySelector("#site-id"),
  address: document.querySelector("#address"),
  comment: document.querySelector("#comment"),
  upstream: document.querySelector("#upstream"),
  skipTlsVerify: document.querySelector("#skip-tls-verify"),
  rewriteRedirects: document.querySelector("#rewrite-redirects"),
  redirectOrigins: document.querySelector("#redirect-origins"),
  hstsEnabled: document.querySelector("#hsts-enabled"),
  securityHeaderProfile: document.querySelector("#security-header-profile"),
  compressionProfile: document.querySelector("#compression-profile"),
  root: document.querySelector("#root"),
  upstreamRow: document.querySelector("#upstream-row"),
  skipTlsVerifyRow: document.querySelector("#skip-tls-verify-row"),
  rewriteRedirectsRow: document.querySelector("#rewrite-redirects-row"),
  redirectOriginsRow: document.querySelector("#redirect-origins-row"),
  hstsEnabledRow: document.querySelector("#hsts-enabled-row"),
  rewriteRedirectsHint: document.querySelector("#rewrite-redirects-hint"),
  redirectOriginsHint: document.querySelector("#redirect-origins-hint"),
  proxyBehaviorOverview: document.querySelector("#proxy-behavior-overview"),
  proxyOverviewHost: document.querySelector("#proxy-overview-host"),
  proxyOverviewForwarded: document.querySelector("#proxy-overview-forwarded"),
  proxyOverviewRequestRoute: document.querySelector("#proxy-overview-request-route"),
  proxyOverviewRedirectList: document.querySelector("#proxy-overview-redirect-list"),
  hstsEnabledHint: document.querySelector("#hsts-enabled-hint"),
  securityHeaderProfileHint: document.querySelector("#security-header-profile-hint"),
  rootRow: document.querySelector("#root-row"),
  extra: document.querySelector("#extra"),
  enabled: document.querySelector("#enabled"),
  logsEnabled: document.querySelector("#logs-enabled"),
  tlsEnabled: document.querySelector("#tls-enabled"),
  acmeIssuerRow: document.querySelector("#acme-issuer-row"),
  acmeIssuer: document.querySelector("#acme-issuer"),
  tlsMinVersion: document.querySelector("#tls-min-version"),
  tlsMaxVersion: document.querySelector("#tls-max-version"),
  tlsControlsHint: document.querySelector("#tls-controls-hint"),
  protectionOverride: document.querySelector("#protection-override"),
  protectionOverrideFields: document.querySelector("#protection-override-fields"),
  siteProtectionEnabled: document.querySelector("#site-protection-enabled"),
  siteProtectionCountryMode: document.querySelector("#site-protection-country-mode"),
  siteProtectionCountries: document.querySelector("#site-protection-countries"),
  siteProtectionBlockedIPs: document.querySelector("#site-protection-blocked-ips"),
  siteProtectionAllowedIPs: document.querySelector("#site-protection-allowed-ips"),
  siteAuthEnabled: document.querySelector("#site-auth-enabled"),
  basicAuthEnabled: document.querySelector("#basic-auth-enabled"),
  basicAuthUsername: document.querySelector("#basic-auth-username"),
  basicAuthPassword: document.querySelector("#basic-auth-password"),
  basicAuthUsernameRow: document.querySelector("#basic-auth-username-row"),
  basicAuthPasswordRow: document.querySelector("#basic-auth-password-row"),
  basicAuthHint: document.querySelector("#basic-auth-hint"),
  delete: document.querySelector("#delete"),
  totalSites: document.querySelector("#total-sites"),
  activeSites: document.querySelector("#active-sites"),
  proxySites: document.querySelector("#proxy-sites"),
  staticSites: document.querySelector("#static-sites"),
  caddymgmVersion: document.querySelector("#caddymgm-version"),
  caddyVersion: document.querySelector("#caddy-version"),
  goVersion: document.querySelector("#go-version"),
  caddymgmUpdate: document.querySelector("#caddymgm-update"),
  caddyUpdate: document.querySelector("#caddy-update"),
  goUpdate: document.querySelector("#go-update"),
  logSiteFilter: document.querySelector("#log-site-filter"),
  logStreamLabel: document.querySelector("#log-stream-label"),
  logList: document.querySelector("#log-list"),
  serviceLogList: document.querySelector("#service-log-list"),
  serviceLogToggle: document.querySelector("#service-log-toggle"),
  siteLogToggle: document.querySelector("#site-log-toggle"),
  oidcLogList: document.querySelector("#oidc-log-list"),
  oidcLogToggle: document.querySelector("#oidc-log-toggle"),
  protectionEventList: document.querySelector("#protection-event-list"),
  protectionEventToggle: document.querySelector("#protection-event-toggle"),
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
  settingsAccessEnabled: document.querySelector("#settings-access-enabled"),
  settingsAccessName: document.querySelector("#settings-access-name"),
  settingsAccessIssuer: document.querySelector("#settings-access-issuer"),
  settingsAccessClientID: document.querySelector("#settings-access-client-id"),
  settingsAccessClientSecret: document.querySelector("#settings-access-client-secret"),
  settingsAccessScopes: document.querySelector("#settings-access-scopes"),
  settingsAccessGateway: document.querySelector("#settings-access-gateway"),
  settingsAccessACME: document.querySelector("#settings-access-acme"),
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
  acmeResult: document.querySelector("#acme-result"),
  acmeLogList: document.querySelector("#acme-log-list"),
};

const viewTitles = {
  dashboard: ["Dashboard", "Web Hosts Overview"],
  "proxy-hosts": ["Web Hosts", "Website Configuration"],
  "web-protection": ["Web Protection", "GEO IP Blocking"],
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
  "web-protection": `
    <svg viewBox="0 0 24 24" fill="none"><path d="M12 3 6 5.5v5.5c0 4 2.56 7.2 6 10 3.44-2.8 6-6 6-10V5.5L12 3Z" /><path d="m9.5 12 1.8 1.8 3.7-4" /></svg>
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
const hostFilters = {
  protocol: "all",
  certificateProvider: "all",
  visibility: "all",
  mode: "all",
  upstreamTls: "all",
  status: "all",
  auth: "all",
  comment: "",
};
const auxiliarySort = {
  certificates: { key: "domain", direction: "asc" },
  "service-logs": { key: "time", direction: "desc" },
  "site-logs": { key: "time", direction: "desc" },
  "oidc-logs": { key: "time", direction: "desc" },
  "protection-events": { key: "time", direction: "desc" },
};
const tableFilters = {
  dashboard: { protocol: "all", certificateProvider: "", visibility: "all", mode: "all", upstreamTls: "all", status: "all", auth: "all", comment: "" },
  certificates: { issuer: "", expires: "all", status: "all" },
  "service-logs": { type: "", message: "", status: "" },
  "site-logs": { method: "all", path: "", status: "" },
  "oidc-logs": { type: "", user: "", ip: "", site: "", status: "" },
  "protection-events": { type: "", site: "", ip: "", message: "" },
};
const hostCollator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });
let settings = null;
let logPollTimer = null;
let acmeDialogTimer = null;
let acmeDialogContext = null;
let serviceLogsExpanded = false;
let siteLogsExpanded = false;
let oidcLogsExpanded = false;
let protectionEventsExpanded = false;
let editingSite = null;
let latestSiteLogs = [];
let latestServiceLogs = [];
let latestOIDCLogs = [];
let latestProtectionEvents = [];
let latestGeoTopIPs = [];
let latestServiceLogsAvailable = true;
const LOG_PREVIEW_LIMIT = 10;

document.querySelector("#logout").addEventListener("click", logout);
document.querySelector("#new-site").addEventListener("click", () => {
  showView("proxy-hosts");
  editSite();
});
document.querySelector("#cancel").addEventListener("click", closeEditor);
els.form.addEventListener("submit", saveSite);
els.form.addEventListener("invalid", handleSiteFormInvalid, true);
els.form.addEventListener("invalid", (event) => {
  const panel = event.target.closest("[data-site-editor-panel]");
  if (panel) showSiteEditorTab(panel.dataset.siteEditorPanel);
}, true);
els.delete.addEventListener("click", deleteSite);
els.logSiteFilter.addEventListener("change", () => {
  siteLogsExpanded = false;
  loadLogs();
  syncLogPolling();
});
els.serviceLogToggle.addEventListener("click", toggleServiceLogsExpanded);
els.siteLogToggle.addEventListener("click", toggleSiteLogsExpanded);
els.oidcLogToggle.addEventListener("click", toggleOIDCLogsExpanded);
els.protectionEventToggle.addEventListener("click", toggleProtectionEventsExpanded);
els.securityOverviewPeriod.addEventListener("change", loadSecurityOverview);
els.geoIPScope.addEventListener("change", renderTopIPs);
els.geoIPHost.addEventListener("change", renderTopIPs);
els.geoIPLimit.addEventListener("change", renderTopIPs);
[els.hostFilterProtocol, els.hostFilterCertificateProvider, els.hostFilterVisibility, els.hostFilterMode, els.hostFilterUpstreamTLS, els.hostFilterStatus, els.hostFilterAuth, els.hostFilterComment].forEach((filter) => {
  filter.addEventListener(filter.tagName === "INPUT" ? "input" : "change", applyHostFilters);
});
els.hostFilterReset.addEventListener("click", resetHostFilters);
document.querySelectorAll(".table-filter").forEach((filter) => {
  filter.addEventListener(filter.tagName === "INPUT" ? "input" : "change", applyTableFilter);
});
document.querySelectorAll("[data-reset-filters]").forEach((button) => {
  button.addEventListener("click", resetTableFilters);
});
els.settingsForm.addEventListener("submit", saveSettings);
els.webProtectionForm.addEventListener("submit", saveWebProtection);
els.webProtectionModeBlock.addEventListener("click", () => setWebProtectionCountryMode("block"));
els.webProtectionModeAllow.addEventListener("click", () => setWebProtectionCountryMode("allow"));
els.manualIPListForm.addEventListener("submit", saveManualIPList);
els.manualIPListAdd.addEventListener("click", () => openManualIPListDialog());
els.manualIPListClose.addEventListener("click", () => els.manualIPListDialog.close());
els.manualIPListCancel.addEventListener("click", () => els.manualIPListDialog.close());
els.assignIPForm.addEventListener("submit", assignIPToManualList);
els.assignIPClose.addEventListener("click", () => els.assignIPDialog.close());
els.assignIPCancel.addEventListener("click", () => els.assignIPDialog.close());
els.manualIPListList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-manual-list-action]");
  if (!button) return;
  const index = Number(button.closest("[data-manual-list-index]")?.dataset.manualListIndex);
  if (button.dataset.manualListAction === "edit") openManualIPListDialog(index);
  if (button.dataset.manualListAction === "remove") removeManualIPList(index);
});
els.externalBlocklistsForm.addEventListener("submit", saveExternalBlocklists);
els.externalBlocklistAdd.addEventListener("click", addExternalBlocklist);
els.externalBlocklistList.addEventListener("click", (event) => {
  const removeButton = event.target.closest("[data-remove-blocklist]");
  if (removeButton) {
    removeButton.closest(".blocklist-row")?.remove();
    els.externalBlocklistEmpty.hidden = els.externalBlocklistList.children.length > 0;
    return;
  }
  const updateButton = event.target.closest("[data-update-blocklist]");
  if (!updateButton) return;
  const url = updateButton.closest(".blocklist-row")?.querySelector('[data-blocklist-field="url"]')?.value.trim();
  if (url) persistExternalBlocklists(url, false);
});
document.querySelectorAll("[data-protection-tab]").forEach((tab) => tab.addEventListener("click", () => showProtectionTab(tab.dataset.protectionTab)));
els.webProtectionCountriesButton.addEventListener("click", openCountryPicker);
els.countryPickerSearch.addEventListener("input", renderCountryPicker);
els.countryPickerList.addEventListener("change", (event) => {
  if (!(event.target instanceof HTMLInputElement) || event.target.type !== "checkbox") return;
  const selected = new Set(selectedProtectionCountries);
  if (event.target.checked) selected.add(event.target.value); else selected.delete(event.target.value);
  selectedProtectionCountries = [...selected].sort();
});
els.countryPickerApply.addEventListener("click", applyCountryPicker);
els.settingsPassword.addEventListener("input", syncSettingsPasswordConfirmation);
els.settingsPasswordConfirm.addEventListener("input", syncSettingsPasswordConfirmation);
els.settingsWebTLSEnabled.addEventListener("change", syncSettingsWebTLS);
els.certificateForm.addEventListener("submit", saveIssuer);
els.issuerNew.addEventListener("click", () => editIssuer({}));
els.issuerReset.addEventListener("click", closeIssuerForm);
els.issuerDelete.addEventListener("click", deleteIssuer);
els.issuerRootCAUploadButton.addEventListener("click", uploadRootCA);
els.tlsEnabled.addEventListener("change", syncTLSMode);
els.tlsMinVersion.addEventListener("change", () => {
  if (els.tlsMinVersion.value === "tls1.3" && els.tlsMaxVersion.value === "tls1.2") els.tlsMaxVersion.value = "tls1.3";
});
els.tlsMaxVersion.addEventListener("change", () => {
  if (els.tlsMaxVersion.value === "tls1.2" && els.tlsMinVersion.value === "tls1.3") els.tlsMinVersion.value = "tls1.2";
});
els.hstsEnabled.addEventListener("change", syncTLSMode);
els.protectionOverride.addEventListener("change", syncProtectionOverride);
els.securityHeaderProfile.addEventListener("change", syncSecurityHeaderProfile);
els.rewriteRedirects.addEventListener("change", syncMode);
els.siteAuthEnabled.addEventListener("change", syncSiteAuth);
els.basicAuthEnabled.addEventListener("change", syncBasicAuth);
els.upstream.addEventListener("input", syncUpstreamTLS);
els.upstream.addEventListener("input", syncProxyBehaviorOverview);
els.address.addEventListener("input", syncProxyBehaviorOverview);
els.address.addEventListener("input", syncEditorHostSummary);
els.comment.addEventListener("input", syncEditorHostSummary);
els.enabled.addEventListener("change", syncEditorHostSummary);
els.redirectOrigins.addEventListener("input", syncProxyBehaviorOverview);
els.acmeDialogClose.addEventListener("click", closeACMEStatus);
els.acmeDialog.addEventListener("close", stopACMEStatusPolling);
els.acmeDialog.addEventListener("cancel", stopACMEStatusPolling);
els.navItems.forEach((item) => item.addEventListener("click", () => showView(item.dataset.view)));
document.querySelectorAll("[data-logs-tab]").forEach((tab) => {
  tab.addEventListener("click", () => showLogsTab(tab.dataset.logsTab));
  tab.addEventListener("keydown", handleLogsTabKeydown);
});
document.querySelectorAll("[data-settings-tab]").forEach((tab) => {
  tab.addEventListener("click", () => showSettingsTab(tab.dataset.settingsTab));
  tab.addEventListener("keydown", handleSettingsTabKeydown);
});
document.querySelectorAll("[data-auth-tab]").forEach((tab) => {
  tab.addEventListener("click", () => showAuthTab(tab.dataset.authTab));
  tab.addEventListener("keydown", handleAuthTabKeydown);
});
document.querySelectorAll("[data-site-editor-tab]").forEach((tab) => {
  tab.addEventListener("click", () => showSiteEditorTab(tab.dataset.siteEditorTab));
  tab.addEventListener("keydown", handleSiteEditorTabKeydown);
});
document.querySelectorAll("[data-sort-table]").forEach((header) => {
  header.addEventListener("click", (event) => {
    const button = event.target.closest(".sort-button");
    if (!button) return;
    const table = header.dataset.sortTable;
    const sort = hostSort[table] || auxiliarySort[table];
    if (!sort) return;
    if (sort.key === button.dataset.sort) {
      sort.direction = sort.direction === "asc" ? "desc" : "asc";
    } else {
      sort.key = button.dataset.sort;
      sort.direction = "asc";
    }
    renderSortedTable(table);
  });
});
document.querySelectorAll("input[name='mode']").forEach((input) => {
  input.addEventListener("change", syncMode);
});

init();

async function init() {
  await Promise.all([loadSites(), loadSettings(), loadProfile(), loadVersions(), loadGeoMap(), loadSecurityOverview()]);
  renderLogs([]);
  renderServiceLogs([]);
  renderOIDCLogs([]);
  renderProtectionEvents([]);
}

async function loadVersions() {
  try {
    const versions = await request("/api/versions");
    renderVersionStatus("caddymgm", versions.caddymgm);
    renderVersionStatus("caddy", versions.caddy);
    renderVersionStatus("go", versions.go);
  } catch (_err) {
    renderVersionStatus("caddymgm", { current: "unknown" });
    renderVersionStatus("caddy", { current: "unknown" });
    renderVersionStatus("go", { current: "unknown" });
  }
}

function renderVersionStatus(name, info = {}) {
  const elements = {
    caddymgm: [els.caddymgmVersion, els.caddymgmUpdate],
    caddy: [els.caddyVersion, els.caddyUpdate],
    go: [els.goVersion, els.goUpdate],
  };
  const [versionElement, updateElement] = elements[name];
  versionElement.textContent = info.current || "unknown";
  updateElement.classList.toggle("update-available", !!info.updateAvailable);
  updateElement.classList.toggle("up-to-date", !!info.latest && !info.updateAvailable);
  updateElement.classList.toggle("check-unavailable", !info.latest);
  if (info.releaseUrl) updateElement.href = info.releaseUrl;
  if (info.updateAvailable) {
    updateElement.textContent = info.latest;
  } else if (info.latest) {
    updateElement.textContent = "Up to date";
  } else {
    updateElement.textContent = "Update check unavailable";
  }
}

function showLogsTab(name, focus = false) {
  document.querySelectorAll("[data-logs-tab]").forEach((tab) => {
    const active = tab.dataset.logsTab === name;
    tab.classList.toggle("active", active);
    tab.setAttribute("aria-selected", String(active));
    tab.tabIndex = active ? 0 : -1;
    if (active && focus) tab.focus();
  });
  document.querySelectorAll("[data-logs-panel]").forEach((panel) => {
    panel.hidden = panel.dataset.logsPanel !== name;
  });
  if (name === "service") loadServiceLogs();
  if (name === "websites") loadLogs();
  if (name === "oidc") loadOIDCLogs();
  if (name === "web-protection") loadProtectionEvents();
}

function handleLogsTabKeydown(event) {
  handleHorizontalTabKeydown(event, "[data-logs-tab]", (tab) => showLogsTab(tab.dataset.logsTab, true));
}

function showSettingsTab(name, focus = false) {
  document.querySelectorAll("[data-settings-tab]").forEach((tab) => {
    const active = tab.dataset.settingsTab === name;
    tab.classList.toggle("active", active);
    tab.setAttribute("aria-selected", String(active));
    tab.tabIndex = active ? 0 : -1;
    if (active && focus) tab.focus();
  });
  document.querySelectorAll("[data-settings-panel]").forEach((panel) => {
    panel.hidden = panel.dataset.settingsPanel !== name;
  });
}

function handleSettingsTabKeydown(event) {
  handleHorizontalTabKeydown(event, "[data-settings-tab]", (tab) => showSettingsTab(tab.dataset.settingsTab, true));
}

function showAuthTab(name, focus = false) {
  document.querySelectorAll("[data-auth-tab]").forEach((tab) => {
    const active = tab.dataset.authTab === name;
    tab.classList.toggle("active", active);
    tab.setAttribute("aria-selected", String(active));
    tab.tabIndex = active ? 0 : -1;
    if (active && focus) tab.focus();
  });
  document.querySelectorAll("[data-auth-panel]").forEach((panel) => {
    panel.hidden = panel.dataset.authPanel !== name;
  });
}

function handleAuthTabKeydown(event) {
  handleHorizontalTabKeydown(event, "[data-auth-tab]", (tab) => showAuthTab(tab.dataset.authTab, true));
}

function showSiteEditorTab(name, focus = false) {
  document.querySelectorAll("[data-site-editor-tab]").forEach((tab) => {
    const active = tab.dataset.siteEditorTab === name;
    tab.classList.toggle("active", active);
    tab.setAttribute("aria-selected", String(active));
    tab.tabIndex = active ? 0 : -1;
    if (active && focus) tab.focus();
  });
  document.querySelectorAll("[data-site-editor-panel]").forEach((panel) => {
    panel.hidden = panel.dataset.siteEditorPanel !== name;
  });
}

function handleSiteFormInvalid(event) {
  const panel = event.target.closest("[data-site-editor-panel]");
  if (panel?.hidden) showSiteEditorTab(panel.dataset.siteEditorPanel);
}

function handleSiteEditorTabKeydown(event) {
  handleHorizontalTabKeydown(event, "[data-site-editor-tab]", (tab) => showSiteEditorTab(tab.dataset.siteEditorTab, true));
}

function handleHorizontalTabKeydown(event, selector, activate) {
  if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
  event.preventDefault();
  const tabs = [...document.querySelectorAll(selector)];
  const current = tabs.indexOf(event.currentTarget);
  let next = event.key === "Home" ? 0 : event.key === "End" ? tabs.length - 1 : current + (event.key === "ArrowRight" ? 1 : -1);
  next = (next + tabs.length) % tabs.length;
  activate(tabs[next]);
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
    loadOIDCLogs();
    loadProtectionEvents();
    syncLogPolling();
  } else {
    stopLogPolling();
  }
  if (view === "dashboard") {
    loadGeoMap();
    loadSecurityOverview();
  }
  if (view === "settings") loadSettings();
  if (view === "certificates") renderCertificatesView();
}

function protectionValues(value) {
  return String(value || "").split(/[\n,]+/).map((item) => item.trim()).filter(Boolean);
}

function setWebProtectionCountryMode(mode) {
  els.webProtectionCountryMode.value = mode;
  const block = mode === "block";
  els.webProtectionModeBlock.classList.toggle("active", block);
  els.webProtectionModeAllow.classList.toggle("active", !block);
  els.webProtectionModeBlock.setAttribute("aria-pressed", String(block));
  els.webProtectionModeAllow.setAttribute("aria-pressed", String(!block));
}

async function loadSecurityOverview() {
  try {
    renderSecurityOverview(await request(`/api/security-overview?period=${encodeURIComponent(els.securityOverviewPeriod.value || "1d")}`));
  } catch (err) {
    els.securityOverviewSummary.textContent = "Security overview unavailable";
    els.securityEvents.innerHTML = `<p class="error-text">${escapeHTML(err.message)}</p>`;
  }
}

function renderSecurityOverview(data) {
  const format = (value) => Number(value || 0).toLocaleString();
  els.securityRequests.textContent = format(data.requests);
  els.securityManagedBlocks.textContent = format(data.managedBlocks);
  els.securityGeoBlocks.textContent = format(data.geoBlocks);
  els.securityManualBlocks.textContent = format(data.manualIPBlocks);
  els.securityExternalBlocks.textContent = format(data.externalBlocks);
  els.securityUnclassified.textContent = format(data.unclassified403);
  els.security4xx.textContent = format(data.clientErrors);
  els.security5xx.textContent = format(data.serverErrors);
  const periodLabel = els.securityOverviewPeriod.options[els.securityOverviewPeriod.selectedIndex]?.text || "24 hours";
  els.securityOverviewSummary.textContent = `${format(data.requests)} retained request${Number(data.requests) === 1 ? "" : "s"} · ${periodLabel}`;
  const rules = data.ruleCounts || {};
  els.securityRuleCounts.innerHTML = `<div><span>Countries selected</span><strong>${format(rules.selectedCountries)}</strong></div><div><span>Manual blocked IPs</span><strong>${format(rules.manualBlockedIPs)}</strong></div><div><span>Allowed IPs</span><strong>${format(rules.allowedIPs)}</strong></div><div><span>External blocked IPs</span><strong>${format(rules.externalBlockedIPs)}</strong></div>`;
  const events = data.events || [];
  els.securityEvents.classList.toggle("has-multiple-events", events.length > 1);
  els.securityEvents.innerHTML = events.length ? events.map((event) => `<article class="security-event"><div><strong>${escapeHTML(event.reason)}</strong><span>${escapeHTML(event.site)} · ${escapeHTML(event.address)}${event.country ? ` · ${escapeHTML(event.country)}` : ""}</span></div><time>${event.time ? new Date(event.time).toLocaleString() : ""}</time></article>`).join("") : '<p class="muted">No managed protection events in retained access logs.</p>';
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[character]));
}

function showProtectionTab(tab) {
  els.webProtectionForm.closest("section").hidden = tab !== "geo";
  els.ipListsPanel.hidden = tab !== "lists";
  document.querySelectorAll("[data-protection-tab]").forEach((button) => button.classList.toggle("active", button.dataset.protectionTab === tab));
}

async function saveExternalBlocklists(event) {
  event.preventDefault();
  await persistExternalBlocklists("", true);
}

function collectExternalBlocklists() {
  return [...els.externalBlocklistList.querySelectorAll(".blocklist-row")].map((row) => ({
    name: row.querySelector('[data-blocklist-field="name"]').value.trim(),
    url: row.querySelector('[data-blocklist-field="url"]').value.trim(),
  }));
}

async function persistExternalBlocklists(refreshURL = "", refreshAll = false) {
  const feeds = collectExternalBlocklists();
  try {
    settings = await request("/api/settings", { method: "PUT", body: JSON.stringify({ ...settings, externalBlocklists: feeds, refreshExternalBlocklists: refreshAll, refreshExternalBlocklistUrl: refreshURL }) });
    renderExternalBlocklists(settings.externalBlocklists || []);
    setStatus(`${feeds.length} external blocklist${feeds.length === 1 ? "" : "s"} active`);
    showConfirmation(refreshURL ? "External blocklist updated" : "External blocklists refreshed", "success");
    loadSecurityOverview();
  } catch (err) { setStatus(err.message); }
}

function renderExternalBlocklists(feeds) {
  els.externalBlocklistList.innerHTML = feeds.map((feed, index) => `<div class="blocklist-row" data-blocklist-row="${index}"><input data-blocklist-field="name" maxlength="80" required placeholder="Blocklist name" value="${escapeHTML(feed.name)}" /><input data-blocklist-field="url" type="url" required placeholder="https://example.org/blocklist.txt" value="${escapeHTML(feed.url)}" /><span class="blocklist-meta">${Number(feed.count || 0).toLocaleString()} IPs</span><span class="blocklist-meta">${feed.updatedAt ? new Date(feed.updatedAt).toLocaleString() : "Not recorded"}</span><span class="blocklist-actions"><button type="button" class="secondary" data-update-blocklist>Update</button><button type="button" class="danger" data-remove-blocklist>Remove</button></span></div>`).join("");
  els.externalBlocklistEmpty.hidden = feeds.length > 0;
  const total = Number(settings?.externalBlockedIpCount || 0);
  els.externalBlocklistTotal.textContent = `${total.toLocaleString()} blocked IP${total === 1 ? "" : "s"} across all lists`;
}

function addExternalBlocklist() {
  const feeds = collectExternalBlocklists();
  feeds.push({ name: "", url: "" });
  renderExternalBlocklists(feeds);
  els.externalBlocklistList.lastElementChild?.querySelector('[data-blocklist-field="name"]')?.focus();
}

let protectionCountries = [];
let selectedProtectionCountries = [];
async function loadProtectionCountries() {
  if (protectionCountries.length) return;
  const data = await request("/api/geo-countries");
  protectionCountries = data.countries || [];
}
function renderSelectedProtectionCountries() {
  const byCode = new Map(protectionCountries.map((country) => [country.code, country]));
  const selected = selectedProtectionCountries.map((code) => byCode.get(code)).filter(Boolean);
  if (!selected.length) {
    els.webProtectionCountriesSummary.innerHTML = '<p class="field-hint">No countries selected.</p>';
    return;
  }
  els.webProtectionCountriesSummary.innerHTML = selected.map((country) => `<span class="selected-country"><img src="/api/geo-flag/${country.code}" width="20" height="15" alt="" /><span>${escapeHTML(country.name)}</span></span>`).join("");
}
async function openCountryPicker() {
  selectedProtectionCountries = [...(settings?.webProtection?.blockedCountries || [])];
  els.countryPickerSearch.value = "";
  els.countryPickerModeHint.textContent = els.webProtectionCountryMode.value === "allow" ? "Only the selected countries will be allowed." : "The selected countries will be blocked.";
  els.countryPickerList.innerHTML = '<p class="muted">Loading countries from GeoLite2...</p>';
  els.countryPickerDialog.showModal();
  try {
    await loadProtectionCountries();
    renderCountryPicker();
  } catch (err) {
    els.countryPickerList.innerHTML = `<p class="error-text">${escapeHTML(err.message)}</p>`;
  }
}
function renderCountryPicker() { const query = els.countryPickerSearch.value.trim().toLowerCase(); const matches = protectionCountries.filter((country) => !query || country.name.toLowerCase().includes(query) || country.code.toLowerCase().includes(query)); els.countryPickerList.innerHTML = matches.length ? matches.map((country) => `<label class="country-picker-row"><input type="checkbox" value="${country.code}" ${selectedProtectionCountries.includes(country.code) ? "checked" : ""} /><img src="/api/geo-flag/${country.code}" width="24" height="18" alt="" loading="lazy" /><strong>${escapeHTML(country.name)}</strong><code>${country.code}</code></label>`).join("") : '<p class="muted">No countries match this search.</p>'; }
function applyCountryPicker() { settings.webProtection = { ...(settings.webProtection || {}), blockedCountries: selectedProtectionCountries }; renderSelectedProtectionCountries(); els.countryPickerDialog.close(); }

async function saveWebProtection(event) {
  event.preventDefault();
  const next = { ...settings, webProtection: {
    enabled: els.webProtectionEnabled.checked,
    countryMode: els.webProtectionCountryMode.value,
    blockedCountries: selectedProtectionCountries,
    blockedIps: settings?.webProtection?.blockedIps || [],
    allowedIps: settings?.webProtection?.allowedIps || [],
  }};
  try {
    settings = await request("/api/settings", { method: "PUT", body: JSON.stringify(next) });
    setStatus("GEO IP blocking defaults saved");
    showConfirmation("GEO IP blocking defaults saved", "success");
    loadSecurityOverview();
  } catch (err) { setStatus(err.message); }
}

function renderManualIPLists(lists) {
  els.manualIPListList.innerHTML = lists.map((list, index) => {
    const entries = list.entries || [];
    const preview = entries.map((entry) => escapeHTML(entry)).join("<br>");
    return `<div class="manual-list-row" data-manual-list-index="${index}"><strong>${escapeHTML(list.name)}</strong><span class="manual-list-mode ${escapeHTML(list.mode)}">${list.mode === "allow" ? "Allow" : "Block"}</span><span>${escapeHTML(list.reference || "—")}</span><span class="ip-entry-preview" tabindex="0"><span>${Number(entries.length).toLocaleString()} IPs</span><span class="ip-entry-tooltip" role="tooltip">${preview || "No entries"}</span></span><span class="blocklist-actions"><button type="button" class="secondary" data-manual-list-action="edit">Edit</button><button type="button" class="danger" data-manual-list-action="remove">Remove</button></span></div>`;
  }).join("");
  els.manualIPListEmpty.hidden = lists.length > 0;
}

function openManualIPListDialog(index = -1) {
  const list = index >= 0 ? (settings?.manualIpLists || [])[index] : null;
  els.manualIPListIndex.value = list ? String(index) : "";
  els.manualIPListName.value = list?.name || "";
  els.manualIPListMode.value = list?.mode || "block";
  els.manualIPListReference.value = list?.reference || "";
  els.manualIPListEntries.value = (list?.entries || []).join("\n");
  els.manualIPListDialog.showModal();
  els.manualIPListName.focus();
}

async function saveManualIPList(event) {
  event.preventDefault();
  const lists = [...(settings?.manualIpLists || [])];
  const list = { name: els.manualIPListName.value.trim(), mode: els.manualIPListMode.value, reference: els.manualIPListReference.value.trim(), entries: protectionValues(els.manualIPListEntries.value) };
  const indexValue = els.manualIPListIndex.value;
  const index = indexValue === "" ? -1 : Number(indexValue);
  if (Number.isInteger(index) && index >= 0) lists[index] = list; else lists.push(list);
  const next = { ...settings, manualIpLists: lists };
  try {
    settings = await request("/api/settings", { method: "PUT", body: JSON.stringify(next) });
    renderManualIPLists(settings.manualIpLists || []);
    els.manualIPListDialog.close();
    setStatus("Manual IP list saved");
    showConfirmation("Manual IP list saved", "success");
    loadSecurityOverview();
  } catch (err) { setStatus(err.message); }
}

async function removeManualIPList(index) {
  const lists = [...(settings?.manualIpLists || [])];
  lists.splice(index, 1);
  try {
    settings = await request("/api/settings", { method: "PUT", body: JSON.stringify({ ...settings, manualIpLists: lists }) });
    renderManualIPLists(settings.manualIpLists || []);
    loadSecurityOverview();
  } catch (err) { setStatus(err.message); }
}

function openAssignIPDialog(address) {
  const lists = settings?.manualIpLists || [];
  if (!lists.length) {
    setStatus("Create a manual IP list before assigning an IP address");
    return;
  }
  els.assignIPValue.value = address;
  els.assignIPAddress.textContent = address;
  els.assignIPList.replaceChildren();
  lists.forEach((list, index) => {
    const option = document.createElement("option");
    option.value = String(index);
    option.textContent = `${list.name} (${list.mode === "allow" ? "Allow" : "Block"})`;
    els.assignIPList.append(option);
  });
  els.assignIPDialog.showModal();
  els.assignIPList.focus();
}

async function assignIPToManualList(event) {
  event.preventDefault();
  const address = els.assignIPValue.value.trim();
  const index = Number(els.assignIPList.value);
  const lists = [...(settings?.manualIpLists || [])].map((list) => ({ ...list, entries: [...(list.entries || [])] }));
  if (!address || !Number.isInteger(index) || !lists[index]) return;
  if (!lists[index].entries.includes(address)) lists[index].entries.push(address);
  try {
    settings = await request("/api/settings", { method: "PUT", body: JSON.stringify({ ...settings, manualIpLists: lists }) });
    renderManualIPLists(settings.manualIpLists || []);
    els.assignIPDialog.close();
    setStatus(`IP address added to ${lists[index].name}`);
    showConfirmation("IP address added to manual list", "success");
    loadSecurityOverview();
  } catch (err) { setStatus(err.message); }
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
    const [loadedSettings, providers] = await Promise.all([request("/api/settings"), request("/api/auth-providers")]);
    settings = loadedSettings;
    const protection = settings.webProtection || {};
    renderExternalBlocklists(settings.externalBlocklists || []);
    renderManualIPLists(settings.manualIpLists || []);
    els.webProtectionEnabled.checked = !!protection.enabled;
    els.webProtectionCountryMode.value = protection.countryMode || "block";
    setWebProtectionCountryMode(els.webProtectionCountryMode.value);
    selectedProtectionCountries = protection.blockedCountries || [];
    try { await loadProtectionCountries(); } catch (_) { protectionCountries = []; }
    renderSelectedProtectionCountries();
    els.settingsUsername.value = settings.username || "admin";
    els.settingsPassword.value = "";
    els.settingsPasswordConfirm.value = "";
    els.settingsOIDCEnabled.checked = !!settings.oidc?.enabled;
    els.settingsOIDCIssuer.value = settings.oidc?.issuerUrl || "";
    els.settingsOIDCClientID.value = settings.oidc?.clientId || "";
    els.settingsOIDCClientSecret.value = "";
    els.settingsOIDCRedirect.value = settings.oidc?.redirectUrl || "";
    els.settingsOIDCScopes.value = settings.oidc?.scopes || "openid profile email";
    els.settingsAccessEnabled.checked = !!providers.oidc?.enabled;
    els.settingsAccessName.value = providers.oidc?.name || "OIDC";
    els.settingsAccessIssuer.value = providers.oidc?.issuerUrl || "";
    els.settingsAccessClientID.value = providers.oidc?.clientId || "";
    els.settingsAccessClientSecret.value = "";
    els.settingsAccessScopes.value = providers.oidc?.scopes || "openid profile email";
    els.settingsAccessGateway.value = providers.oidc?.gatewayUrl || "";
    els.settingsWebHost.value = settings.webInterface?.host || "";
    els.settingsWebTLSEnabled.checked = !!settings.webInterface?.tlsEnabled;
    els.settingsLogRetention.value = settings.logRetention || 100;
    els.settingsCaddyMode.value = settings.caddyMode || "file";
    els.settingsCaddyAPIURL.value = settings.caddyApiUrl || "";
    renderCertificatesView();
    renderIssuerOptions();
    renderHostLists();
    els.settingsWebACME.value = settings.webInterface?.acmeIssuerId || "";
    els.settingsAccessACME.value = providers.oidc?.acmeIssuerId || "";
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
    webProtection: settings?.webProtection || {},
    externalBlocklists: settings?.externalBlocklists || [],
  };
  try {
    settings = await request("/api/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    await request("/api/auth-providers", {
      method: "PUT",
      body: JSON.stringify({
        oidc: {
          id: "oidc",
          name: els.settingsAccessName.value,
          enabled: els.settingsAccessEnabled.checked,
          issuerUrl: els.settingsAccessIssuer.value,
          clientId: els.settingsAccessClientID.value,
          clientSecret: els.settingsAccessClientSecret.value,
          scopes: els.settingsAccessScopes.value,
          gatewayUrl: els.settingsAccessGateway.value,
          acmeIssuerId: els.settingsAccessACME.value,
        },
      }),
    });
    els.settingsAccessClientSecret.value = "";
    els.settingsPassword.value = "";
    els.settingsPasswordConfirm.value = "";
    syncSettingsPasswordConfirmation();
    setStatus("Settings saved");
    showConfirmation("Settings saved");
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
    showConfirmation("Root CA uploaded");
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
    showConfirmation(message);

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
  updateSortHeaders();
  const tlsSites = sortedCertificates();
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

function sortedCertificates() {
  const sort = auxiliarySort.certificates;
  const direction = sort.direction === "desc" ? -1 : 1;
  const filters = tableFilters.certificates;
  const tlsSites = sites.filter((site) => {
    if (!site.tlsMode || site.tlsMode === "off") return false;
    const issuer = certificateIssuerName(site).toLowerCase();
    const expiry = site.certificateExpiresAt ? "known" : "unknown";
    const status = site.enabled ? "active" : "disabled";
    return (!filters.issuer || issuer.includes(filters.issuer))
      && (filters.expires === "all" || expiry === filters.expires)
      && (filters.status === "all" || status === filters.status);
  });
  return tlsSites.map((site, index) => ({ site, index })).sort((left, right) => {
    let result;
    if (sort.key === "expires") {
      result = certificateExpiryValue(left.site) - certificateExpiryValue(right.site);
    } else {
      const value = (site) => {
        if (sort.key === "issuer") return certificateIssuerName(site);
        if (sort.key === "status") return site.enabled ? "Active" : "Disabled";
        return site.address || "";
      };
      result = hostCollator.compare(value(left.site), value(right.site));
    }
    return result === 0 ? left.index - right.index : result * direction;
  }).map(({ site }) => site);
}

function certificateExpiryValue(site) {
  const value = new Date(site.certificateExpiresAt || 0).getTime();
  return Number.isNaN(value) ? 0 : value;
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

function certificateProviderName(site) {
  if (!site?.tlsMode || site.tlsMode === "off") return "None";
  if (site.tlsMode === "internal") return "Caddy Internal";
  return certificateIssuerName(site);
}

function renderIssuerOptions() {
  const current = els.acmeIssuer.value;
  const settingsCurrent = els.settingsWebACME.value;
  const accessCurrent = els.settingsAccessACME.value;
  els.acmeIssuer.innerHTML = "";
  els.settingsWebACME.innerHTML = "";
  els.settingsAccessACME.innerHTML = "";
  for (const issuer of settings?.acmeIssuers || []) {
    const option = document.createElement("option");
    option.value = issuer.id;
    option.textContent = issuer.name;
    els.acmeIssuer.append(option);
    els.settingsWebACME.append(option.cloneNode(true));
    els.settingsAccessACME.append(option.cloneNode(true));
  }
  els.acmeIssuer.value = current;
  els.settingsWebACME.value = settingsCurrent;
  els.settingsAccessACME.value = accessCurrent;
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
  renderHostFilterOptions();
  updateSortHeaders();
  renderSiteList(els.siteList, true, "web-hosts");
}

function updateSortHeaders() {
  document.querySelectorAll("[data-sort-table]").forEach((header) => {
    const sort = hostSort[header.dataset.sortTable] || auxiliarySort[header.dataset.sortTable];
    if (!sort) return;
    header.querySelectorAll(".sort-button").forEach((button) => {
      const active = button.dataset.sort === sort.key;
      button.dataset.direction = active ? sort.direction : "";
      button.setAttribute("aria-label", `${button.textContent}: ${active ? (sort.direction === "asc" ? "sorted ascending" : "sorted descending") : "not sorted"}`);
      button.parentElement.setAttribute("aria-sort", active ? (sort.direction === "asc" ? "ascending" : "descending") : "none");
    });
  });
}

function renderSortedTable(table) {
  updateSortHeaders();
  if (table === "dashboard" || table === "web-hosts") {
    renderHostLists();
  } else if (table === "certificates") {
    renderIssuedCertificates();
  } else if (table === "service-logs") {
    renderServiceLogs(latestServiceLogs, latestServiceLogsAvailable);
  } else if (table === "site-logs") {
    renderLogs(latestSiteLogs);
  } else if (table === "oidc-logs") {
    renderOIDCLogs(latestOIDCLogs);
  } else if (table === "protection-events") {
    renderProtectionEvents(latestProtectionEvents);
  }
}

function logMatchesFilters(entry, table) {
  const filters = tableFilters[table];
  if (!filters) return true;
  const values = {
    type: String(entry.action || "service").replaceAll("_", " ").toLowerCase(),
    message: String(entry.message || "").toLowerCase(),
    method: String(entry.method || "").toLowerCase(),
    path: String(entry.path || entry.message || "").toLowerCase(),
    user: String(entry.username || entry.email || "").toLowerCase(),
    ip: String(entry.ip || "").toLowerCase(),
    site: String(entry.site || "").toLowerCase(),
    status: String(entry.status || "").toLowerCase(),
  };
  return Object.entries(filters).every(([key, filter]) => !filter || filter === "all" || values[key]?.includes(filter));
}

function sortedLogs(logs, table) {
  const sort = auxiliarySort[table];
  const direction = sort.direction === "desc" ? -1 : 1;
  return logs.filter((entry) => logMatchesFilters(entry, table)).map((entry, index) => ({ entry, index })).sort((left, right) => {
    let result;
    if (sort.key === "time") {
      const leftTime = new Date(left.entry.time || 0).getTime();
      const rightTime = new Date(right.entry.time || 0).getTime();
      result = (Number.isNaN(leftTime) ? 0 : leftTime) - (Number.isNaN(rightTime) ? 0 : rightTime);
    } else {
      const value = (entry) => {
        if (sort.key === "type") return entry.action || "service";
        if (sort.key === "method") return entry.method || "";
        if (sort.key === "message") return entry.message || "";
        if (sort.key === "path") return entry.path || entry.message || "";
        if (sort.key === "user") return entry.username || entry.email || "";
        if (sort.key === "ip") return entry.ip || "";
        if (sort.key === "site") return entry.site || "";
        return String(entry.status || "");
      };
      result = hostCollator.compare(value(left.entry), value(right.entry));
    }
    return result === 0 ? left.index - right.index : result * direction;
  }).map(({ entry }) => entry);
}

function authenticationForSite(site) {
  if (site.basicAuthEnabled) return { key: "basic", label: "Basic" };
  if (site.authEnabled) return { key: "oidc", label: "OIDC" };
  return { key: "disabled", label: "Disabled" };
}

function hostSortValue(site, key) {
  switch (key) {
    case "certificateProvider": return certificateProviderName(site);
    case "visibility": return visibilityForSite(site).label;
    case "protocol": return protocolForSite(site);
    case "mode": return site.mode === "static" ? "Static" : "Proxy";
    case "target": return site.mode === "static" ? site.root : site.upstream;
    case "upstreamTls": return site.mode === "proxy" ? (site.skipTlsVerify ? "Skipped" : "Verified") : "-";
    case "status": return site.enabled ? "Active" : "Inactive";
    case "auth": return authenticationForSite(site).label;
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

function renderHostFilterOptions() {
  const current = els.hostFilterCertificateProvider.value;
  const providers = [...new Set(sites.map(certificateProviderName))].sort((left, right) => hostCollator.compare(left, right));
  els.hostFilterCertificateProvider.innerHTML = "";
  const all = document.createElement("option");
  all.value = "all";
  all.textContent = "All";
  els.hostFilterCertificateProvider.append(all);
  for (const provider of providers) {
    const option = document.createElement("option");
    option.value = provider.toLowerCase();
    option.textContent = provider;
    els.hostFilterCertificateProvider.append(option);
  }
  els.hostFilterCertificateProvider.value = [...els.hostFilterCertificateProvider.options].some((option) => option.value === current) ? current : "all";
  hostFilters.certificateProvider = els.hostFilterCertificateProvider.value;
}

function siteMatchesFilters(site, filters) {
  const protocol = protocolForSite(site).toLowerCase();
  const provider = certificateProviderName(site).toLowerCase();
  const visibility = visibilityForSite(site).label.toLowerCase();
  const mode = site.mode === "static" ? "static" : "proxy";
  const upstreamTls = site.mode !== "proxy" ? "not-applicable" : site.skipTlsVerify ? "skipped" : "verified";
  const status = site.enabled ? "active" : "inactive";
  const auth = authenticationForSite(site).key;
  const comment = (site.comment || "").toLowerCase();
  return (filters.protocol === "all" || protocol === filters.protocol)
    && (!filters.certificateProvider || filters.certificateProvider === "all" || provider.includes(filters.certificateProvider))
    && (filters.visibility === "all" || visibility === filters.visibility)
    && (filters.mode === "all" || mode === filters.mode)
    && (filters.upstreamTls === "all" || upstreamTls === filters.upstreamTls)
    && (filters.status === "all" || status === filters.status)
    && (filters.auth === "all" || auth === filters.auth)
    && (!filters.comment || comment.includes(filters.comment));
}

function filteredSitesFor(table) {
  const sorted = sortedSitesFor(table);
  if (table === "web-hosts") return sorted.filter((site) => siteMatchesFilters(site, hostFilters));
  if (table === "dashboard") return sorted.filter((site) => siteMatchesFilters(site, tableFilters.dashboard));
  return sorted;
}

function applyHostFilters() {
  hostFilters.protocol = els.hostFilterProtocol.value;
  hostFilters.certificateProvider = els.hostFilterCertificateProvider.value;
  hostFilters.visibility = els.hostFilterVisibility.value;
  hostFilters.mode = els.hostFilterMode.value;
  hostFilters.upstreamTls = els.hostFilterUpstreamTLS.value;
  hostFilters.status = els.hostFilterStatus.value;
  hostFilters.auth = els.hostFilterAuth.value;
  hostFilters.comment = els.hostFilterComment.value.trim().toLowerCase();
  renderSiteList(els.siteList, true, "web-hosts");
}

function resetHostFilters() {
  document.querySelectorAll("select.column-filter").forEach((filter) => { filter.value = "all"; });
  els.hostFilterComment.value = "";
  applyHostFilters();
}

function applyTableFilter(event) {
  const input = event.currentTarget;
  const filters = tableFilters[input.dataset.filterTable];
  if (!filters) return;
  filters[input.dataset.filterKey] = input.value.trim().toLowerCase();
  renderSortedTable(input.dataset.filterTable);
}

function resetTableFilters(event) {
  const table = event.currentTarget.dataset.resetFilters;
  const filters = tableFilters[table];
  if (!filters) return;
  document.querySelectorAll("[data-filter-table=" + table + "]").forEach((input) => {
    input.value = input.tagName === "SELECT" ? "all" : "";
  });
  for (const key of Object.keys(filters)) filters[key] = "";
  if (table === "dashboard") Object.assign(filters, { protocol: "all", visibility: "all", mode: "all", upstreamTls: "all", status: "all", auth: "all" });
  if (table === "certificates") Object.assign(filters, { expires: "all", status: "all" });
  if (table === "site-logs") filters.method = "all";
  renderSortedTable(table);
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

  const visibleSites = filteredSitesFor(table);
  if (table === "web-hosts") {
    els.hostFilterSummary.textContent = visibleSites.length + " of " + sites.length + " websites";
    els.hostFilterReset.disabled = Object.entries(hostFilters).every(([key, value]) => key === "comment" ? value === "" : value === "all");
  }
  if (!visibleSites.length) {
    const empty = document.createElement("div");
    empty.className = "site-row empty filtered-empty";
    empty.textContent = "No websites match the selected filters.";
    container.append(empty);
    return;
  }

  for (const site of visibleSites) {
    const row = document.createElement("div");
    row.className = `site-row ${editable ? "site-row-editable" : "site-row-dashboard"}`;
    row.innerHTML = editable
      ? `
          <span class="badge"></span>
          <strong></strong>
          <span class="certificate-provider"></span>
          <span class="badge"></span>
          <span></span>
          <span class="target"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="target"></span>
          <div class="site-actions"></div>
        `
      : `
          <span class="badge"></span>
          <strong></strong>
          <span class="certificate-provider"></span>
          <span class="badge"></span>
          <span></span>
          <span class="target"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="badge"></span>
          <span class="target"></span>
        `;
    row.children[1].textContent = site.address;
    row.children[2].textContent = certificateProviderName(site);
    row.children[2].title = certificateProviderName(site);
    const visibility = visibilityForSite(site);
    row.children[3].textContent = visibility.label;
    row.children[3].classList.add(visibility.className);
    row.children[3].title = visibility.description;
    const protocol = protocolForSite(site);
    row.children[0].textContent = protocol;
    row.children[0].classList.toggle("secure", protocol === "https");
    row.children[0].classList.toggle("warn", protocol === "http");
    row.children[4].textContent = site.mode === "static" ? "Static" : "Proxy";
    row.children[5].textContent = site.mode === "static" ? site.root : site.upstream;
    row.children[6].textContent = site.mode === "proxy" ? (site.skipTlsVerify ? "Skipped" : "Verified") : "-";
    row.children[6].classList.toggle("secure", site.mode === "proxy" && !site.skipTlsVerify);
    row.children[6].classList.toggle("warn", site.mode === "proxy" && !!site.skipTlsVerify);
    row.children[6].classList.toggle("off", site.mode !== "proxy");
    row.children[7].textContent = site.enabled ? "Active" : "Inactive";
    row.children[7].classList.toggle("off", !site.enabled);
    const authentication = authenticationForSite(site);
    row.children[8].textContent = authentication.label;
    row.children[8].classList.toggle("secure", authentication.key !== "disabled");
    row.children[8].classList.toggle("off", authentication.key === "disabled");
    row.children[9].textContent = site.comment || "-";
    row.children[9].title = site.comment || "";
    if (editable) {
      const actions = row.children[10];
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
        rewriteRedirects: site.rewriteRedirects !== false,
        redirectOrigins: site.redirectOrigins || [],
        hstsEnabled: !!site.hstsEnabled,
        securityHeaderProfile: site.securityHeaderProfile || "",
        compressionProfile: site.compressionProfile || "",
        root: site.root || "",
        extraDirectives: site.extraDirectives || "",
        logsEnabled: !!site.logsEnabled,
        tlsMode: site.tlsMode || "off",
        acmeIssuerId: site.tlsMode === "acme" ? site.acmeIssuerId || "" : "",
        tlsMinVersion: site.tlsMinVersion || "",
        tlsMaxVersion: site.tlsMaxVersion || "",
        authEnabled: !!site.authEnabled,
        authProviderId: site.authEnabled ? "oidc" : "",
        basicAuthEnabled: !!site.basicAuthEnabled,
        basicAuthUsername: site.basicAuthUsername || "",
        basicAuthPassword: "",
        enabled: nextEnabled,
      }),
    });
    setStatus(`Website ${nextEnabled ? "activated" : "deactivated"}`);
    showConfirmation(nextEnabled ? "Website activated" : "Website deactivated");
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
  latestSiteLogs = logs;
  els.logList.innerHTML = "";
  els.siteLogToggle.hidden = true;
  updateSortHeaders();
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

  const sorted = sortedLogs(logs, "site-logs");
  if (!sorted.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No website log lines match the selected filters.";
    els.logList.append(empty);
    return;
  }
  const visibleLogs = siteLogsExpanded ? sorted : sorted.slice(0, LOG_PREVIEW_LIMIT);
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
  syncLogToggle(els.siteLogToggle, sorted.length, siteLogsExpanded);
}

function renderServiceLogs(logs, available = true) {
  latestServiceLogs = logs;
  latestServiceLogsAvailable = available;
  els.serviceLogList.innerHTML = "";
  els.serviceLogToggle.hidden = true;
  updateSortHeaders();
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

  const sorted = sortedLogs(logs, "service-logs");
  if (!sorted.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No service log lines match the selected filters.";
    els.serviceLogList.append(empty);
    return;
  }
  const visibleLogs = serviceLogsExpanded ? sorted : sorted.slice(0, LOG_PREVIEW_LIMIT);
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
  syncLogToggle(els.serviceLogToggle, sorted.length, serviceLogsExpanded);
}

async function loadOIDCLogs() {
  try {
    const data = await request("/api/logs?source=oidc");
    renderOIDCLogs(data.logs || []);
  } catch (err) { setStatus(err.message); }
}

function renderOIDCLogs(logs) {
  latestOIDCLogs = logs;
  els.oidcLogList.innerHTML = "";
  els.oidcLogToggle.hidden = true;
  updateSortHeaders();
  if (!logs.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No OIDC authentication events yet.";
    els.oidcLogList.append(empty);
    return;
  }
  const sorted = sortedLogs(logs, "oidc-logs");
  if (!sorted.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No OIDC events match the selected filters.";
    els.oidcLogList.append(empty);
    return;
  }
  const visibleLogs = oidcLogsExpanded ? sorted : sorted.slice(0, LOG_PREVIEW_LIMIT);
  for (const entry of visibleLogs) {
    const row = document.createElement("div");
    row.className = "log-row oidc-log-row";
    for (let index = 0; index < 6; index += 1) row.append(document.createElement(index === 0 ? "time" : "span"));
    row.children[0].textContent = new Date(entry.time).toLocaleTimeString();
    row.children[0].title = new Date(entry.time).toLocaleString();
    row.children[1].textContent = String(entry.action || "event").replaceAll("_", " ");
    row.children[2].textContent = entry.username || entry.email || "-";
    row.children[2].title = entry.email || "";
    row.children[3].textContent = entry.ip || "unknown";
    row.children[4].textContent = entry.site || "-";
    row.children[4].title = entry.message || entry.site || "";
    row.children[5].textContent = entry.status || "-";
    row.children[5].className = "log-status badge";
    row.children[5].classList.toggle("off", !["success", "pending"].includes(String(entry.status || "").toLowerCase()));
    els.oidcLogList.append(row);
  }
  syncLogToggle(els.oidcLogToggle, sorted.length, oidcLogsExpanded);
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

function toggleOIDCLogsExpanded() {
  oidcLogsExpanded = !oidcLogsExpanded;
  loadOIDCLogs();
}

async function loadProtectionEvents() {
  try {
    const data = await request("/api/security-overview?events=all");
    renderProtectionEvents((data.events || []).map((event) => ({
      time: event.time,
      action: event.reason,
      site: event.site,
      ip: event.address,
      message: event.country || "Unknown",
    })));
  } catch (err) { setStatus(err.message); }
}

function renderProtectionEvents(events) {
  latestProtectionEvents = events;
  els.protectionEventList.innerHTML = "";
  els.protectionEventToggle.hidden = true;
  updateSortHeaders();
  if (!events.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No managed Web Protection events in retained access logs.";
    els.protectionEventList.append(empty);
    return;
  }
  const sorted = sortedLogs(events, "protection-events");
  if (!sorted.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No Web Protection events match the selected filters.";
    els.protectionEventList.append(empty);
    return;
  }
  const visibleEvents = protectionEventsExpanded ? sorted : sorted.slice(0, LOG_PREVIEW_LIMIT);
  for (const entry of visibleEvents) {
    const row = document.createElement("div");
    row.className = "log-row protection-log-row";
    for (let index = 0; index < 6; index += 1) row.append(document.createElement(index === 0 ? "time" : "span"));
    row.children[0].textContent = new Date(entry.time).toLocaleTimeString();
    row.children[0].title = new Date(entry.time).toLocaleString();
    row.children[1].textContent = entry.action || "Protection rule";
    row.children[2].textContent = entry.site || "-";
    row.children[3].textContent = entry.ip || "unknown";
    row.children[4].textContent = entry.message || "Unknown";
    const addButton = document.createElement("button");
    addButton.type = "button";
    addButton.className = "secondary assign-ip-button";
    addButton.textContent = "Add to list";
    addButton.addEventListener("click", () => openAssignIPDialog(entry.ip));
    row.children[5].append(addButton);
    els.protectionEventList.append(row);
  }
  syncLogToggle(els.protectionEventToggle, sorted.length, protectionEventsExpanded);
}

function toggleProtectionEventsExpanded() {
  protectionEventsExpanded = !protectionEventsExpanded;
  loadProtectionEvents();
}

function syncLogPolling() {
  stopLogPolling();
  const logsViewActive = document.querySelector("#view-logs").classList.contains("active");
  if (!logsViewActive) return;
  logPollTimer = window.setInterval(() => {
    loadServiceLogs();
    loadOIDCLogs();
    loadProtectionEvents();
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
  els.formTitle.textContent = site ? site.comment || site.address : "New Website";
  showSiteEditorTab("general");
  els.address.value = site?.address || "";
  els.comment.value = site?.comment || "";
  els.upstream.value = site?.upstream || "";
  els.skipTlsVerify.checked = !!site?.skipTlsVerify;
  els.rewriteRedirects.checked = site ? site.rewriteRedirects !== false : true;
  els.redirectOrigins.value = (site?.redirectOrigins || []).join("\n");
  els.hstsEnabled.checked = !!site?.hstsEnabled;
  els.securityHeaderProfile.value = site?.securityHeaderProfile || "";
  els.compressionProfile.value = site?.compressionProfile || "";
  els.root.value = site?.root || "";
  els.extra.value = site?.extraDirectives || "";
  els.enabled.checked = site?.enabled ?? true;
  els.logsEnabled.checked = site?.logsEnabled ?? true;
  els.tlsEnabled.checked = !!site && site?.tlsMode !== "off";
  els.tlsMinVersion.value = site?.tlsMinVersion || "";
  els.tlsMaxVersion.value = site?.tlsMaxVersion || "";
  els.siteAuthEnabled.checked = !!site?.authEnabled;
  els.basicAuthEnabled.checked = !!site?.basicAuthEnabled;
  els.basicAuthUsername.value = site?.basicAuthUsername || "";
  els.basicAuthPassword.value = "";
  els.protectionOverride.checked = !!site?.protectionOverride;
  els.siteProtectionEnabled.checked = !!site?.webProtection?.enabled;
  els.siteProtectionCountryMode.value = site?.webProtection?.countryMode || "block";
  els.siteProtectionCountries.value = (site?.webProtection?.blockedCountries || []).join(", ");
  els.siteProtectionBlockedIPs.value = (site?.webProtection?.blockedIps || []).join("\n");
  els.siteProtectionAllowedIPs.value = (site?.webProtection?.allowedIps || []).join("\n");
  syncProtectionOverride();
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
  syncSiteAuth();
  syncSecurityHeaderProfile();
  syncBasicAuth();
  syncEditorHostSummary();
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
  setFieldVisible(els.rewriteRedirectsRow, mode === "proxy");
  setFieldVisible(els.rewriteRedirectsHint, mode === "proxy");
  setFieldVisible(els.proxyBehaviorOverview, mode === "proxy");
  const showRedirectOrigins = mode === "proxy" && els.rewriteRedirects.checked;
  setFieldVisible(els.redirectOriginsRow, showRedirectOrigins);
  setFieldVisible(els.redirectOriginsHint, showRedirectOrigins);
  setFieldVisible(els.rootRow, mode === "static");
  els.upstream.required = mode === "proxy";
  els.root.required = mode === "static";
  els.upstream.disabled = mode !== "proxy";
  els.skipTlsVerify.disabled = mode !== "proxy";
  els.rewriteRedirects.disabled = mode !== "proxy";
  els.redirectOrigins.disabled = !showRedirectOrigins;
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
  syncProxyBehaviorOverview();
  syncEditorHostSummary();
}

function syncEditorHostSummary() {
  const address = els.address.value.trim();
  const comment = els.comment.value.trim();
  const mode = getMode();
  const tlsEnabled = els.tlsEnabled.checked;
  const enabled = els.enabled.checked;

  els.formTitle.textContent = comment || address || (editingSite ? "Website" : "New Website");
  els.editorHostAddress.textContent = address || "Address not configured";
  els.editorHostProtocol.textContent = tlsEnabled ? "HTTPS" : "HTTP";
  els.editorHostProtocol.className = `badge ${tlsEnabled ? "secure" : "warn"}`;
  els.editorHostMode.textContent = mode === "proxy" ? "Reverse Proxy" : mode === "static" ? "Static Files" : "Not configured";
  els.editorHostMode.className = `badge ${mode ? "public" : "off"}`;
  els.editorHostStatus.textContent = enabled ? "Active" : "Inactive";
  els.editorHostStatus.className = `badge ${enabled ? "secure" : "off"}`;
}

function syncProxyBehaviorOverview() {
  if (getMode() !== "proxy") return;
  const publicHost = els.address.value.trim() || "Public website host";
  const publicScheme = els.tlsEnabled.checked ? "https" : "http";
  const publicOrigin = `${publicScheme}://${publicHost}`;
  els.proxyOverviewHost.textContent = publicHost;
  els.proxyOverviewForwarded.textContent = `${publicHost} · ${publicScheme}`;

  const origins = [];
  let upstreamOrigin = "";
  try {
    const upstream = new URL(els.upstream.value.trim());
    upstreamOrigin = upstream.origin;
    origins.push(upstreamOrigin);
  } catch (_) {
    // The regular form validation reports an incomplete upstream.
  }
  els.proxyOverviewRequestRoute.textContent = upstreamOrigin ? `${publicOrigin} → ${upstreamOrigin}` : "Upstream not configured";
  els.redirectOrigins.value.split(/\r?\n/).map((origin) => origin.trim()).filter(Boolean).forEach((origin) => origins.push(origin.replace(/\/$/, "")));
  const uniqueOrigins = [...new Set(origins.map((origin) => origin.toLowerCase()))];
  els.proxyOverviewRedirectList.replaceChildren();
  if (!els.rewriteRedirects.checked || uniqueOrigins.length === 0) {
    const empty = document.createElement("span");
    empty.className = "muted";
    empty.textContent = els.rewriteRedirects.checked ? "No mapping available" : "Redirect rewriting disabled";
    els.proxyOverviewRedirectList.append(empty);
    return;
  }
  uniqueOrigins.forEach((origin) => {
    const rule = document.createElement("code");
    rule.textContent = `${origin} → ${publicOrigin}`;
    els.proxyOverviewRedirectList.append(rule);
  });
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
  els.hstsEnabled.disabled = !enabled;
  els.hstsEnabledRow.classList.toggle("disabled", !enabled);
  els.hstsEnabledHint.classList.toggle("disabled", !enabled);
  els.acmeIssuerRow.hidden = !enabled;
  els.acmeIssuer.required = enabled;
  els.acmeIssuer.disabled = !enabled;
  els.tlsMinVersion.disabled = !enabled;
  els.tlsMaxVersion.disabled = !enabled;
  els.tlsControlsHint.classList.toggle("disabled", !enabled);
  if (!enabled) {
    els.acmeIssuer.value = "";
    els.hstsEnabled.checked = false;
    els.siteAuthEnabled.checked = false;
    els.basicAuthEnabled.checked = false;
    syncBasicAuthFields();
  } else if (!els.acmeIssuer.value && els.acmeIssuer.options.length > 0) {
    els.acmeIssuer.value = els.acmeIssuer.options[0].value;
  }
  syncProxyBehaviorOverview();
  syncEditorHostSummary();
}

function syncSecurityHeaderProfile() {
  const descriptions = {
    standard: "Adds nosniff, strict-origin referrer policy, SAMEORIGIN framing, and removes the Server header.",
    strict: "Adds a restrictive Content Security Policy, DENY framing, disabled browser permissions, no-referrer, nosniff, and removes the Server header. This can break applications that load external or inline resources.",
  };
  els.securityHeaderProfileHint.textContent = descriptions[els.securityHeaderProfile.value] || "No managed security headers. Existing upstream headers remain unchanged.";
}

function syncSiteAuth() {
  const enabled = els.siteAuthEnabled.checked;
  if (enabled) {
    els.basicAuthEnabled.checked = false;
    if (!els.tlsEnabled.checked) {
      els.tlsEnabled.checked = true;
      syncTLSMode();
    }
  }
  syncBasicAuthFields();
}

function syncBasicAuth() {
  if (els.basicAuthEnabled.checked) {
    els.siteAuthEnabled.checked = false;
    if (!els.tlsEnabled.checked) {
      els.tlsEnabled.checked = true;
      syncTLSMode();
    }
  }
  syncBasicAuthFields();
}

function syncBasicAuthFields() {
  const enabled = els.basicAuthEnabled.checked;
  setFieldVisible(els.basicAuthUsernameRow, enabled);
  setFieldVisible(els.basicAuthPasswordRow, enabled);
  setFieldVisible(els.basicAuthHint, enabled);
  els.basicAuthUsername.disabled = !enabled;
  els.basicAuthPassword.disabled = !enabled;
  els.basicAuthUsername.required = enabled;
  els.basicAuthPassword.required = enabled && !editingSite?.basicAuthEnabled;
  els.basicAuthPassword.placeholder = editingSite?.basicAuthEnabled ? "Leave empty to keep current password" : "At least 8 characters";
}

function syncProtectionOverride() {
  els.protectionOverrideFields.hidden = !els.protectionOverride.checked;
}

async function saveSite(event) {
  event.preventDefault();
  const previousSite = editingSite ? { ...editingSite } : null;
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
    rewriteRedirects: mode === "proxy" && els.rewriteRedirects.checked,
    redirectOrigins: mode === "proxy" && els.rewriteRedirects.checked
      ? els.redirectOrigins.value.split(/\r?\n/).map((origin) => origin.trim()).filter(Boolean)
      : [],
    hstsEnabled: els.tlsEnabled.checked && els.hstsEnabled.checked,
    securityHeaderProfile: els.securityHeaderProfile.value,
    compressionProfile: els.compressionProfile.value,
    root: els.root.value,
    extraDirectives: els.extra.value,
    logsEnabled: els.logsEnabled.checked,
    tlsMode: els.tlsEnabled.checked ? "acme" : "off",
    acmeIssuerId: els.tlsEnabled.checked ? els.acmeIssuer.value : "",
    tlsMinVersion: els.tlsEnabled.checked ? els.tlsMinVersion.value : "",
    tlsMaxVersion: els.tlsEnabled.checked ? els.tlsMaxVersion.value : "",
    enabled: els.enabled.checked,
    authEnabled: els.siteAuthEnabled.checked,
    authProviderId: els.siteAuthEnabled.checked ? "oidc" : "",
    basicAuthEnabled: els.basicAuthEnabled.checked,
    basicAuthUsername: els.basicAuthEnabled.checked ? els.basicAuthUsername.value : "",
    basicAuthPassword: els.basicAuthEnabled.checked ? els.basicAuthPassword.value : "",
    protectionOverride: els.protectionOverride.checked,
    webProtection: els.protectionOverride.checked ? {
      enabled: els.siteProtectionEnabled.checked,
      countryMode: els.siteProtectionCountryMode.value,
      blockedCountries: protectionValues(els.siteProtectionCountries.value),
      blockedIps: protectionValues(els.siteProtectionBlockedIPs.value),
      allowedIps: protectionValues(els.siteProtectionAllowedIPs.value),
    } : {},
  };
  const id = els.id.value;
  try {
    const saved = await request(id ? `/api/sites/${id}` : "/api/sites", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload),
    });
    els.basicAuthPassword.value = "";
    await loadSites();
    await loadLogs();
    const savedID = saved?.id || id;
    const refreshedSite = sites.find((site) => site.id === savedID) || saved || { ...payload, id: savedID };
    editingSite = { ...refreshedSite };
    els.id.value = refreshedSite.id || savedID || "";
    els.formTitle.textContent = "Edit Website";
    els.delete.hidden = !els.id.value;
    if (payload.tlsMode !== "acme") {
      setStatus("Website saved · TLS disabled");
      showConfirmation("Website saved · TLS disabled", "success");
      return;
    }
    if (!payload.enabled) {
      setStatus("Website saved · TLS configured, website inactive");
      showConfirmation("TLS configured · Website inactive", "pending");
      return;
    }
    if (refreshedSite.certificateExpiresAt) {
      const message = `TLS certificate active until ${formatCertificateExpiry(refreshedSite.certificateExpiresAt)}`;
      setStatus(message);
      showConfirmation(message, "success");
      return;
    }

    if (shouldOpenACMEStatus(previousSite, payload, refreshedSite)) showACMEStatus(refreshedSite);
    setStatus(`Checking TLS certificate for ${refreshedSite.address}...`);
    showConfirmation("Website saved · Checking TLS certificate…", "pending");
    const verifiedSite = await waitForCertificate(refreshedSite.id, refreshedSite.address);
    if (verifiedSite) {
      editingSite = { ...verifiedSite };
      setACMEResult("success", `Certificate issued · Valid until ${formatCertificateExpiry(verifiedSite.certificateExpiresAt)}`);
      await loadSites();
      const message = `TLS certificate issued successfully for ${verifiedSite.address}`;
      setStatus(message);
      showConfirmation(message, "success");
    } else {
      setACMEResult("error", "Certificate not issued within 30 seconds. Review the Caddy activity below.");
      const message = `TLS certificate was not issued for ${refreshedSite.address}`;
      setStatus(message);
      showConfirmation(message, "error");
    }
  } catch (err) {
    setStatus(err.message);
    setACMEResult("error", err.message);
    showConfirmation(`Save failed · ${err.message}`, "error");
  }
}

async function waitForCertificate(siteID, address, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const data = await request("/api/sites");
    const candidate = (data.sites || []).find((site) => site.id === siteID)
      || (data.sites || []).find((site) => normalizeAddress(site.address) === normalizeAddress(address));
    if (candidate?.certificateExpiresAt) return candidate;
    await new Promise((resolve) => window.setTimeout(resolve, 2000));
  }
  return null;
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
  setACMEResult("pending", "Checking whether Caddy issued and stored the certificate…");
  els.acmeDialog.showModal();
  renderACMEActivityLogs([], true);
  stopACMEStatusPolling();
  loadServiceLogs();
  acmeDialogTimer = window.setInterval(loadServiceLogs, 2000);
}

function setACMEResult(kind, message) {
  if (!els.acmeResult) return;
  els.acmeResult.className = `certificate-check ${kind}`;
  els.acmeResult.textContent = message;
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
    await loadSites();    await loadLogs();
    showConfirmation("Website deleted");
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

function visibilityForSite(site) {
  const host = hostnameFromSiteAddress(site?.address);
  const internal = isInternalHostname(host);
  return internal
    ? { label: "Internal", className: "internal", description: "Inferred from a local hostname or private IP address" }
    : { label: "Public", className: "public", description: "Inferred from a public hostname or IP address" };
}

function hostnameFromSiteAddress(address) {
  const value = String(address || "").trim().toLowerCase();
  if (!value || value.startsWith(":")) return "";
  try {
    return new URL(value.includes("://") ? value : `http://${value}`).hostname
      .replace(/^\[|\]$/g, "")
      .replace(/^\*\./, "")
      .replace(/\.$/, "");
  } catch (_) {
    return value.replace(/^\*\./, "").replace(/:\d+$/, "").replace(/\.$/, "");
  }
}

function isInternalHostname(host) {
  if (!host || host === "localhost") return true;
  if (host.includes(":")) return isInternalIPAddress(host);
  if (isInternalIPAddress(host)) return true;
  if (!host.includes(".")) return true;
  const internalSuffixes = [
    ".local", ".localdomain", ".localhost", ".internal", ".lan", ".lds", ".home.arpa", ".test",
  ];
  return internalSuffixes.some((suffix) => host.endsWith(suffix));
}

function isInternalIPAddress(host) {
  const ipv4 = host.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/);
  if (ipv4) {
    const octets = ipv4.slice(1).map(Number);
    if (octets.some((octet) => octet > 255)) return false;
    const [first, second] = octets;
    return first === 0 || first === 10 || first === 127
      || (first === 100 && second >= 64 && second <= 127)
      || (first === 169 && second === 254)
      || (first === 172 && second >= 16 && second <= 31)
      || (first === 192 && second === 168);
  }
  const ipv6 = host.toLowerCase();
  return ipv6 === "::" || ipv6 === "::1" || ipv6.startsWith("fc") || ipv6.startsWith("fd")
    || /^fe[89ab]/.test(ipv6);
}

function setStatus(message) {
  const text = String(message || "");
  els.status.textContent = text.length > 96 ? `${text.slice(0, 93)}...` : text;
  els.status.title = text;
}

let confirmationTimer = null;

function showConfirmation(message, kind = "success") {
  window.clearTimeout(confirmationTimer);
  els.saveConfirmation.textContent = String(message || "Saved");
  els.saveConfirmation.classList.remove("success", "pending", "error");
  els.saveConfirmation.classList.add(kind);
  els.saveConfirmation.hidden = false;
  requestAnimationFrame(() => els.saveConfirmation.classList.add("visible"));
  confirmationTimer = window.setTimeout(() => {
    els.saveConfirmation.classList.remove("visible");
    window.setTimeout(() => { els.saveConfirmation.hidden = true; }, 180);
  }, kind === "error" ? 7000 : 4200);
}


async function loadGeoMap() {
  try {
    const data = await request("/api/geo-map");
    renderGeoMap(data);
  } catch (err) {
    renderGeoMap({ available: false, message: err.message, locations: [], topIps: [] });
  }
}

function renderGeoMap(data = {}) {
  const locations = Array.isArray(data.locations) ? data.locations : [];
  const topIPs = Array.isArray(data.topIps) ? data.topIps : [];
  els.geoMap.innerHTML = "";
  els.geoMapDetails.innerHTML = "";
  els.geoMapDetails.hidden = true;
  latestGeoTopIPs = topIPs;
  syncGeoTopHostOptions();
  renderTopIPs();

  if (!data.available) {
    els.geoMapSummary.textContent = "GeoLite2 database missing";
    appendGeoMapEmpty(data.message || "GeoLite2 City database is not configured.");
    return;
  }
  const requestCount = Number(data.requests || 0);
  els.geoMapSummary.textContent = `${requestCount} public request${requestCount === 1 ? "" : "s"} · ${locations.length} location${locations.length === 1 ? "" : "s"}`;
  if (!locations.length) {
    appendGeoMapEmpty("No public website accesses found in the retained logs.");
    return;
  }

  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("viewBox", "0 0 1000 500");
  svg.setAttribute("aria-hidden", "true");
  svg.innerHTML = `<g class="geo-graticule"><path d="M0 125h1000M0 250h1000M0 375h1000M250 0v500M500 0v500M750 0v500"/></g>
    <image class="geo-land-map" href="/world-map.svg" x="0" y="0" width="1000" height="500" preserveAspectRatio="none"/>`;
  for (const location of locations) {
    const x = ((Number(location.longitude) + 180) / 360) * 1000;
    const y = ((90 - Number(location.latitude)) / 180) * 500;
    const marker = document.createElementNS(svg.namespaceURI, "circle");
    marker.setAttribute("cx", String(x));
    marker.setAttribute("cy", String(y));
    marker.setAttribute("r", "8");
    marker.setAttribute("class", "geo-marker");
    marker.setAttribute("tabindex", "0");
    marker.setAttribute("role", "button");
    marker.setAttribute("aria-label", `${geoLocationLabel(location)}, ${location.count} requests`);
    marker.addEventListener("click", () => showGeoLocation(location, marker));
    marker.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        showGeoLocation(location, marker);
      }
    });
    svg.append(marker);
  }
  els.geoMap.append(svg);
  addGeoMapZoomControls(svg);
}

function addGeoMapZoomControls(svg) {
  const minZoom = 1, maxZoom = 4, zoomStep = 0.25;
  let zoom = minZoom, centerX = 500, centerY = 250, dragStart = null;
  const controls = document.createElement("div");
  controls.className = "geo-zoom-controls";
  controls.setAttribute("aria-label", "Map zoom controls");
  const zoomOut = document.createElement("button");
  zoomOut.type = "button"; zoomOut.textContent = "−"; zoomOut.setAttribute("aria-label", "Zoom out");
  const slider = document.createElement("input");
  slider.type = "range"; slider.min = String(minZoom); slider.max = String(maxZoom);
  slider.step = String(zoomStep); slider.value = String(zoom); slider.setAttribute("aria-label", "Map zoom level");
  const zoomIn = document.createElement("button");
  zoomIn.type = "button"; zoomIn.textContent = "+"; zoomIn.setAttribute("aria-label", "Zoom in");
  const reset = document.createElement("button");
  reset.type = "button"; reset.className = "geo-zoom-reset"; reset.textContent = "Reset";
  reset.setAttribute("aria-label", "Reset map zoom and position");
  const clampCenter = () => {
    const width = 1000 / zoom, height = 500 / zoom;
    centerX = Math.min(1000 - width / 2, Math.max(width / 2, centerX));
    centerY = Math.min(500 - height / 2, Math.max(height / 2, centerY));
  };
  const render = () => {
    clampCenter();
    const width = 1000 / zoom, height = 500 / zoom;
    svg.setAttribute("viewBox", `${centerX - width / 2} ${centerY - height / 2} ${width} ${height}`);
    svg.querySelectorAll(".geo-marker").forEach((marker) => marker.setAttribute("r", String(8 / zoom)));
    slider.value = String(zoom);
    slider.setAttribute("aria-valuetext", `${Math.round(zoom * 100)} percent`);
    zoomOut.disabled = zoom <= minZoom; zoomIn.disabled = zoom >= maxZoom;
    reset.disabled = zoom === minZoom && centerX === 500 && centerY === 250;
    svg.classList.toggle("zoomed", zoom > minZoom);
  };
  const setZoom = (nextZoom) => { zoom = Math.min(maxZoom, Math.max(minZoom, Number(nextZoom))); render(); };
  zoomOut.addEventListener("click", () => setZoom(zoom - zoomStep));
  zoomIn.addEventListener("click", () => setZoom(zoom + zoomStep));
  slider.addEventListener("input", () => setZoom(slider.value));
  reset.addEventListener("click", () => { zoom = minZoom; centerX = 500; centerY = 250; render(); });
  svg.addEventListener("wheel", (event) => { event.preventDefault(); setZoom(zoom + (event.deltaY < 0 ? zoomStep : -zoomStep)); }, { passive: false });
  svg.addEventListener("pointerdown", (event) => {
    if (zoom <= minZoom || event.button !== 0 || event.target.closest(".geo-marker")) return;
    dragStart = { x: event.clientX, y: event.clientY, centerX, centerY };
    svg.setPointerCapture(event.pointerId); svg.classList.add("dragging");
  });
  svg.addEventListener("pointermove", (event) => {
    if (!dragStart) return;
    const bounds = svg.getBoundingClientRect();
    centerX = dragStart.centerX - ((event.clientX - dragStart.x) / bounds.width) * (1000 / zoom);
    centerY = dragStart.centerY - ((event.clientY - dragStart.y) / bounds.height) * (500 / zoom);
    render();
  });
  const stopDragging = () => { dragStart = null; svg.classList.remove("dragging"); };
  svg.addEventListener("pointerup", stopDragging);
  svg.addEventListener("pointercancel", stopDragging);
  controls.append(zoomOut, slider, zoomIn, reset);
  els.geoMap.append(controls);
  render();
}

function syncGeoTopHostOptions() {
  const current = els.geoIPHost.value;
  const hosts = [...new Set(latestGeoTopIPs.flatMap((ip) => Array.isArray(ip.sites) ? ip.sites : []))]
    .sort((left, right) => hostCollator.compare(left, right));
  els.geoIPHost.innerHTML = "";
  const all = document.createElement("option");
  all.value = "all";
  all.textContent = "All hosts";
  els.geoIPHost.append(all);
  for (const host of hosts) {
    const option = document.createElement("option");
    option.value = host;
    option.textContent = host;
    els.geoIPHost.append(option);
  }
  els.geoIPHost.value = hosts.includes(current) ? current : "all";
}

function renderTopIPs() {
  els.geoTopIPList.innerHTML = "";
  const scope = els.geoIPScope.value || "all";
  const host = els.geoIPHost.value || "all";
  const limit = Number(els.geoIPLimit.value || 10);
  const filtered = latestGeoTopIPs.filter((ip) => {
    const scopeMatches = scope === "all" || ip.scope === scope;
    const hostMatches = host === "all" || (ip.sites || []).includes(host);
    return scopeMatches && hostMatches;
  });
  filtered.sort((left, right) => {
    const leftCount = host === "all" ? Number(left.count || 0) : Number(left.siteCounts?.[host] || 0);
    const rightCount = host === "all" ? Number(right.count || 0) : Number(right.siteCounts?.[host] || 0);
    return rightCount - leftCount || hostCollator.compare(left.address || "", right.address || "");
  });
  const visible = filtered.slice(0, limit);
  els.geoTopIPSummary.textContent = `${visible.length} of ${filtered.length} IP${filtered.length === 1 ? "" : "s"}`;

  if (!visible.length) {
    const empty = document.createElement("div");
    empty.className = "geo-sidebar-empty";
    empty.textContent = "No client IPs match these filters.";
    els.geoTopIPList.append(empty);
    return;
  }
  visible.forEach((ip, index) => {
    const displayCount = host === "all" ? Number(ip.count || 0) : Number(ip.siteCounts?.[host] || 0);
    const row = document.createElement("div");
    row.className = "geo-top-ip-row";
    const rank = document.createElement("span");
    rank.className = "geo-ip-rank";
    rank.textContent = String(index + 1);
    const content = document.createElement("div");
    const primary = document.createElement("div");
    primary.className = "geo-ip-primary";
    const country = document.createElement("span");
    country.className = "geo-ip-country";
    country.textContent = ip.scope === "internal" ? "Local network" : (ip.country || ip.countryCode || "Unknown country");
    if (ip.countryCode && ip.country) country.title = ip.countryCode;
    const address = document.createElement("code");
    address.textContent = ip.address || "Unknown IP";
    const requests = document.createElement("span");
    requests.className = "geo-request-count";
    requests.textContent = `${displayCount} request${displayCount === 1 ? "" : "s"}`;
    const badge = document.createElement("span");
    badge.className = `geo-scope-badge ${ip.scope === "internal" ? "internal" : "external"}`;
    badge.textContent = ip.scope === "internal" ? "Internal" : "External";
    const addButton = document.createElement("button");
    addButton.type = "button";
    addButton.className = "secondary assign-ip-button";
    addButton.textContent = "Add to list";
    addButton.addEventListener("click", () => openAssignIPDialog(ip.address || ""));
    primary.append(country, address, requests, badge, addButton);
    const meta = document.createElement("span");
    meta.textContent = host === "all" ? ((ip.sites || []).join(", ") || "Unknown website") : host;
    content.append(primary, meta);
    row.append(rank, content);
    els.geoTopIPList.append(row);
  });
}

function showGeoLocation(location, marker) {
  els.geoMap.querySelectorAll(".geo-marker").forEach((item) => item.classList.toggle("active", item === marker));
  els.geoMapDetails.innerHTML = "";
  els.geoMapDetails.hidden = false;
  const heading = document.createElement("div");
  heading.className = "geo-location-heading";
  const headingCopy = document.createElement("div");
  headingCopy.className = "geo-location-heading-copy";
  const title = document.createElement("strong");
  title.textContent = geoLocationLabel(location);
  const count = document.createElement("span");
  count.textContent = String(location.count || 0) + " requests";
  headingCopy.append(title, count);
  const close = document.createElement("button");
  close.type = "button";
  close.className = "geo-details-close";
  close.setAttribute("aria-label", "Close location details");
  close.textContent = "×";
  close.addEventListener("click", closeGeoLocation);
  heading.append(headingCopy, close);
  els.geoMapDetails.append(heading);
  for (const ip of location.ips || []) {
    const row = document.createElement("div");
    row.className = "geo-ip-row";
    const address = document.createElement("code");
    address.textContent = ip.address || "Unknown IP";
    const primary = document.createElement("div");
    primary.className = "geo-location-ip-primary";
    const requests = document.createElement("span");
    requests.className = "geo-request-count";
    requests.textContent = `${ip.count || 0} request${Number(ip.count || 0) === 1 ? "" : "s"}`;
    const meta = document.createElement("span");
    meta.textContent = (ip.sites || []).join(", ") || "Unknown website";
    primary.append(address, requests);
    row.append(primary, meta);
    els.geoMapDetails.append(row);
  }
}

function closeGeoLocation() {
  els.geoMap.querySelectorAll(".geo-marker").forEach((item) => item.classList.remove("active"));
  els.geoMapDetails.hidden = true;
  els.geoMapDetails.innerHTML = "";
}

function geoLocationLabel(location) {
  return [location.city, location.country || location.countryCode].filter(Boolean).join(", ") || "Unknown location";
}

function appendGeoMapEmpty(message) {
  const empty = document.createElement("div");
  empty.className = "geo-map-empty";
  empty.textContent = message;
  els.geoMap.append(empty);
}
