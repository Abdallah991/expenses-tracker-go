import { UserRepository } from "../repositories/user.repository";
import { RefreshTokenRepository } from "../repositories/refreshToken.repository";
import { EmailService, getEmailService } from "./email.service";
import { hashPassword, comparePassword, validatePassword } from "../auth/password";
import {
  generateAccessToken,
  generateRefreshToken,
  getTokenExpiry,
  getRefreshTokenExpiry,
  generateSecureToken,
} from "../auth/jwt";
import {
  RegisterRequest,
  LoginRequest,
  RefreshTokenRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  ResendVerificationRequest,
} from "../types/requests";
import { SuccessResponse, LoginResponse, RefreshTokenResponse } from "../types/responses";
import { toUserResponse } from "../types/models";
import {
  ValidationError,
  ConflictError,
  AuthenticationError,
  AuthorizationError,
  LockedError,
} from "../utils/errors.util";

export class AuthService {
  private userRepository: UserRepository;
  private refreshTokenRepository: RefreshTokenRepository;
  private emailService: EmailService;

  constructor() {
    this.userRepository = new UserRepository();
    this.refreshTokenRepository = new RefreshTokenRepository();
    this.emailService = getEmailService();
  }

  async register(data: RegisterRequest): Promise<SuccessResponse> {
    // Validate password strength
    const passwordError = validatePassword(data.password);
    if (passwordError) {
      throw new ValidationError("Password validation failed", passwordError.message);
    }

    // Check if email already exists
    const emailExists = await this.userRepository.emailExists(data.email);
    if (emailExists) {
      throw new ConflictError("Email already registered");
    }

    // Hash password
    const hashedPassword = await hashPassword(data.password);

    // Generate verification token
    const verificationToken = generateSecureToken(32);
    const verificationExpires = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours

    // Create user
    // const userId =
    await this.userRepository.create(
      data.email,
      hashedPassword,
      verificationToken,
      verificationExpires
    );

    // Send verification email (don't fail registration if email fails)
    try {
      await this.emailService.sendVerificationEmail(data.email, verificationToken);
    } catch (err) {
      console.error("Failed to send verification email:", err);
    }

    return {
      message: "User registered successfully. Please check your email to verify your account.",
    };
  }

  async login(data: LoginRequest): Promise<LoginResponse> {
    // Get user from database
    const user = await this.userRepository.findByEmail(data.email);
    if (!user) {
      throw new AuthenticationError("Invalid credentials");
    }

    // Check if account is locked
    if (user.locked_until && new Date() < user.locked_until) {
      throw new LockedError("Too many failed login attempts. Please try again later.");
    }

    // Check if email is verified
    if (!user.email_verified) {
      throw new AuthorizationError("Please verify your email before logging in");
    }

    // Verify password
    try {
      await comparePassword(user.password_hash, data.password);
    } catch (err) {
      // Increment failed login attempts
      const newAttempts = user.failed_login_attempts + 1;
      let lockedUntil: Date | null = null;

      if (newAttempts >= 5) {
        lockedUntil = new Date(Date.now() + 15 * 60 * 1000); // 15 minutes
      }

      await this.userRepository.updateFailedLoginAttempts(user.id, newAttempts, lockedUntil);
      throw new AuthenticationError("Invalid credentials");
    }

    // Reset failed login attempts on successful login
    await this.userRepository.resetFailedLoginAttempts(user.id);

    // Generate tokens
    const accessToken = generateAccessToken(user.id, user.email);
    const refreshToken = generateRefreshToken();

    // Store refresh token in database
    const refreshExpires = new Date(Date.now() + getRefreshTokenExpiry() * 1000);
    await this.refreshTokenRepository.create(user.id, refreshToken, refreshExpires);

    return {
      message: "Login successful",
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_in: getTokenExpiry(),
      user: toUserResponse(user),
    };
  }

  async verifyEmail(token: string): Promise<SuccessResponse> {
    const userId = await this.userRepository.findByVerificationToken(token);
    if (!userId) {
      throw new ValidationError("Invalid or expired token");
    }

    await this.userRepository.verifyEmail(userId);

    return {
      message: "Email verified successfully. You can now log in.",
    };
  }

  async resendVerification(data: ResendVerificationRequest): Promise<SuccessResponse> {
    const user = await this.userRepository.findByEmail(data.email);

    // Don't reveal if email exists or not
    if (!user) {
      return {
        message: "If the email exists and is not verified, a verification email has been sent.",
      };
    }

    if (user.email_verified) {
      throw new ValidationError("Email already verified");
    }

    // Generate new verification token
    const verificationToken = generateSecureToken(32);
    const verificationExpires = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours

    // Update verification token
    await this.userRepository.updateVerificationToken(
      user.id,
      verificationToken,
      verificationExpires
    );

    // Send verification email
    await this.emailService.sendVerificationEmail(data.email, verificationToken);

    return {
      message: "Verification email sent successfully.",
    };
  }

  async forgotPassword(data: ForgotPasswordRequest): Promise<SuccessResponse> {
    const user = await this.userRepository.findByEmail(data.email);

    // Don't reveal if email exists or not
    if (!user) {
      return {
        message: "If the email exists, a password reset link has been sent.",
      };
    }

    // Generate reset token
    const resetToken = generateSecureToken(32);
    const resetExpires = new Date(Date.now() + 60 * 60 * 1000); // 1 hour

    // Update reset token
    await this.userRepository.updateResetToken(user.id, resetToken, resetExpires);

    // Send reset email
    await this.emailService.sendPasswordResetEmail(data.email, resetToken);

    return {
      message: "If the email exists, a password reset link has been sent.",
    };
  }

  async resetPassword(data: ResetPasswordRequest): Promise<SuccessResponse> {
    // Validate new password strength
    const passwordError = validatePassword(data.new_password);
    if (passwordError) {
      throw new ValidationError("Password validation failed", passwordError.message);
    }

    // Verify reset token
    const userId = await this.userRepository.findByResetToken(data.token);
    if (!userId) {
      throw new ValidationError("Invalid or expired reset token");
    }

    // Hash new password
    const hashedPassword = await hashPassword(data.new_password);

    // Update password and clear reset token
    await this.userRepository.updatePassword(userId, hashedPassword);

    // Invalidate all refresh tokens for this user
    await this.refreshTokenRepository.deleteAllByUserId(userId);

    return {
      message: "Password reset successfully. Please log in with your new password.",
    };
  }

  async refreshToken(data: RefreshTokenRequest): Promise<RefreshTokenResponse> {
    const tokenData = await this.refreshTokenRepository.findByToken(data.refresh_token);
    if (!tokenData) {
      throw new AuthenticationError("Invalid or expired refresh token");
    }

    // Generate new access token
    const accessToken = generateAccessToken(tokenData.userId, tokenData.email);

    return {
      message: "Token refreshed successfully",
      access_token: accessToken,
      expires_in: getTokenExpiry(),
    };
  }

  async logout(userId: number, refreshToken: string): Promise<SuccessResponse> {
    const deleted = await this.refreshTokenRepository.deleteByTokenAndUserId(refreshToken, userId);
    if (!deleted) {
      throw new ValidationError("Invalid refresh token");
    }

    return {
      message: "Logged out successfully",
    };
  }
}
