import { EmailService } from './index';
import type { EmailMessage, EmailTemplate } from '@/types/email';
import type { EmailSettings } from '@/types';
import React from 'react';

// Template registry
export const emailTemplates = {
  welcome: () => import('./templates/welcome'),
  emailVerification: () => import('./templates/email-verification'),
  passwordReset: () => import('./templates/password-reset'),
  passwordChanged: () => import('./templates/password-changed'),
  loginAlert: () => import('./templates/login-alert'),
  shareNotification: () => import('./templates/share-notification'),
  teamInvitation: () => import('./templates/team-invitation'),
  subscriptionConfirmation: () => import('./templates/subscription-confirmation'),
  paymentSuccess: () => import('./templates/payment-success'),
  paymentFailed: () => import('./templates/payment-failed'),
  quotaWarning: () => import('./templates/quota-warning'),
  quotaExceeded: () => import('./templates/quota-exceeded'),
  subscriptionExpiring: () => import('./templates/subscription-expiring'),
  subscriptionExpired: () => import('./templates/subscription-expired'),
  subscriptionCancelled: () => import('./templates/subscription-cancelled'),
  upgradeNotification: () => import('./templates/upgrade-notification'),
  downgradeNotification: () => import('./templates/downgrade-notification'),
} as const;

// Email template renderer
export async function renderEmailTemplate<T extends keyof typeof emailTemplates>(
  templateName: T,
  props: any
): Promise<EmailTemplate> {
  try {
    const module = await emailTemplates[templateName]();
    const TemplateComponent = module[Object.keys(module)[0]];
    
    const { renderToStaticMarkup } = await import('react-dom/server');
    const html = renderToStaticMarkup(React.createElement(TemplateComponent, props));
    
    // Generate text version by stripping HTML
    const text = html
      .replace(/<[^>]*>/g, '')
      .replace(/\s+/g, ' ')
      .trim();
    
    // Extract subject from title or use default
    const subjectMatch = html.match(/<title[^>]*>([^<]+)<\/title>/);
    const subject = subjectMatch ? subjectMatch[1] : `${props.appName} Notification`;
    
    return { subject, html, text };
  } catch (error) {
    console.error(`Failed to render email template ${templateName}:`, error);
    throw new Error(`Email template rendering failed: ${templateName}`);
  }
}

// Email sender utility
export async function sendEmail(
  config: EmailSettings,
  message: EmailMessage
): Promise<boolean> {
  try {
    const emailService = new EmailService(config);
    const result = await emailService.send(message);
    
    if (!result.success) {
      console.error('Email send failed:', result.error);
      return false;
    }
    
    console.log('Email sent successfully:', result.messageId);
    return true;
  } catch (error) {
    console.error('Email service error:', error);
    return false;
  }
}

// Template email sender
export async function sendTemplateEmail<T extends keyof typeof emailTemplates>(
  config: EmailSettings,
  templateName: T,
  to: string | string[],
  templateProps: any
): Promise<boolean> {
  try {
    const template = await renderEmailTemplate(templateName, templateProps);
    
    const message: EmailMessage = {
      to,
      from: config.from,
      replyTo: config.replyTo,
      subject: template.subject,
      html: template.html,
      text: template.text,
    };
    
    return await sendEmail(config, message);
  } catch (error) {
    console.error(`Failed to send template email ${templateName}:`, error);
    return false;
  }
}

// Email validation utility
export function validateEmailAddress(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

// Batch email sender with rate limiting
export async function sendBatchEmails(
  config: EmailSettings,
  messages: EmailMessage[],
  batchSize: number = 10,
  delayMs: number = 1000
): Promise<{ success: number; failed: number }> {
  const emailService = new EmailService(config);
  let success = 0;
  let failed = 0;
  
  for (let i = 0; i < messages.length; i += batchSize) {
    const batch = messages.slice(i, i + batchSize);
    
    const results = await Promise.allSettled(
      batch.map(message => emailService.send(message))
    );
    
    results.forEach(result => {
      if (result.status === 'fulfilled' && result.value.success) {
        success++;
      } else {
        failed++;
      }
    });
    
    // Rate limiting delay
    if (i + batchSize < messages.length) {
      await new Promise(resolve => setTimeout(resolve, delayMs));
    }
  }
  
  return { success, failed };
}

export default EmailService;