import { Pool } from "pg";
import { getDB } from "../config/database";
// import { RefreshToken } from '../types/models';

export class RefreshTokenRepository {
  private db: Pool;

  constructor() {
    this.db = getDB();
  }

  async create(userId: number, token: string, expiresAt: Date): Promise<void> {
    await this.db.query(
      `INSERT INTO refresh_tokens (user_id, token, expires_at)
       VALUES ($1, $2, $3)`,
      [userId, token, expiresAt]
    );
  }

  async findByToken(token: string): Promise<{ userId: number; email: string } | null> {
    const result = await this.db.query(
      `SELECT u.id, u.email 
       FROM users u
       JOIN refresh_tokens rt ON u.id = rt.user_id
       WHERE rt.token = $1 
       AND rt.expires_at > NOW()
       AND u.email_verified = true`,
      [token]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return {
      userId: result.rows[0].id,
      email: result.rows[0].email,
    };
  }

  async deleteByTokenAndUserId(token: string, userId: number): Promise<boolean> {
    const result = await this.db.query(
      `DELETE FROM refresh_tokens 
       WHERE token = $1 AND user_id = $2`,
      [token, userId]
    );
    return (result.rowCount || 0) > 0;
  }

  async deleteAllByUserId(userId: number): Promise<void> {
    await this.db.query("DELETE FROM refresh_tokens WHERE user_id = $1", [userId]);
  }
}
