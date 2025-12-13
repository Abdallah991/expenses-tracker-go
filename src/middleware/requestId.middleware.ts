import { Request, Response, NextFunction } from 'express';
import { randomUUID } from 'crypto';

export interface RequestWithId extends Request {
  id?: string;
}

/**
 * Adds a unique request ID to each request for tracing and logging
 */
export function requestIdMiddleware(
  req: RequestWithId,
  res: Response,
  next: NextFunction
): void {
  req.id = randomUUID();
  res.setHeader('X-Request-ID', req.id);
  next();
}



