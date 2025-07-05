import crypto from 'crypto';
import bcrypt from 'bcryptjs';
import base32 from 'base32.js';
/**
 * Generate a cryptographically secure random string
 */
export function generateSecretKey(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

/**
 * Generate a secure random token
 */
export function generateToken(length: number = 32): string {
  return crypto.randomBytes(length).toString('base64url');
}

/**
 * Hash a password using bcrypt
 */
export async function hashPassword(password: string): Promise<string> {
  const saltRounds = 12;
  return bcrypt.hash(password, saltRounds);
}

/**
 * Verify a password against its hash
 */
export async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

/**
 * Generate MD5 hash
 */
export function generateMD5(data: string | Buffer): string {
  return crypto.createHash('md5').update(data).digest('hex');
}

/**
 * Generate SHA256 hash
 */
export function generateSHA256(data: string | Buffer): string {
  return crypto.createHash('sha256').update(data).digest('hex');
}

/**
 * Generate file checksum
 */
export function generateFileChecksum(buffer: Buffer, algorithm: 'md5' | 'sha256' = 'sha256'): string {
  return crypto.createHash(algorithm).update(buffer).digest('hex');
}
/**
 * Encrypt data using AES-256-GCM
 */
export function encrypt(data: string, key: string): {
  encrypted: string;
  iv: string;
  tag: string;
} {
  const iv = crypto.randomBytes(16);
  const keyBuffer = Buffer.from(key, 'hex');
  const cipher = crypto.createCipheriv('aes-256-gcm', keyBuffer, iv);
  
  let encrypted = cipher.update(data, 'utf8', 'hex');
  encrypted += cipher.final('hex');
  
  const tag = cipher.getAuthTag();
  
  return {
    encrypted,
    iv: iv.toString('hex'),
    tag: tag.toString('hex'),
  };
}

/**
 * Decrypt data using AES-256-GCM
 */
export function decrypt(encryptedData: {
  encrypted: string;
  iv: string;
  tag: string;
}, key: string): string {
  const iv = Buffer.from(encryptedData.iv, 'hex');
  const keyBuffer = Buffer.from(key, 'hex');
  const decipher = crypto.createDecipheriv('aes-256-gcm', keyBuffer, iv);
  decipher.setAuthTag(Buffer.from(encryptedData.tag, 'hex'));
  
  let decrypted = decipher.update(encryptedData.encrypted, 'hex', 'utf8');
  decrypted += decipher.final('utf8');
  
  return decrypted;
}

/**
 * Generate TOTP (Time-based One-Time Password)
 */


export function generateTOTP(secret: string, timeStep: number = 30): string {
  const time = Math.floor(Date.now() / 1000 / timeStep);
  const decoder = new base32.Decoder();
  const secretBuffer = Buffer.from(decoder.write(secret).finalize());
  const hmac = crypto.createHmac('sha1', secretBuffer);

  const timeBuffer = Buffer.alloc(8);
  timeBuffer.writeUInt32BE(Math.floor(time / 0x100000000), 0);
  timeBuffer.writeUInt32BE(time & 0xffffffff, 4);

  const hash = hmac.update(timeBuffer).digest();
  const offset = hash[hash.length - 1] & 0xf;

  const code = ((hash[offset] & 0x7f) << 24) |
    ((hash[offset + 1] & 0xff) << 16) |
    ((hash[offset + 2] & 0xff) << 8) |
    (hash[offset + 3] & 0xff);

  return (code % 1000000).toString().padStart(6, '0');
}

/**
 * Verify TOTP code
 */
export function verifyTOTP(secret: string, token: string, window: number = 1): boolean {
  const timeStep = 30;
  const currentTime = Math.floor(Date.now() / 1000 / timeStep);
  
  for (let i = -window; i <= window; i++) {
    const testTime = currentTime + i;
    const testToken = generateTOTP(secret, timeStep);
    
    if (testToken === token) {
      return true;
    }
  }
  
  return false;
}

/**
 * Generate API key
 */
export function generateAPIKey(prefix: string = 'dk'): string {
  const randomPart = generateSecretKey(16);
  return `${prefix}_${randomPart}`;
}

/**
 * Hash API key for storage
 */
export function hashAPIKey(apiKey: string): string {
  return generateSHA256(apiKey);
}
