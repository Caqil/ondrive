import mongoose, { Document, Schema } from 'mongoose';

export interface ISettings extends Document {
  _id: string;
  
  // App Configuration
  app: {
    name: string;
    description: string;
    logo?: string;
    favicon?: string;
    primaryColor: string;
    secondaryColor: string;
    maxFileSize: number; // bytes
    maxFilesPerUpload: number;
    allowedFileTypes: string[];
    defaultUserQuota: number; // bytes
    enableRegistration: boolean;
    enableGuestUploads: boolean;
    enablePublicSharing: boolean;
    maintenanceMode: boolean;
    maintenanceMessage?: string;
  };
  
  // Storage Configuration
  storage: {
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
  };
  
  // Email Configuration
  email: {
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
  };
  
  // Security Configuration
  security: {
    enableTwoFactor: boolean;
    sessionTimeout: number; // minutes
    maxLoginAttempts: number;
    lockoutDuration: number; // minutes
    passwordMinLength: number;
    passwordRequireNumbers: boolean;
    passwordRequireSymbols: boolean;
    passwordRequireUppercase: boolean;
    passwordRequireLowercase: boolean;
    enableCaptcha: boolean;
    captchaSiteKey?: string;
    captchaSecretKey?: string;
  };
  
  // Features Configuration
  features: {
    enableSearch: boolean;
    enableOCR: boolean;
    enableThumbnails: boolean;
    enableVersioning: boolean;
    enableTeams: boolean;
    enableAPI: boolean;
    enableWebDAV: boolean;
    enableOfflineSync: boolean;
    maxVersionsPerFile: number;
  };
  
  // Analytics Configuration
  analytics: {
    enabled: boolean;
    provider?: 'google' | 'mixpanel' | 'amplitude';
    trackingId?: string;
    apiKey?: string;
    collectUsageStats: boolean;
    collectErrorLogs: boolean;
  };
  
  // Backup Configuration
  backup: {
    enabled: boolean;
    frequency: 'daily' | 'weekly' | 'monthly';
    retention: number; // days
    destination: 'local' | 's3' | 'gcs';
    encryption: boolean;
    encryptionKey?: string;
  };
  
  // Notifications Configuration
  notifications: {
    enablePush: boolean;
    pushProvider?: 'firebase' | 'pusher' | 'websocket';
    firebaseServerKey?: string;
    pusherAppId?: string;
    pusherKey?: string;
    pusherSecret?: string;
    pusherCluster?: string;
  };
  
  // Rate Limiting
  rateLimiting: {
    enabled: boolean;
    uploadLimit: number; // requests per minute
    downloadLimit: number; // requests per minute
    apiLimit: number; // requests per minute
    shareLimit: number; // requests per minute
  };
  
  // System
  version: string;
  lastUpdatedBy: mongoose.Types.ObjectId;
  
  createdAt: Date;
  updatedAt: Date;
}

