import { body, query } from "express-validator";

export const registerValidator = [
  body("email").isEmail().withMessage("Invalid email format").normalizeEmail({
    gmail_remove_dots: false,
    gmail_remove_subaddress: false,
    gmail_convert_googlemaildotcom: false,
  }),
  body("password")
    .isLength({ min: 8 })
    .withMessage("Password must be at least 8 characters long")
    .matches(/[A-Z]/)
    .withMessage("Password must contain at least one uppercase letter")
    .matches(/[a-z]/)
    .withMessage("Password must contain at least one lowercase letter")
    .matches(/[0-9]/)
    .withMessage("Password must contain at least one number")
    .matches(/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`]/)
    .withMessage("Password must contain at least one special character"),
];

export const loginValidator = [
  body("email").isEmail().withMessage("Invalid email format").normalizeEmail({
    gmail_remove_dots: false,
    gmail_remove_subaddress: false,
    gmail_convert_googlemaildotcom: false,
  }),
  body("password").notEmpty().withMessage("Password is required"),
];

export const resendVerificationValidator = [
  body("email").isEmail().withMessage("Invalid email format").normalizeEmail({
    gmail_remove_dots: false,
    gmail_remove_subaddress: false,
    gmail_convert_googlemaildotcom: false,
  }),
];

export const forgotPasswordValidator = [
  body("email").isEmail().withMessage("Invalid email format").normalizeEmail({
    gmail_remove_dots: false,
    gmail_remove_subaddress: false,
    gmail_convert_googlemaildotcom: false,
  }),
];

export const resetPasswordValidator = [
  body("token").notEmpty().withMessage("Token is required"),
  body("new_password")
    .isLength({ min: 8 })
    .withMessage("Password must be at least 8 characters long")
    .matches(/[A-Z]/)
    .withMessage("Password must contain at least one uppercase letter")
    .matches(/[a-z]/)
    .withMessage("Password must contain at least one lowercase letter")
    .matches(/[0-9]/)
    .withMessage("Password must contain at least one number")
    .matches(/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`]/)
    .withMessage("Password must contain at least one special character"),
];

export const refreshTokenValidator = [
  body("refresh_token").notEmpty().withMessage("Refresh token is required"),
];

export const verifyEmailValidator = [query("token").notEmpty().withMessage("Token is required")];

export const logoutValidator = [
  body("refresh_token").notEmpty().withMessage("Refresh token is required"),
];
