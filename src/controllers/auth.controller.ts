import { Request, Response, NextFunction } from "express";
import { AuthenticatedRequest, getUserIDFromRequest } from "../middleware/auth.middleware";
import { AuthService } from "../services/auth.service";
import {
  RegisterRequest,
  LoginRequest,
  RefreshTokenRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  ResendVerificationRequest,
  LogoutRequest,
} from "../types/requests";

export class AuthController {
  private authService: AuthService;

  constructor() {
    this.authService = new AuthService();
  }

  async register(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: RegisterRequest = req.body;
      const result = await this.authService.register(data);
      res.status(201).json(result);
    } catch (error) {
      next(error);
    }
  }

  async login(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: LoginRequest = req.body;
      const result = await this.authService.login(data);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async verifyEmail(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const token = req.query.token as string;
      if (!token) {
        res
          .status(400)
          .json({ error: "Token required", message: "Please provide a verification token" });
        return;
      }
      const result = await this.authService.verifyEmail(token);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async resendVerification(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: ResendVerificationRequest = req.body;
      const result = await this.authService.resendVerification(data);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async forgotPassword(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: ForgotPasswordRequest = req.body;
      const result = await this.authService.forgotPassword(data);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async redirectResetPassword(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const token = req.query.token as string;
      if (!token) {
        res.status(400).json({ error: "Token required", message: "Please provide a reset token" });
        return;
      }

      // Verify reset token exists and is valid
      const authService = new AuthService();
      // We'll need to check the token, but for now just render the HTML
      // The actual validation happens in resetPassword

      // Get user agent to detect mobile devices
      const userAgent = req.headers["user-agent"] || "";
      const isMobile = this.isMobileDevice(userAgent);

      // Get deep link scheme from environment
      const deepLinkScheme = process.env.MOBILE_DEEP_LINK_SCHEME || "myexpenses://";

      // For mobile devices, redirect to custom scheme
      if (isMobile) {
        const deepLinkURL = `${deepLinkScheme}reset-password?token=${token}`;
        res.redirect(302, deepLinkURL);
        return;
      }

      // For web browsers, serve an HTML page with redirect and form fallback
      const appURL = process.env.APP_URL || "http://localhost:8080";

      const htmlContent = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Reset Password - Expenses Tracker</title>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 50px auto; padding: 20px; }
    .container { background-color: #f9f9f9; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .header { background-color: #f44336; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; margin: -30px -30px 20px -30px; }
    .button { display: inline-block; padding: 12px 24px; background-color: #f44336; color: white; text-decoration: none; border-radius: 4px; margin: 10px 5px; cursor: pointer; border: none; font-size: 16px; }
    .button:hover { background-color: #d32f2f; }
    .button-secondary { background-color: #2196F3; }
    .button-secondary:hover { background-color: #1976D2; }
    .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 4px; margin: 20px 0; }
    .form-group { margin: 20px 0; }
    label { display: block; margin-bottom: 5px; font-weight: bold; }
    input[type="password"] { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 16px; box-sizing: border-box; }
    .hidden { display: none; }
  </style>
  <script>
    // Try to open mobile app immediately
    window.location.href = "${deepLinkScheme}reset-password?token=${token}";
    
    // If app doesn't open, show form after 2 seconds
    setTimeout(function() {
      document.getElementById('resetForm').classList.remove('hidden');
    }, 2000);
  </script>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Reset Your Password</h1>
    </div>
    <div id="redirecting" style="text-align: center; padding: 20px;">
      <p>Redirecting to app...</p>
      <p>If the app doesn't open, use the form below.</p>
    </div>
    <div id="resetForm" class="hidden">
      <p>Please enter your new password:</p>
      <form id="passwordResetForm" onsubmit="submitReset(event)">
        <input type="hidden" id="resetToken" value="${token}">
        <div class="form-group">
          <label for="new_password">New Password:</label>
          <input type="password" id="new_password" name="new_password" required minlength="8">
        </div>
        <button type="submit" class="button">Reset Password</button>
      </form>
      <div id="result" style="margin-top: 20px;"></div>
      <div class="warning">
        <p><strong>Note:</strong> For best security, please use the mobile app or API endpoint directly.</p>
      </div>
    </div>
    <script>
      function submitReset(event) {
        event.preventDefault();
        var token = document.getElementById('resetToken').value;
        var password = document.getElementById('new_password').value;
        var resultDiv = document.getElementById('result');
        
        fetch('${appURL}/auth/reset-password', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            token: token,
            new_password: password
          })
        })
        .then(response => response.json())
        .then(data => {
          if (data.message) {
            resultDiv.innerHTML = '<div style="background-color: #d4edda; color: #155724; padding: 15px; border-radius: 4px; margin-top: 20px;"><strong>Success!</strong> ' + data.message + '</div>';
            document.getElementById('passwordResetForm').style.display = 'none';
          } else if (data.error) {
            resultDiv.innerHTML = '<div style="background-color: #f8d7da; color: #721c24; padding: 15px; border-radius: 4px; margin-top: 20px;"><strong>Error:</strong> ' + data.error + (data.details ? ' - ' + data.details : '') + '</div>';
          }
        })
        .catch(error => {
          resultDiv.innerHTML = '<div style="background-color: #f8d7da; color: #721c24; padding: 15px; border-radius: 4px; margin-top: 20px;"><strong>Error:</strong> Failed to reset password. Please try again.</div>';
        });
      }
    </script>
  </div>
</body>
</html>
      `;

      res.setHeader("Content-Type", "text/html; charset=utf-8");
      res.status(200).send(htmlContent);
    } catch (error) {
      next(error);
    }
  }

  async resetPassword(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: ResetPasswordRequest = req.body;
      const result = await this.authService.resetPassword(data);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async refreshToken(req: Request, res: Response, next: NextFunction): Promise<void> {
    try {
      const data: RefreshTokenRequest = req.body;
      const result = await this.authService.refreshToken(data);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  async logout(req: AuthenticatedRequest, res: Response, next: NextFunction): Promise<void> {
    try {
      const userId = getUserIDFromRequest(req);
      const data: LogoutRequest = req.body;
      const result = await this.authService.logout(userId, data.refresh_token);
      res.status(200).json(result);
    } catch (error) {
      next(error);
    }
  }

  // Status for Auth Controller
  async status(_req: Request, res: Response): Promise<void> {
    res.status(200).json({
      status: "live",
      application: "Express TypeScript Web Server",
      message: "Application is live and running!",
    });
  }

  private isMobileDevice(userAgent: string): boolean {
    const mobileKeywords = [
      "android",
      "iphone",
      "ipad",
      "ipod",
      "blackberry",
      "windows phone",
      "mobile",
      "opera mini",
      "iemobile",
    ];
    const lowerUA = userAgent.toLowerCase();
    return mobileKeywords.some((keyword) => lowerUA.includes(keyword));
  }
}
