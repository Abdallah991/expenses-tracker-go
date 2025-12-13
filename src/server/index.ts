import { createApp } from "./app";
import { env } from "../config/env";

const app = createApp();

// Start the server
app.listen(env.PORT, () => {
  console.log(`✅ Starting server on port ${env.PORT}...`);
  console.log("🔐 Authentication system initialized");
  console.log("📧 Email service initialized");
});
