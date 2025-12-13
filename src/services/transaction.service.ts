import { TransactionRepository } from '../repositories/transaction.repository';
import { Transaction } from '../types/models';
import { CreateTransactionRequest } from '../types/requests';
import { ValidationError } from '../utils/errors.util';

export class TransactionService {
  private transactionRepository: TransactionRepository;

  constructor() {
    this.transactionRepository = new TransactionRepository();
  }

  async getTransactions(userId: number): Promise<Transaction[]> {
    return await this.transactionRepository.findAllByUserId(userId);
  }

  async createTransaction(userId: number, data: CreateTransactionRequest): Promise<Transaction> {
    if (typeof data.amount !== 'number' || data.amount <= 0) {
      throw new ValidationError('Transaction amount must be a positive number greater than zero');
    }

    return await this.transactionRepository.create(userId, data.amount);
  }
}



