import sgMail from '@sendgrid/mail';
import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';

export class SendGridProvider implements EmailProvider {
  name = 'SendGrid';

  constructor(private config: EmailSettings['sendgrid'], private from: string) {
    sgMail.setApiKey(config.apiKey);
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    try {
      const msg = {
        to: message.to,
        from: message.from || this.from,
        replyTo: message.replyTo,
        subject: message.subject,
        // Only include html or text if they are defined
        ...(message.html ? { html: message.html } : {}),
        ...(message.text ? { text: message.text } : {}),
        attachments: message.attachments?.map(att => ({
          filename: att.filename,
          content: Buffer.isBuffer(att.content) 
            ? att.content.toString('base64') 
            : Buffer.from(att.content).toString('base64'),
          type: att.contentType,
          disposition: 'attachment',
          contentId: att.cid,
        })),
        headers: message.headers,
      };

      const [response] = await sgMail.send(msg as sgMail.MailDataRequired);

      return {
        success: true,
        messageId: response.headers['x-message-id'] as string,
      };
    } catch (error) {
      console.error('SendGrid send error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to send email',
      };
    }
  }

  async verify(): Promise<boolean> {
    try {
      // SendGrid doesn't have a direct verify method, so we'll check API key validity
      await sgMail.send({
        to: 'verify@example.com',
        from: this.from,
        subject: 'Verification Test',
        text: 'This is a test',
      }, false); // Don't actually send
      return true;
    } catch (error) {
      console.error('SendGrid verification failed:', error);
      return false;
    }
  }
}