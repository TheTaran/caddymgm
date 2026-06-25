const form = document.querySelector("#login-form");
const errorBox = document.querySelector("#login-error");

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
    if (!response.ok) throw new Error(data.error || "Login fehlgeschlagen");
    window.location.assign("/");
  } catch (err) {
    errorBox.textContent = err.message;
    errorBox.hidden = false;
  }
});
