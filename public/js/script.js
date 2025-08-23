document.getElementById("apiCallBtnInstallation").addEventListener("click", async () => {
  try {
    const params = new URLSearchParams();
    params.append("timezone_offset_in_minutes", "180");
    params.append("device_os", "linux");
    params.append("client_type", "web");
    params.append("app_version", "1.0.0");
    params.append("locale", "ar");

    const response = await fetch("http://127.0.0.1:8080/api/v1/installation/create", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Accept-Language": "ar",
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: params.toString(),
    });

    const data = await response.json();
    console.log("API response:", data);
  } catch (error) {
    console.error("API call failed:", error);
    alert("API call failed!" + error);
  }
});

document.getElementById("logout_btn").addEventListener("click", async () => {
  try {
    const response = await fetch("http://127.0.0.1:8080/api/v1/auth/logout", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Accept-Language": "ar",
        "Content-Type": "application/x-www-form-urlencoded",
      },
    });

    const data = await response.json();
    console.log("API response:", data);
  } catch (error) {
    console.error("API call failed:", error);
    alert("API call failed!" + error);
  }
});

const params = new URLSearchParams(window.location.search);

params.forEach((value, key) => {
  const el = document.getElementById(key);
  if (el) {
    if (el.tagName === "IMG") {
      el.src = value; // Set image src
    } else {
      el.textContent = value; // Set text
    }
  }
});
