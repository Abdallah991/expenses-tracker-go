import { Router } from "express";
import { AuthController } from "../controllers/auth.controller";
import { requireAuth } from "../middleware/auth.middleware";
import {
  registrationRateLimitMiddleware,
  loginRateLimitMiddleware,
  verificationResendRateLimitMiddleware,
  passwordResetRateLimitMiddleware,
} from "../middleware/rateLimit.middleware";
import { validate } from "../middleware/validation.middleware";
import {
  registerValidator,
  loginValidator,
  resendVerificationValidator,
  forgotPasswordValidator,
  resetPasswordValidator,
  refreshTokenValidator,
  verifyEmailValidator,
  logoutValidator,
} from "../validators/auth.validator";

const router = Router();
const authController = new AuthController();

// Public routes
router.post(
  "/register",
  registrationRateLimitMiddleware,
  validate(registerValidator),
  authController.register.bind(authController)
);
router.post(
  "/login",
  loginRateLimitMiddleware,
  validate(loginValidator),
  authController.login.bind(authController)
);
router.get(
  "/verify-email",
  validate(verifyEmailValidator),
  authController.verifyEmail.bind(authController)
);
router.post(
  "/resend-verification",
  verificationResendRateLimitMiddleware,
  validate(resendVerificationValidator),
  authController.resendVerification.bind(authController)
);
router.post(
  "/forgot-password",
  passwordResetRateLimitMiddleware,
  validate(forgotPasswordValidator),
  authController.forgotPassword.bind(authController)
);
router.get("/reset-password-redirect", authController.redirectResetPassword.bind(authController));
router.post(
  "/reset-password",
  validate(resetPasswordValidator),
  authController.resetPassword.bind(authController)
);
router.post(
  "/refresh",
  validate(refreshTokenValidator),
  authController.refreshToken.bind(authController)
);

// Protected routes
router.post(
  "/logout",
  requireAuth,
  validate(logoutValidator),
  authController.logout.bind(authController)
);

export default router;
