import { UserResponse } from './models';

export interface SuccessResponse {
  message: string;
}

export interface ErrorResponse {
  error: string;
  details?: string;
}

export interface LoginResponse {
  message: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: UserResponse;
}

export interface RefreshTokenResponse {
  message: string;
  access_token: string;
  expires_in: number;
}



