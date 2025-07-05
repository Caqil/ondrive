import { Resend } from 'resend';
import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';

export class ResendProvider implements EmailProvider {
  name = 'Resend';
  private resend: Resend;

  constructor(private config: EmailSettings['resend'], private from: string) {
    this.resend = new Resend(config.apiKey);
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    try {
      const result = await this.resend.emails.send({
        from: message.from || this.from,
        to: Array.isArray(message.to) ? message.to : [message.to],
        replyTo: message.replyTo,
        subject: message.subject,
        html: message.html,
        text: message.text || '',
        attachments: message.attachments?.map(att => ({
          filename: att.filename,
          content: att.content,
          contentType: att.contentType,
        })),
        headers: message.headers,
      });

      return {
        success: true,
        messageId: result.data?.id,
      };
    } catch (error) {
      console.error('Resend send error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to send email',
      };
    }
  }

  async verify(): Promise<boolean> {
    try {
      // Resend doesn't have a direct verify method, but we can check if the API key works
      await this.resend.domains.list();
      return true;
    } catch (error) {
      console.error('Resend verification failed:', error);
      return false;
    }
  }
}