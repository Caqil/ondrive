import { User } from ".";

export interface LoginRequest {
  email: string;
  password: string;
  rememberMe?: boolean;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  acceptTerms: boolean;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

export interface TwoFactorSetupRequest {
  secret: string;
  code: string;
}

export interface TwoFactorVerifyRequest {
  code: string;
}

export interface AuthSession {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresAt: Date;
}

export interface OAuthProvider {
  id: string;
  name: string;
  enabled: boolean;
  clientId: string;
  scope: string[];
}