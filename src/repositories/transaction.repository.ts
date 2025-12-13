import { Pool } from 'pg';
import { getDB } from '../config/database';
import { Transaction } from '../types/models';

export class TransactionRepository {
  private db: Pool;

  constructor() {
    this.db = getDB();
  }

  async findAllByUserId(userId: number): Promise<Transaction[]> {
    const result = await this.db.query(
      'SELECT id, amount, user_id FROM transaction WHERE user_id = $1 ORDER BY id DESC',
      [userId]
    );

    return result.rows.map(row => ({
      id: row.id,
      amount: parseFloat(row.amount),
      user_id: row.user_id,
    }));
  }

  async create(userId: number, amount: number): Promise<Transaction> {
    const result = await this.db.query(
      'INSERT INTO transaction (amount, user_id) VALUES ($1, $2) RETURNING id',
      [amount, userId]
    );

    return {
      id: result.rows[0].id,
      amount: amount,
      user_id: userId,
    };
  }
}



