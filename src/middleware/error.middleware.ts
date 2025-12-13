import { Request, Response, NextFunction } from "express";
import { AppError } from "../utils/errors.util";

export function errorHandler(
  err: Error | AppError,
  _req: Request,
  res: Response,
  next: NextFunction
): void {
  // If response already sent, delegate to default Express error handler
  if (res.headersSent) {
    return next(err);
  }

  // Handle known operational errors
  if (err instanceof AppError && err.isOperational) {
    res.status(err.statusCode).json({
      error: err.name,
      message: err.message,
    });
    return;
  }

  // Handle validation errors from express-validator
  if (err.name === "ValidationError" || err.message.includes("validation")) {
    res.status(400).json({
      error: "ValidationError",
      message: err.message,
    });
    return;
  }

  // Log unexpected errors
  console.error("Unexpected error:", err);

  // Send generic error response for unexpected errors
  res.status(500).json({
    error: "InternalServerError " + err,
    message: process.env.NODE_ENV === "production" ? "An unexpected error occurred" : err.message,
  });
}
