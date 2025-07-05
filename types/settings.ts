import { BaseDocument, ObjectId } from ".";

export interface AppSettings {
  name: string;
  description: string;
  logo?: string;
  favicon?: string;
  primaryColor: string;
  secondaryColor: string;
  maxFileSize: number;
  maxFilesPerUpload: number;
  allowedFileTypes: string[];
  defaultUserQuota: number;
  enableRegistration: boolean;
  enableGuestUploads: boolean;
  enablePublicSharing: boolean;
  maintenanceMode: boolean;
  maintenanceMessage?: string;
}

export interface StorageSettings {
  default: 'local' | 's3' | 'r2' | 'wasabi' | 'gcs' | 'azure';
  providers: {
    local: {
      enabled: boolean;
      path: string;
      maxSize: number;
    };
    s3: {
      enabled: boolean;
      accessKeyId: string;
      secretAccessKey: string;
      region: string;
      bucket: string;
      endpoint?: string;
      forcePathStyle: boolean;
    };
    r2: {
      enabled: boolean;
      accessKeyId: string;
      secretAccessKey: string;
      accountId: string;
      bucket: string;
      endpoint: string;
    };
    wasabi: {
      enabled: boolean;
      accessKeyId: string;
      secretAccessKey: string;
      region: string;
      bucket: string;
      endpoint: string;
    };
    gcs: {
      enabled: boolean;
      projectId: string;
      keyFilename: string;
      bucket: string;
    };
    azure: {
      enabled: boolean;
      accountName: string;
      accountKey: string;
      containerName: string;
    };
  };
}

export interface EmailSettings {
  enabled: boolean;
  provider: 'smtp' | 'sendgrid' | 'ses' | 'mailgun' | 'resend';
  from: string;
  replyTo?: string;
  smtp: {
    host: string;
    port: number;
    secure: boolean;
    auth: {
      user: string;
      pass: string;
    };
  };
  sendgrid: {
    apiKey: string;
  };
  ses: {
    accessKeyId: string;
    secretAccessKey: string;
    region: string;
  };
  mailgun: {
    apiKey: string;
    domain: string;
  };
  resend: {
    apiKey: string;
  };
  templates: {
    welcome: boolean;
    shareNotification: boolean;
    passwordReset: boolean;
    emailVerification: boolean;
  };
}

export interface SecuritySettings {
  enableTwoFactor: boolean;
  sessionTimeout: number;
  maxLoginAttempts: number;
  lockoutDuration: number;
  passwordMinLength: number;
  passwordRequireNumbers: boolean;
  passwordRequireSymbols: boolean;
  passwordRequireUppercase: boolean;
  passwordRequireLowercase: boolean;
  enableCaptcha: boolean;
  captchaSiteKey?: string;
  captchaSecretKey?: string;
}

export interface FeatureSettings {
  enableSearch: boolean;
  enableOCR: boolean;
  enableThumbnails: boolean;
  enableVersioning: boolean;
  enableTeams: boolean;
  enableAPI: boolean;
  enableWebDAV: boolean;
  enableOfflineSync: boolean;
  maxVersionsPerFile: number;
}

export interface RateLimitSettings {
  enabled: boolean;
  uploadLimit: number;
  downloadLimit: number;
  apiLimit: number;
  shareLimit: number;
}

export interface Settings extends BaseDocument {
  app: AppSettings;
  storage: StorageSettings;
  email: EmailSettings;
  security: SecuritySettings;
  features: FeatureSettings;
  analytics: {
    enabled: boolean;
    provider?: 'google' | 'mixpanel' | 'amplitude';
    trackingId?: string;
    apiKey?: string;
    collectUsageStats: boolean;
    collectErrorLogs: boolean;
  };
  backup: {
    enabled: boolean;
    frequency: 'daily' | 'weekly' | 'monthly';
    retention: number;
    destination: 'local' | 's3' | 'gcs';
    encryption: boolean;
    encryptionKey?: string;
  };
  notifications: {
    enablePush: boolean;
    pushProvider?: 'firebase' | 'pusher' | 'websocket';
    firebaseServerKey?: string;
    pusherAppId?: string;
    pusherKey?: string;
    pusherSecret?: string;
    pusherCluster?: string;
  };
  rateLimiting: RateLimitSettings;
  version: string;
  lastUpdatedBy: ObjectId;
}

export interface UpdateSettingsRequest {
  app?: Partial<AppSettings>;
  storage?: Partial<StorageSettings>;
  email?: Partial<EmailSettings>;
  security?: Partial<SecuritySettings>;
  features?: Partial<FeatureSettings>;
  rateLimiting?: Partial<RateLimitSettings>;
}