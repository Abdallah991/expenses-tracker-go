import { Pool } from "pg";
import { getDB } from "../config/database";
import { User } from "../types/models";
// import { NotFoundError } from '../utils/errors.util';

export class UserRepository {
  private db: Pool;

  constructor() {
    this.db = getDB();
  }

  async findByEmail(email: string): Promise<User | null> {
    const result = await this.db.query(
      `SELECT id, email, password_hash, email_verified, verification_token, 
              verification_token_expires, reset_token, reset_token_expires,
              failed_login_attempts, locked_until, created_at, updated_at
       FROM users WHERE email = $1`,
      [email]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return this.mapRowToUser(result.rows[0]);
  }

  async findById(id: number): Promise<User | null> {
    const result = await this.db.query(
      `SELECT id, email, password_hash, email_verified, verification_token, 
              verification_token_expires, reset_token, reset_token_expires,
              failed_login_attempts, locked_until, created_at, updated_at
       FROM users WHERE id = $1`,
      [id]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return this.mapRowToUser(result.rows[0]);
  }

  async findByEmailAndVerified(email: string): Promise<User | null> {
    const result = await this.db.query(
      `SELECT id, email, password_hash, email_verified, verification_token, 
              verification_token_expires, reset_token, reset_token_expires,
              failed_login_attempts, locked_until, created_at, updated_at
       FROM users WHERE email = $1 AND email_verified = true`,
      [email]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return this.mapRowToUser(result.rows[0]);
  }

  async emailExists(email: string): Promise<boolean> {
    const result = await this.db.query("SELECT id FROM users WHERE email = $1", [email]);
    return result.rows.length > 0;
  }

  async create(
    email: string,
    passwordHash: string,
    verificationToken: string,
    verificationExpires: Date
  ): Promise<number> {
    const result = await this.db.query(
      `INSERT INTO users (email, password_hash, verification_token, verification_token_expires)
       VALUES ($1, $2, $3, $4)
       RETURNING id`,
      [email, passwordHash, verificationToken, verificationExpires]
    );
    return result.rows[0].id;
  }

  async updateVerificationToken(
    userId: number,
    verificationToken: string,
    verificationExpires: Date
  ): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET verification_token = $1, verification_token_expires = $2, updated_at = NOW()
       WHERE id = $3`,
      [verificationToken, verificationExpires, userId]
    );
  }

  async verifyEmail(userId: number): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET email_verified = true, verification_token = NULL, verification_token_expires = NULL, updated_at = NOW()
       WHERE id = $1`,
      [userId]
    );
  }

  async findByVerificationToken(token: string): Promise<number | null> {
    const result = await this.db.query(
      `SELECT id FROM users 
       WHERE verification_token = $1 
       AND verification_token_expires > NOW()`,
      [token]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return result.rows[0].id;
  }

  async updateResetToken(userId: number, resetToken: string, resetExpires: Date): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET reset_token = $1, reset_token_expires = $2, updated_at = NOW()
       WHERE id = $3`,
      [resetToken, resetExpires, userId]
    );
  }

  async findByResetToken(token: string): Promise<number | null> {
    const result = await this.db.query(
      `SELECT id FROM users 
       WHERE reset_token = $1 
       AND reset_token_expires > NOW()`,
      [token]
    );

    if (result.rows.length === 0) {
      return null;
    }

    return result.rows[0].id;
  }

  async updatePassword(userId: number, passwordHash: string): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET password_hash = $1, reset_token = NULL, reset_token_expires = NULL, 
           failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
       WHERE id = $2`,
      [passwordHash, userId]
    );
  }

  async updateFailedLoginAttempts(
    userId: number,
    attempts: number,
    lockedUntil: Date | null
  ): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET failed_login_attempts = $1, locked_until = $2, updated_at = NOW()
       WHERE id = $3`,
      [attempts, lockedUntil, userId]
    );
  }

  async resetFailedLoginAttempts(userId: number): Promise<void> {
    await this.db.query(
      `UPDATE users 
       SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
       WHERE id = $1`,
      [userId]
    );
  }

  async checkUserExistsAndVerified(userId: number): Promise<boolean> {
    const result = await this.db.query(
      "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND email_verified = true)",
      [userId]
    );
    return result.rows[0].exists;
  }

  private mapRowToUser(row: any): User {
    return {
      id: row.id,
      email: row.email,
      password_hash: row.password_hash,
      email_verified: row.email_verified,
      verification_token: row.verification_token,
      verification_token_expires: row.verification_token_expires
        ? new Date(row.verification_token_expires)
        : null,
      reset_token: row.reset_token,
      reset_token_expires: row.reset_token_expires ? new Date(row.reset_token_expires) : null,
      failed_login_attempts: row.failed_login_attempts,
      locked_until: row.locked_until ? new Date(row.locked_until) : null,
      created_at: new Date(row.created_at),
      updated_at: new Date(row.updated_at),
    };
  }
}
