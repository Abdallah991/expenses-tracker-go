import rateLimit from 'express-rate-limit';

// Helper function to get client IP
function getClientIP(req: any): string {
  // Check X-Forwarded-For header (for reverse proxies)
  const xff = req.headers['x-forwarded-for'];
  if (xff) {
    // X-Forwarded-For can contain multiple IPs, take the first one
    return Array.isArray(xff) ? xff[0] : xff.split(',')[0].trim();
  }

  // Check X-Real-IP header (for nginx)
  const xri = req.headers['x-real-ip'];
  if (xri) {
    return Array.isArray(xri) ? xri[0] : xri;
  }

  // Fall back to RemoteAddr
  return req.ip || req.connection?.remoteAddress || 'unknown';
}

// Login rate limit: 5 requests per minute
export const loginRateLimitMiddleware = rateLimit({
  windowMs: 60 * 1000, // 1 minute
  max: 5,
  message: 'Rate limit exceeded. Please try again later.',
  keyGenerator: getClientIP,
  standardHeaders: true,
  legacyHeaders: false,
});

// Registration rate limit: 10 requests per hour
export const registrationRateLimitMiddleware = rateLimit({
  windowMs: 60 * 60 * 1000, // 1 hour
  max: 10,
  message: 'Rate limit exceeded. Please try again later.',
  keyGenerator: getClientIP,
  standardHeaders: true,
  legacyHeaders: false,
});

// Password reset rate limit: 3 requests per hour
export const passwordResetRateLimitMiddleware = rateLimit({
  windowMs: 60 * 60 * 1000, // 1 hour
  max: 3,
  message: 'Rate limit exceeded. Please try again later.',
  keyGenerator: getClientIP,
  standardHeaders: true,
  legacyHeaders: false,
});

// Verification resend rate limit: 5 requests per hour
export const verificationResendRateLimitMiddleware = rateLimit({
  windowMs: 60 * 60 * 1000, // 1 hour
  max: 5,
  message: 'Rate limit exceeded. Please try again later.',
  keyGenerator: getClientIP,
  standardHeaders: true,
  legacyHeaders: false,
});

// General API rate limit: 100 requests per minute
export const generalAPIRateLimitMiddleware = rateLimit({
  windowMs: 60 * 1000, // 1 minute
  max: 100,
  message: 'Rate limit exceeded. Please try again later.',
  keyGenerator: getClientIP,
  standardHeaders: true,
  legacyHeaders: false,
});



