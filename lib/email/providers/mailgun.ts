import Mailgun from 'mailgun.js';
import formData from 'form-data';
import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';

export class MailgunProvider implements EmailProvider {
  name = 'Mailgun';
  private mg: any;

  constructor(private config: EmailSettings['mailgun'], private from: string) {
    const mailgun = new Mailgun(formData);
    this.mg = mailgun.client({
      username: 'api',
      key: config.apiKey,
    });
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    try {
      const messageData = {
        from: message.from || this.from,
        to: Array.isArray(message.to) ? message.to.join(', ') : message.to,
        'h:Reply-To': message.replyTo,
        subject: message.subject,
        html: message.html,
        text: message.text,
        attachment: message.attachments?.map(att => ({
          filename: att.filename,
          data: att.content,
          contentType: att.contentType,
        })),
        ...message.headers,
      };

      const result = await this.mg.messages.create(this.config.domain, messageData);

      return {
        success: true,
        messageId: result.id,
      };
    } catch (error) {
      console.error('Mailgun send error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to send email',
      };
    }
  }

  async verify(): Promise<boolean> {
    try {
      await this.mg.domains.get(this.config.domain);
      return true;
    } catch (error) {
      console.error('Mailgun verification failed:', error);
      return false;
    }
  }
}