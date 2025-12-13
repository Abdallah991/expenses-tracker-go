import express, { Express } from "express";
import cors from "cors";
import dotenv from "dotenv";
import helmet from "helmet";
import morgan from "morgan";
import compression from "compression";
import { initDB } from "../config/database";
import { initJWT } from "../auth/jwt";
import { initEmailService } from "../services/email.service";
// import routes from "../routes";
import { errorHandler } from "../middleware/error.middleware";
import { requestIdMiddleware } from "../middleware/requestId.middleware";
import { env } from "../config/env";

dotenv.config();

export function createApp(): Express {
  const app: Express = express();

  // Initialize services
  try {
    initDB();
    initJWT();
    initEmailService();
  } catch (error) {
    console.error("❌ Failed to initialize services:", error);
    throw error;
  }

  // Security middleware (should be first)
  app.use(helmet());

  // CORS configuration
  app.use(cors());

  // Request ID for tracing (before other middleware that might log)
  app.use(requestIdMiddleware);

  // HTTP request logging
  if (env.NODE_ENV === "development") {
    app.use(morgan("dev")); // Colored output for development
  } else {
    app.use(morgan("combined")); // Standard Apache combined log format for production
  }

  // Response compression (gzip)
  app.use(compression());

  // Body parsing middleware
  app.use(express.json({ limit: "1mb" }));
  app.use(express.urlencoded({ extended: true, limit: "1mb" })); // For form data support

  // Routes, import routes after the Database is initialized
  const routes = require("../routes").default;
  app.use("/", routes);
  // Error handling middleware (must be last)
  app.use(errorHandler);

  return app;
}
