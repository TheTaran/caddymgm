const els = {
  status: document.querySelector("#status"),
  siteList: document.querySelector("#site-list"),
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
};

let sites = [];

document.querySelector("#refresh").addEventListener("click", loadSites);
document.querySelector("#new-site").addEventListener("click", () => editSite());
document.querySelector("#cancel").addEventListener("click", () => (els.editor.hidden = true));
document.querySelector("#show-config").addEventListener("click", showConfig);
document.querySelector("#close-config").addEventListener("click", () => els.configDialog.close());
els.form.addEventListener("submit", saveSite);
els.delete.addEventListener("click", deleteSite);
document.querySelectorAll("input[name='mode']").forEach((input) => {
  input.addEventListener("change", syncMode);
});

loadSites();

async function loadSites() {
  setStatus("Lade Websites...");
  try {
    const data = await request("/api/sites");
    sites = data.sites || [];
    renderSites();
    renderMetrics();
    setStatus(`${sites.length} Website${sites.length === 1 ? "" : "s"} verwaltet`);
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

function renderSites() {
  els.siteList.innerHTML = "";
  if (!sites.length) {
    const empty = document.createElement("div");
    empty.className = "site-row empty";
    empty.textContent = "Noch keine Websites angelegt.";
    els.siteList.append(empty);
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
      <button class="secondary" type="button">Bearbeiten</button>
    `;
    row.children[0].textContent = site.address;
    row.children[1].textContent = site.mode === "static" ? "Static" : "Proxy";
    row.children[2].textContent = site.mode === "static" ? site.root : site.upstream;
    row.children[3].textContent = site.enabled ? "Aktiv" : "Inaktiv";
    row.children[3].classList.toggle("off", !site.enabled);
    row.children[4].addEventListener("click", () => editSite(site));
    els.siteList.append(row);
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
  } catch (err) {
    setStatus(err.message);
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
