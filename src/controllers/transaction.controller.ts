import { Response, NextFunction } from 'express';
import { AuthenticatedRequest, getUserIDFromRequest } from '../middleware/auth.middleware';
import { TransactionService } from '../services/transaction.service';
import { CreateTransactionRequest } from '../types/requests';

export class TransactionController {
  private transactionService: TransactionService;

  constructor() {
    this.transactionService = new TransactionService();
  }

  async getTransactions(req: AuthenticatedRequest, res: Response, next: NextFunction): Promise<void> {
    try {
      const userId = getUserIDFromRequest(req);
      const transactions = await this.transactionService.getTransactions(userId);
      res.status(200).json(transactions);
    } catch (error) {
      next(error);
    }
  }

  async createTransaction(req: AuthenticatedRequest, res: Response, next: NextFunction): Promise<void> {
    try {
      const userId = getUserIDFromRequest(req);
      const data: CreateTransactionRequest = req.body;
      const transaction = await this.transactionService.createTransaction(userId, data);
      res.status(201).json(transaction);
    } catch (error) {
      next(error);
    }
  }
}



