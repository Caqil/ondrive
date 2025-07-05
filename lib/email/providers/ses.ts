import { SESClient, SendEmailCommand } from '@aws-sdk/client-ses';
import type { EmailProvider, EmailMessage, EmailSendResult } from '@/types/email';
import type { EmailSettings } from '@/types';

export class SESProvider implements EmailProvider {
  name = 'Amazon SES';
  private client: SESClient;

  constructor(private config: EmailSettings['ses'], private from: string) {
    this.client = new SESClient({
      region: config.region,
      credentials: {
        accessKeyId: config.accessKeyId,
        secretAccessKey: config.secretAccessKey,
      },
    });
  }

  async send(message: EmailMessage): Promise<EmailSendResult> {
    try {
      const command = new SendEmailCommand({
        Source: message.from || this.from,
        Destination: {
          ToAddresses: Array.isArray(message.to) ? message.to : [message.to],
        },
        Message: {
          Subject: {
            Data: message.subject,
            Charset: 'UTF-8',
          },
          Body: {
            Html: message.html ? {
              Data: message.html,
              Charset: 'UTF-8',
            } : undefined,
            Text: message.text ? {
              Data: message.text,
              Charset: 'UTF-8',
            } : undefined,
          },
        },
        ReplyToAddresses: message.replyTo ? [message.replyTo] : undefined,
      });

      const result = await this.client.send(command);

      return {
        success: true,
        messageId: result.MessageId,
      };
    } catch (error) {
      console.error('SES send error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to send email',
      };
    }
  }

  async verify(): Promise<boolean> {
    try {
      // Try to get account send quota to verify credentials
      const { GetSendQuotaCommand } = await import('@aws-sdk/client-ses');
      await this.client.send(new GetSendQuotaCommand({}));
      return true;
    } catch (error) {
      console.error('SES verification failed:', error);
      return false;
    }
  }
}
