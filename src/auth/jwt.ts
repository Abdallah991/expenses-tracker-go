import jwt from "jsonwebtoken";
import crypto from "crypto";
import dotenv from "dotenv";

dotenv.config();

export interface JWTClaims {
  userId: number;
  email: string;
  issuedAt?: number; // When the token was issued (iat)
  expiresAt?: number; // When the token expires (exp)
  notBefore?: number; // Token not valid before this time (nbf)
  issuer?: string; // Who issued the token (iss)
  subject?: string; // Subject of the token, typically user ID (sub)
}

let jwtSecret: string;
let accessExpiry: number; // in seconds
let refreshExpiry: number; // in seconds

export function initJWT(): void {
  jwtSecret = process.env.JWT_SECRET || "";
  if (!jwtSecret) {
    throw new Error("JWT_SECRET environment variable is required");
  }

  // Parse access token expiry (default: 15 minutes)
  const accessExpiryStr = process.env.JWT_ACCESS_EXPIRY || "15m";
  accessExpiry = parseDuration(accessExpiryStr);

  // Parse refresh token expiry (default: 7 days)
  const refreshExpiryStr = process.env.JWT_REFRESH_EXPIRY || "168h";
  refreshExpiry = parseDuration(refreshExpiryStr);
}

function parseDuration(duration: string): number {
  // Handle formats like "15m", "168h", "7d", "30s"
  const match = duration.match(/^(\d+)([smhd])$/);
  if (!match) {
    throw new Error(
      `Invalid duration format: ${duration}. Expected format: <number><unit> (e.g., 15m, 168h, 7d)`
    );
  }

  const value = parseInt(match[1], 10);
  const unit = match[2];

  switch (unit) {
    case "s":
      return value; // seconds
    case "m":
      return value * 60; // minutes to seconds
    case "h":
      return value * 3600; // hours to seconds
    case "d":
      return value * 86400; // days to seconds
    default:
      throw new Error(`Unknown duration unit: ${unit}. Supported units: s, m, h, d`);
  }
}

export function generateAccessToken(userId: number, email: string): string {
  const now = Math.floor(Date.now() / 1000);

  // Create payload with descriptive names for TypeScript
  const claims: JWTClaims = {
    userId: userId,
    email: email,
    issuedAt: now,
    expiresAt: now + accessExpiry,
    notBefore: now,
    issuer: "expenses-tracker",
    subject: userId.toString(),
  };

  // Map to standard JWT claim names for token signing
  const jwtPayload = {
    userId: claims.userId,
    email: claims.email,
    iat: claims.issuedAt,
    exp: claims.expiresAt,
    nbf: claims.notBefore,
    iss: claims.issuer,
    sub: claims.subject,
  };

  return jwt.sign(jwtPayload, jwtSecret, { algorithm: "HS256" });
}

export function generateRefreshToken(): string {
  return generateSecureToken(32);
}

export function validateAccessToken(tokenString: string): JWTClaims {
  try {
    // Decode token with standard JWT claim names
    const decoded = jwt.verify(tokenString, jwtSecret, {
      algorithms: ["HS256"],
    }) as any;

    // Map standard JWT claim names to descriptive TypeScript interface
    const claims: JWTClaims = {
      userId: decoded.userId,
      email: decoded.email,
      issuedAt: decoded.iat,
      expiresAt: decoded.exp,
      notBefore: decoded.nbf,
      issuer: decoded.iss,
      subject: decoded.sub,
    };

    return claims;
  } catch (err) {
    throw new Error("Invalid token");
  }
}

export function extractTokenFromHeader(authHeader: string): string {
  if (!authHeader) {
    throw new Error("authorization header is required");
  }

  const bearerPrefix = "Bearer ";
  if (!authHeader.startsWith(bearerPrefix)) {
    throw new Error("authorization header must start with 'Bearer '");
  }

  const token = authHeader.substring(bearerPrefix.length);
  if (!token) {
    throw new Error("token cannot be empty");
  }

  return token;
}

export function getTokenExpiry(): number {
  return accessExpiry;
}

export function getRefreshTokenExpiry(): number {
  return refreshExpiry;
}

export function generateSecureToken(length: number): string {
  const bytes = crypto.randomBytes(length);
  return bytes.toString("hex");
}
