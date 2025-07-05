import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';
import { SMTPProvider } from './providers/smtp';
import { SendGridProvider } from './providers/sendgrid';
import { SESProvider } from './providers/ses';
import { MailgunProvider } from './providers/mailgun';
import { ResendProvider } from './providers/resend';

export class EmailService {
  private provider: EmailProvider;

  constructor(private config: EmailSettings) {
    this.provider = this.createProvider();
  }

  private createProvider(): EmailProvider {
    const { provider, from } = this.config;

    switch (provider) {
      case 'smtp':
        return new SMTPProvider(this.config.smtp, from);
      case 'sendgrid':
        return new SendGridProvider(this.config.sendgrid, from);
      case 'ses':
        return new SESProvider(this.config.ses, from);
      case 'mailgun':
        return new MailgunProvider(this.config.mailgun, from);
      case 'resend':
        return new ResendProvider(this.config.resend, from);
      default:
        throw new Error(`Unsupported email provider: ${provider}`);
    }
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    if (!this.config.enabled) {
      console.warn('Email service is disabled');
      return { success: false, error: 'Email service is disabled' };
    }

    return this.provider.send(message);
  }

  async verify(): Promise<boolean> {
    return this.provider.verify();
  }

  getProviderName(): string {
    return this.provider.name;
  }
}

// Base template component for consistency
export const BaseEmailTemplate = ({ children, ...props }: { children: React.ReactNode } & any) => (
  <html>
    <head>
      <meta charSet="utf-8" />
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <title>{props.subject}</title>
      <style>{`
        body { 
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
          line-height: 1.6;
          color: #333;
          max-width: 600px;
          margin: 0 auto;
          padding: 20px;
        }
        .header { 
          text-align: center; 
          margin-bottom: 32px; 
          padding-bottom: 20px;
          border-bottom: 2px solid #f0f0f0;
        }
        .logo { max-width: 150px; height: auto; }
        .content { margin: 32px 0; }
        .button { 
          display: inline-block;
          padding: 12px 24px;
          background: ${props.primaryColor || '#007bff'};
          color: white;
          text-decoration: none;
          border-radius: 6px;
          margin: 16px 0;
        }
        .footer { 
          margin-top: 48px;
          padding-top: 20px;
          border-top: 1px solid #e0e0e0;
          font-size: 14px;
          color: #666;
          text-align: center;
        }
        .code {
          background: #f8f9fa;
          border: 1px solid #e9ecef;
          border-radius: 4px;
          padding: 12px;
          font-family: monospace;
          font-size: 18px;
          text-align: center;
          letter-spacing: 2px;
          margin: 16px 0;
        }
      `}</style>
    </head>
    <body>
      <div className="header">
        {props.logoUrl && <img src={props.logoUrl} alt={props.appName} className="logo" />}
        <h1>{props.appName}</h1>
      </div>
      <div className="content">
        {children}
      </div>
      <div className="footer">
        <p>
          Need help? Contact us at <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
        </p>
        {props.unsubscribeUrl && (
          <p>
            <a href={props.unsubscribeUrl}>Unsubscribe from these emails</a>
          </p>
        )}
      </div>
    </body>
  </html>
);

// Export all providers and utilities
export * from './providers/smtp';
export * from './providers/sendgrid';
export * from './providers/ses';
export * from './providers/mailgun';
export * from './providers/resend';