import { Router } from "express";
import { TransactionController } from "../controllers/transaction.controller";
import { requireAuth } from "../middleware/auth.middleware";
import { validate } from "../middleware/validation.middleware";
import { createTransactionValidator } from "../validators/transaction.validator";

const router = Router();
const transactionController = new TransactionController();

// All transaction routes require authentication
router.get("/", requireAuth, transactionController.getTransactions.bind(transactionController));
router.post(
  "/",
  requireAuth,
  validate(createTransactionValidator),
  transactionController.createTransaction.bind(transactionController)
);

export default router;
