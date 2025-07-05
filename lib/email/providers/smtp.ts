import nodemailer from 'nodemailer';
import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';

export class SMTPProvider implements EmailProvider {
  name = 'SMTP';
  private transporter: nodemailer.Transporter;

  constructor(private config: EmailSettings['smtp'], private from: string) {
    this.transporter = nodemailer.createTransporter({
      host: config.host,
      port: config.port,
      secure: config.secure,
      auth: config.auth ? {
        user: config.auth.user,
        pass: config.auth.pass,
      } : undefined,
      tls: {
        rejectUnauthorized: false,
      },
    });
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    try {
      const result = await this.transporter.sendMail({
        from: message.from || this.from,
        to: Array.isArray(message.to) ? message.to.join(', ') : message.to,
        replyTo: message.replyTo,
        subject: message.subject,
        html: message.html,
        text: message.text,
        attachments: message.attachments?.map(att => ({
          filename: att.filename,
          content: att.content,
          contentType: att.contentType,
          cid: att.cid,
        })),
        headers: message.headers,
      });

      return {
        success: true,
        messageId: result.messageId,
      };
    } catch (error) {
      console.error('SMTP send error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to send email',
      };
    }
  }

  async verify(): Promise<boolean> {
    try {
      await this.transporter.verify();
      return true;
    } catch (error) {
      console.error('SMTP verification failed:', error);
      return false;
    }
  }
}