import { Router } from "express";
import authRoutes from "./auth.routes";
import transactionRoutes from "./transaction.routes";
import { AuthController } from "../controllers/auth.controller";

const router = Router();
const authController = new AuthController();

// Health check
router.get("/status", (req, res) => authController.status(req, res));

// API routes
router.use("/auth", authRoutes);
router.use("/transaction", transactionRoutes);

export default router;
