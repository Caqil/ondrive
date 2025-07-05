import { BaseDocument, ObjectId } from ".";

export interface NotificationChannels {
  inApp: boolean;
  email: boolean;
  push: boolean;
}

export interface NotificationDeliveryStatus {
  inApp: 'pending' | 'delivered' | 'failed';
  email: 'pending' | 'sent' | 'delivered' | 'failed';
  push: 'pending' | 'sent' | 'delivered' | 'failed';
}

export interface Notification extends BaseDocument {
  user: ObjectId;
  type: 'share_received' | 'share_accepted' | 'file_uploaded' | 'file_deleted' | 'team_invite' | 'payment_success' | 'payment_failed' | 'storage_limit' | 'trial_ending' | 'subscription_cancelled' | 'security_alert';
  title: string;
  message: string;
  relatedResource?: ObjectId;
  relatedResourceType?: 'file' | 'folder' | 'share' | 'team' | 'subscription' | 'payment';
  actionUrl?: string;
  actionText?: string;
  isRead: boolean;
  readAt?: Date;
  channels: NotificationChannels;
  deliveryStatus: NotificationDeliveryStatus;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  metadata: Record<string, any>;
  expiresAt?: Date;
}

export interface CreateNotificationRequest {
  userId: ObjectId;
  type: Notification['type'];
  title: string;
  message: string;
  relatedResourceId?: ObjectId;
  relatedResourceType?: string;
  actionUrl?: string;
  actionText?: string;
  channels?: Partial<NotificationChannels>;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
  metadata?: Record<string, any>;
}
