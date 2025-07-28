document.getElementById("apiCallBtn").addEventListener("click", async () => {
  try {
    const response = await fetch("http://127.0.0.1:8080", {
      method: "GET", // or "POST", "PUT", etc.
      headers: {
        // "Content-Type": "application/json",
        // "Authorization": "Bearer YOUR_TOKEN" // if needed
      },
    });

    const data = await response.json();
    console.log("API response:", data);
    alert("Success: " + JSON.stringify(data));
  } catch (error) {
    console.error("API call failed:", error);
    alert("API call failed!" + error);
  }
});