const settingsSchema = new Schema<ISettings>({
  app: {
    name: {
      type: String,
      default: 'Drive Clone',
      maxlength: 100
    },
    description: {
      type: String,
      default: 'Secure cloud storage solution',
      maxlength: 500
    },
    logo: String,
    favicon: String,
    primaryColor: {
      type: String,
      default: '#3B82F6',
      match: /^#[0-9A-F]{6}$/i
    },
    secondaryColor: {
      type: String,
      default: '#64748B',
      match: /^#[0-9A-F]{6}$/i
    },
    maxFileSize: {
      type: Number,
      default: 750 * 1024 * 1024 * 1024, // 750GB
      min: 1024 * 1024 // 1MB minimum
    },
    maxFilesPerUpload: {
      type: Number,
      default: 100,
      min: 1,
      max: 1000
    },
    allowedFileTypes: [{
      type: String,
      lowercase: true
    }],
    defaultUserQuota: {
      type: Number,
      default: 15 * 1024 * 1024 * 1024, // 15GB
      min: 1024 * 1024 // 1MB minimum
    },
    enableRegistration: {
      type: Boolean,
      default: true
    },
    enableGuestUploads: {
      type: Boolean,
      default: false
    },
    enablePublicSharing: {
      type: Boolean,
      default: true
    },
    maintenanceMode: {
      type: Boolean,
      default: false
    },
    maintenanceMessage: String
  },
  
  storage: {
    default: {
      type: String,
      enum: ['local', 's3', 'r2', 'wasabi', 'gcs', 'azure'],
      default: 'local'
    },
    providers: {
      local: {
        enabled: {
          type: Boolean,
          default: true
        },
        path: {
          type: String,
          default: './uploads'
        },
        maxSize: {
          type: Number,
          default: 1024 * 1024 * 1024 * 1024 // 1TB
        }
      },
      s3: {
        enabled: {
          type: Boolean,
          default: false
        },
        accessKeyId: String,
        secretAccessKey: String,
        region: String,
        bucket: String,
        endpoint: String,
        forcePathStyle: {
          type: Boolean,
          default: false
        }
      },
      r2: {
        enabled: {
          type: Boolean,
          default: false
        },
        accessKeyId: String,
        secretAccessKey: String,
        accountId: String,
        bucket: String,
        endpoint: String
      },
      wasabi: {
        enabled: {
          type: Boolean,
          default: false
        },
        accessKeyId: String,
        secretAccessKey: String,
        region: String,
        bucket: String,
        endpoint: String
      },
      gcs: {
        enabled: {
          type: Boolean,
          default: false
        },
        projectId: String,
        keyFilename: String,
        bucket: String
      },
      azure: {
        enabled: {
          type: Boolean,
          default: false
        },
        accountName: String,
        accountKey: String,
        containerName: String
      }
    }
  },
  
  email: {
    enabled: {
      type: Boolean,
      default: false
    },
    provider: {
      type: String,
      enum: ['smtp', 'sendgrid', 'ses', 'mailgun', 'resend'],
      default: 'smtp'
    },
    from: {
      type: String,
      default: 'noreply@example.com'
    },
    replyTo: String,
    smtp: {
      host: String,
      port: {
        type: Number,
        default: 587
      },
      secure: {
        type: Boolean,
        default: false
      },
      auth: {
        user: String,
        pass: String
      }
    },
    sendgrid: {
      apiKey: String
    },
    ses: {
      accessKeyId: String,
      secretAccessKey: String,
      region: String
    },
    mailgun: {
      apiKey: String,
      domain: String
    },
    resend: {
      apiKey: String
    },
    templates: {
      welcome: {
        type: Boolean,
        default: true
      },
      shareNotification: {
        type: Boolean,
        default: true
      },
      passwordReset: {
        type: Boolean,
        default: true
      },
      emailVerification: {
        type: Boolean,
        default: true
      }
    }
  },
  
  security: {
    enableTwoFactor: {
      type: Boolean,
      default: false
    },
    sessionTimeout: {
      type: Number,
      default: 60 * 24, // 24 hours
      min: 15
    },
    maxLoginAttempts: {
      type: Number,
      default: 5,
      min: 3,
      max: 10
    },
    lockoutDuration: {
      type: Number,
      default: 15,
      min: 5
    },
    passwordMinLength: {
      type: Number,
      default: 8,
      min: 6,
      max: 128
    },
    passwordRequireNumbers: {
      type: Boolean,
      default: true
    },
    passwordRequireSymbols: {
      type: Boolean,
      default: true
    },
    passwordRequireUppercase: {
      type: Boolean,
      default: true
    },
    passwordRequireLowercase: {
      type: Boolean,
      default: true
    },
    enableCaptcha: {
      type: Boolean,
      default: false
    },
    captchaSiteKey: String,
    captchaSecretKey: String
  },
  
  features: {
    enableSearch: {
      type: Boolean,
      default: true
    },
    enableOCR: {
      type: Boolean,
      default: true
    },
    enableThumbnails: {
      type: Boolean,
      default: true
    },
    enableVersioning: {
      type: Boolean,
      default: true
    },
    enableTeams: {
      type: Boolean,
      default: true
    },
    enableAPI: {
      type: Boolean,
      default: true
    },
    enableWebDAV: {
      type: Boolean,
      default: false
    },
    enableOfflineSync: {
      type: Boolean,
      default: true
    },
    maxVersionsPerFile: {
      type: Number,
      default: 10,
      min: 1,
      max: 100
    }
  },
  
  analytics: {
    enabled: {
      type: Boolean,
      default: false
    },
    provider: {
      type: String,
      enum: ['google', 'mixpanel', 'amplitude']
    },
    trackingId: String,
    apiKey: String,
    collectUsageStats: {
      type: Boolean,
      default: true
    },
    collectErrorLogs: {
      type: Boolean,
      default: true
    }
  },
  
  backup: {
    enabled: {
      type: Boolean,
      default: false
    },
    frequency: {
      type: String,
      enum: ['daily', 'weekly', 'monthly'],
      default: 'daily'
    },
    retention: {
      type: Number,
      default: 30,
      min: 1
    },
    destination: {
      type: String,
      enum: ['local', 's3', 'gcs'],
      default: 'local'
    },
    encryption: {
      type: Boolean,
      default: true
    },
    encryptionKey: String
  },
  
  notifications: {
    enablePush: {
      type: Boolean,
      default: false
    },
    pushProvider: {
      type: String,
      enum: ['firebase', 'pusher', 'websocket']
    },
    firebaseServerKey: String,
    pusherAppId: String,
    pusherKey: String,
    pusherSecret: String,
    pusherCluster: String
  },
  
  rateLimiting: {
    enabled: {
      type: Boolean,
      default: true
    },
    uploadLimit: {
      type: Number,
      default: 10 // per minute
    },
    downloadLimit: {
      type: Number,
      default: 60 // per minute
    },
    apiLimit: {
      type: Number,
      default: 100 // per minute
    },
    shareLimit: {
      type: Number,
      default: 20 // per minute
    }
  },
  
  version: {
    type: String,
    default: '1.0.0'
  },
  lastUpdatedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  }
}, {
  timestamps: true,
  collection: 'settings'
});

// Ensure only one settings document exists
settingsSchema.index({}, { unique: true });

export const Settings = mongoose.models.Settings || mongoose.model<ISettings>('Settings', settingsSchema);
