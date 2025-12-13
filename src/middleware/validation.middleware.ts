import { Request, Response, NextFunction } from "express";
import { validationResult, ValidationChain } from "express-validator";
import { ValidationError } from "../utils/errors.util";

export function validate(validations: ValidationChain[]) {
  return async (req: Request, _res: Response, next: NextFunction): Promise<void> => {
    // Run all validations
    await Promise.all(validations.map((validation) => validation.run(req)));

    // Check for validation errors
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      const errorMessages = errors.array().map((err) => {
        if ("msg" in err) {
          return err.msg;
        }
        return "Validation failed";
      });
      throw new ValidationError("Validation failed", errorMessages.join(", "));
    }

    next();
  };
}
