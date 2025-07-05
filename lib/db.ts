import mongoose from 'mongoose';
import type { ConnectOptions } from 'mongoose';

// Connection state
interface ConnectionState {
  isConnected: boolean;
  promise: Promise<typeof mongoose> | null;
}

const connection: ConnectionState = {
  isConnected: false,
  promise: null,
};

/**
 * Connect to MongoDB database
 */
export async function connectDB(): Promise<typeof mongoose> {
  if (connection.isConnected) {
    return mongoose;
  }

  if (connection.promise) {
    return connection.promise;
  }

  const MONGODB_URI = process.env.MONGODB_URI;
  if (!MONGODB_URI) {
    throw new Error('Please define the MONGODB_URI environment variable');
  }

  const options: ConnectOptions = {
    maxPoolSize: 10, // Maximum number of connections
    serverSelectionTimeoutMS: 5000, // Keep trying to send operations for 5 seconds
    socketTimeoutMS: 45000, // Close sockets after 45 seconds of inactivity
  };

  connection.promise = mongoose.connect(MONGODB_URI, options).then((mongoose) => {
    connection.isConnected = true;
    console.log('Connected to MongoDB');
    return mongoose;
  });

  return connection.promise;
}

/**
 * Disconnect from MongoDB
 */
export async function disconnectDB(): Promise<void> {
  if (connection.isConnected) {
    await mongoose.disconnect();
    connection.isConnected = false;
    connection.promise = null;
    console.log('Disconnected from MongoDB');
  }
}

/**
 * Get database connection status
 */
export function getConnectionStatus(): {
  isConnected: boolean;
  readyState: number;
  host?: string;
  name?: string;
} {
  return {
    isConnected: connection.isConnected,
    readyState: mongoose.connection.readyState,
    host: mongoose.connection.host,
    name: mongoose.connection.name,
  };
}

/**
 * Health check for database
 */
export async function healthCheck(): Promise<{
  status: 'healthy' | 'unhealthy';
  latency: number;
  error?: string;
}> {
  const startTime = Date.now();
  
  try {
    if (!mongoose.connection.db) {
      throw new Error('Database connection is not established');
    }
    await mongoose.connection.db.admin().ping();
    const latency = Date.now() - startTime;
    
    return {
      status: 'healthy',
      latency,
    };
  } catch (error) {
    const latency = Date.now() - startTime;
    
    return {
      status: 'unhealthy',
      latency,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

/**
 * Database statistics
 */
export async function getDatabaseStats(): Promise<{
  collections: number;
  documents: number;
  dataSize: number;
  indexSize: number;
  storageSize: number;
}> {
  try {
    if (!mongoose.connection.db) {
      throw new Error('Database connection is not established');
    }
    const stats = await mongoose.connection.db.stats();
    
    return {
      collections: stats.collections,
      documents: stats.objects,
      dataSize: stats.dataSize,
      indexSize: stats.indexSize,
      storageSize: stats.storageSize,
    };
  } catch (error) {
    throw new Error(`Failed to get database stats: ${error}`);
  }
}

/**
 * Backup database
 */
export async function backupDatabase(outputPath: string): Promise<void> {
  // This would typically use mongodump or similar
  throw new Error('Database backup not implemented');
}

/**
 * Restore database
 */
export async function restoreDatabase(backupPath: string): Promise<void> {
  // This would typically use mongorestore or similar
  throw new Error('Database restore not implemented');
}