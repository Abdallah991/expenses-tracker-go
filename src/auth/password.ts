import bcrypt from "bcrypt";

const MIN_PASSWORD_LENGTH = 8;
const MAX_PASSWORD_LENGTH = 72; // bcrypt max length
const BCRYPT_COST = 12;

const hasUppercase = /[A-Z]/;
const hasLowercase = /[a-z]/;
const hasNumber = /[0-9]/;
const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`]/;

const weakPasswords = [
  "password",
  "123456",
  "123456789",
  "qwerty",
  "abc123",
  "password123",
  "admin",
  "letmein",
  "welcome",
  "monkey",
];

export class PasswordValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PasswordValidationError";
  }
}

export function hashPassword(password: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const validationError = validatePassword(password);
    if (validationError) {
      reject(validationError);
      return;
    }

    bcrypt.hash(password, BCRYPT_COST, (err, hash) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(hash);
    });
  });
}

export function comparePassword(hashedPassword: string, password: string): Promise<void> {
  return new Promise((resolve, reject) => {
    bcrypt.compare(password, hashedPassword, (err, result) => {
      if (err) {
        reject(err);
        return;
      }
      if (!result) {
        reject(new Error("Invalid password"));
        return;
      }
      resolve();
    });
  });
}

export function validatePassword(password: string): PasswordValidationError | null {
  // Length - short scenario
  if (password.length < MIN_PASSWORD_LENGTH) {
    return new PasswordValidationError("Password must be at least 8 characters long");
  }

  // Length - long scenario
  if (password.length > MAX_PASSWORD_LENGTH) {
    return new PasswordValidationError("Password must be no more than 72 characters long");
  }

  // Check for common weak passwords
  const lowerPassword = password.toLowerCase();
  for (const weak of weakPasswords) {
    if (lowerPassword.includes(weak)) {
      return new PasswordValidationError("Password contains common weak patterns");
    }
  }

  // Check character requirements - uppercase
  if (!hasUppercase.test(password)) {
    return new PasswordValidationError("Password must contain at least one uppercase letter");
  }

  // lowercase
  if (!hasLowercase.test(password)) {
    return new PasswordValidationError("Password must contain at least one lowercase letter");
  }

  // numeric
  if (!hasNumber.test(password)) {
    return new PasswordValidationError("Password must contain at least one number");
  }

  // special character
  if (!hasSpecial.test(password)) {
    return new PasswordValidationError("Password must contain at least one special character");
  }

  // Check for repeated characters (more than 3 in a row)
  if (hasRepeatedChars(password)) {
    return new PasswordValidationError(
      "Password cannot contain more than 3 repeated characters in a row"
    );
  }

  return null;
}

function hasRepeatedChars(password: string): boolean {
  if (password.length < 4) {
    return false;
  }

  for (let i = 0; i < password.length - 3; i++) {
    if (
      password[i] === password[i + 1] &&
      password[i + 1] === password[i + 2] &&
      password[i + 2] === password[i + 3]
    ) {
      return true;
    }
  }

  return false;
}

export function isPasswordValidationError(err: unknown): err is PasswordValidationError {
  return err instanceof PasswordValidationError;
}
