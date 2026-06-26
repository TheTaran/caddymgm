const form = document.querySelector("#login-form");
const errorBox = document.querySelector("#login-error");
const oidcButton = document.querySelector("#oidc-login");
const divider = document.querySelector("#login-divider");
const disabledBox = document.querySelector("#login-disabled");
const methodNote = document.querySelector("#login-method-note");

init();

async function init() {
  applyLoginModes(false, false, false);
  const params = new URLSearchParams(window.location.search);
  const error = params.get("error");
  if (error) {
    errorBox.textContent = friendlyError(error);
    errorBox.hidden = false;
  }

  try {
    const response = await fetch("/api/auth/config");
    const config = await response.json();
    if (!response.ok) throw new Error(config.error || "Could not load auth config");

    const authEnabled = !!config.authEnabled;
    const localEnabled = !!config.localAuthEnabled;
    const oidcEnabled = !!config.oidcAuthEnabled;

    if (!authEnabled) {
      applyLoginModes(false, false, true);
      disabledBox.hidden = false;
      methodNote.hidden = false;
      methodNote.textContent = "Open the interface directly. No login is required.";
      return;
    }

    applyLoginModes(localEnabled, oidcEnabled, false);

    if (!localEnabled && !oidcEnabled) {
      methodNote.hidden = false;
      methodNote.textContent = "Authentication is enabled, but no login method is available.";
      errorBox.textContent = "Enable local authentication or OIDC in the environment.";
      errorBox.hidden = false;
      return;
    }

    if (localEnabled && oidcEnabled) {
      methodNote.hidden = false;
      methodNote.textContent = "Use local admin access or continue with your identity provider.";
    } else if (localEnabled) {
      methodNote.hidden = false;
      methodNote.textContent = "Local admin login is enabled.";
    } else if (oidcEnabled) {
      methodNote.hidden = true;
      methodNote.textContent = "";
    }
  } catch (err) {
    errorBox.textContent = err.message;
    errorBox.hidden = false;
  }
}

function applyLoginModes(localEnabled, oidcEnabled, authDisabled) {
  const username = document.querySelector("#username");
  const password = document.querySelector("#password");

  form.hidden = !localEnabled || authDisabled;
  oidcButton.hidden = !oidcEnabled || authDisabled;
  divider.hidden = !(localEnabled && oidcEnabled) || authDisabled;

  username.required = !!localEnabled && !authDisabled;
  password.required = !!localEnabled && !authDisabled;
  username.disabled = !localEnabled || authDisabled;
  password.disabled = !localEnabled || authDisabled;
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  errorBox.hidden = true;

  const payload = {
    username: document.querySelector("#username").value,
    password: document.querySelector("#password").value,
  };

  try {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || "Login failed");
    window.location.assign("/");
  } catch (err) {
    errorBox.textContent = err.message;
    errorBox.hidden = false;
  }
});

oidcButton.addEventListener("click", () => {
  window.location.assign("/api/auth/oidc/start");
});

function friendlyError(code) {
  switch (code) {
    case "missing_oidc_response":
      return "OIDC login did not return the expected response.";
    case "oidc_not_available":
      return "OIDC is not available with the current configuration.";
    case "invalid_oidc_state":
      return "The OIDC login session is invalid or expired.";
    case "oidc_exchange_failed":
      return "The OIDC authorization code could not be exchanged.";
    case "missing_id_token":
      return "The identity provider did not return an ID token.";
    case "invalid_id_token":
      return "The ID token could not be verified.";
    case "invalid_oidc_claims":
      return "The ID token claims could not be read.";
    default:
      return "Login failed.";
  }
}
