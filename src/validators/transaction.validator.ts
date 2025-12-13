import { body } from 'express-validator';

export const createTransactionValidator = [
  body('amount')
    .isFloat({ min: 0.01 })
    .withMessage('Amount must be a positive number greater than zero'),
];



