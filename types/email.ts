import { ReactNode } from 'react';

export interface EmailTemplate {
  subject: string;
  html: string;
  text: string;
}

export interface EmailMessage {
  to: string | string[];
  from?: string;
  replyTo?: string;
  subject: string;
  html?: string;
  text?: string;
  attachments?: EmailAttachment[];
  headers?: Record<string, string>;
}

export interface EmailAttachment {
  filename: string;
  content: Buffer | string;
  contentType?: string;
  cid?: string;
}

export interface EmailProvider {
  name: string;
  send: (message: EmailMessage) => Promise<EmailSendResult>;
  verify: () => Promise<boolean>;
}

export interface EmailSendResult {
  success: boolean;
  messageId?: string;
  error?: string;
}

export interface EmailTemplateProps {
  appName: string;
  logoUrl?: string;
  primaryColor?: string;
  supportEmail: string;
  unsubscribeUrl?: string;
}

// Template-specific props interfaces
export interface WelcomeEmailProps extends EmailTemplateProps {
  userName: string;
  loginUrl: string;
}

export interface EmailVerificationProps extends EmailTemplateProps {
  userName: string;
  verificationUrl: string;
  verificationCode: string;
}

export interface PasswordResetProps extends EmailTemplateProps {
  userName: string;
  resetUrl: string;
  expiresIn: string;
}

export interface PasswordChangedProps extends EmailTemplateProps {
  userName: string;
  changeTime: string;
  ipAddress: string;
  userAgent: string;
}

export interface LoginAlertProps extends EmailTemplateProps {
  userName: string;
  loginTime: string;
  ipAddress: string;
  userAgent: string;
  location?: string;
}

export interface ShareNotificationProps extends EmailTemplateProps {
  userName: string;
  senderName: string;
  fileName: string;
  shareUrl: string;
  message?: string;
  expiresAt?: string;
}

export interface TeamInvitationProps extends EmailTemplateProps {
  userName: string;
  inviterName: string;
  teamName: string;
  inviteUrl: string;
  message?: string;
}

export interface SubscriptionConfirmationProps extends EmailTemplateProps {
  userName: string;
  planName: string;
  amount: string;
  currency: string;
  billingCycle: string;
  nextBillingDate: string;
  manageUrl: string;
}

export interface PaymentSuccessProps extends EmailTemplateProps {
  userName: string;
  amount: string;
  currency: string;
  planName: string;
  invoiceNumber: string;
  date: string;
  receiptUrl: string;
}

export interface PaymentFailedProps extends EmailTemplateProps {
  userName: string;
  amount: string;
  currency: string;
  planName: string;
  retryUrl: string;
  reason?: string;
}

export interface QuotaWarningProps extends EmailTemplateProps {
  userName: string;
  usagePercentage: number;
  usedStorage: string;
  totalStorage: string;
  upgradeUrl: string;
}

export interface QuotaExceededProps extends EmailTemplateProps {
  userName: string;
  usedStorage: string;
  totalStorage: string;
  upgradeUrl: string;
  cleanupUrl: string;
}

export interface SubscriptionExpiringProps extends EmailTemplateProps {
  userName: string;
  planName: string;
  expiresAt: string;
  renewUrl: string;
}

export interface SubscriptionExpiredProps extends EmailTemplateProps {
  userName: string;
  planName: string;
  expiredAt: string;
  renewUrl: string;
}

export interface SubscriptionCancelledProps extends EmailTemplateProps {
  userName: string;
  planName: string;
  cancelledAt: string;
  accessUntil: string;
  reactivateUrl: string;
}

export interface UpgradeNotificationProps extends EmailTemplateProps {
  userName: string;
  oldPlan: string;
  newPlan: string;
  upgradeDate: string;
  newFeatures: string[];
  manageUrl: string;
}

export interface DowngradeNotificationProps extends EmailTemplateProps {
  userName: string;
  oldPlan: string;
  newPlan: string;
  downgradeDate: string;
  removedFeatures: string[];
  manageUrl: string;
}
