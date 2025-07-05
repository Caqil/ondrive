import { BaseDocument, LanguageCode, ObjectId } from "./global";

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  language: LanguageCode;
  timezone: string;
  emailNotifications: boolean;
  pushNotifications: boolean;
  defaultView: 'grid' | 'list';
  uploadQuality: 'original' | 'high' | 'medium';
}

export interface UserProviders {
  google?: {
    id: string;
    email: string;
  };
  github?: {
    id: string;
    username: string;
  };
}

export interface User extends BaseDocument {
  email: string;
  password?: string;
  name: string;
  avatar?: string;
  role: 'viewer' | 'user' | 'moderator' | 'admin';
  emailVerified: boolean;
  emailVerificationToken?: string;
  passwordResetToken?: string;
  passwordResetExpires?: Date;
  twoFactorEnabled: boolean;
  twoFactorSecret?: string;
  lastLogin?: Date;
  isActive: boolean;
  isBanned: boolean;
  banReason?: string;
  bannedAt?: Date;
  bannedBy?: ObjectId;
  storageUsed: number;
  storageQuota: number;
  currentTeam?: ObjectId;
  providers: UserProviders;
  preferences: UserPreferences;
  subscription?: ObjectId;
  subscriptionStatus: 'active' | 'inactive' | 'trial' | 'expired' | 'cancelled';
  trialEndsAt?: Date;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
}

export interface UpdateUserRequest {
  name?: string;
  avatar?: string;
  preferences?: Partial<UserPreferences>;
}

export interface UserWithStats extends User {
  fileCount: number;
  folderCount: number;
  shareCount: number;
  teamCount: number;
}