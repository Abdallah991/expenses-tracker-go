import dotenv from "dotenv";

dotenv.config();

export const env = {
  // Database
  DATABASE_URL: process.env.DATABASE_URL || "",

  // JWT
  JWT_SECRET: process.env.JWT_SECRET || "",
  JWT_ACCESS_EXPIRY: process.env.JWT_ACCESS_EXPIRY || "15m",
  JWT_REFRESH_EXPIRY: process.env.JWT_REFRESH_EXPIRY || "168h",

  // Email
  RESEND_API_KEY: process.env.RESEND_API_KEY || "",
  FROM_EMAIL: process.env.FROM_EMAIL || "",

  // Application
  APP_URL: process.env.APP_URL || "http://localhost:8080",
  PORT: parseInt(process.env.PORT || "8080", 10),
  NODE_ENV: process.env.NODE_ENV || "development",

  // Mobile
  MOBILE_DEEP_LINK_SCHEME: process.env.MOBILE_DEEP_LINK_SCHEME || "myexpenses://",
};

// Validate required environment variables
const requiredEnvVars: (keyof typeof env)[] = [
  "DATABASE_URL",
  "JWT_SECRET",
  "RESEND_API_KEY",
  "FROM_EMAIL",
];

for (const varName of requiredEnvVars) {
  if (!env[varName]) {
    throw new Error(`Missing required environment variable: ${varName}`);
  }
}
