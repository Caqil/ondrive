export interface DatabaseConnection {
  url: string;
  name: string;
  options: {
    useNewUrlParser: boolean;
    useUnifiedTopology: boolean;
    maxPoolSize: number;
    serverSelectionTimeoutMS: number;
    socketTimeoutMS: number;
    bufferMaxEntries: number;
    bufferCommands: boolean;
  };
}

export interface DatabaseStats {
  collections: number;
  documents: number;
  dataSize: number;
  indexSize: number;
  storageSize: number;
  avgObjSize: number;
  connections: {
    current: number;
    available: number;
    totalCreated: number;
  };
}

export interface BackupConfig {
  enabled: boolean;
  schedule: string; // cron expression
  retention: number; // days
  destination: 'local' | 's3' | 'gcs';
  encryption: boolean;
  compression: boolean;
}