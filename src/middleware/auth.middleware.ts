import { Request, Response, NextFunction } from "express";
import { validateAccessToken, extractTokenFromHeader } from "../auth/jwt";
import { getDB } from "../config/database";
import { AuthenticationError } from "../utils/errors.util";

export interface AuthenticatedRequest extends Request {
  user_id?: number;
  user_email?: string;
}

export async function requireAuth(
  req: AuthenticatedRequest,
  _res: Response,
  next: NextFunction
): Promise<void> {
  try {
    const authHeader = req.headers.authorization || "";
    const tokenString = extractTokenFromHeader(authHeader);
    const claims = validateAccessToken(tokenString);

    // Check if user still exists and is active
    const db = getDB();
    const result = await db.query(
      "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND email_verified = true)",
      [claims.userId]
    );

    const userExists = result.rows[0].exists;
    if (!userExists) {
      throw new AuthenticationError("User not found or not verified");
    }

    // Add user information to request
    req.user_id = claims.userId;
    req.user_email = claims.email;

    next();
  } catch (err) {
    next(err);
  }
}

export function getUserIDFromRequest(req: AuthenticatedRequest): number {
  if (!req.user_id) {
    throw new Error("user ID not found in request");
  }
  return req.user_id;
}

export function getUserEmailFromRequest(req: AuthenticatedRequest): string {
  if (!req.user_email) {
    throw new Error("user email not found in request");
  }
  return req.user_email;
}
