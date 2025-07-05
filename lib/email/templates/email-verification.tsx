import React from "react";
import { EmailVerificationProps } from "@/types/email";

export const EmailVerificationTemplate: React.FC<EmailVerificationProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Verify your email address</title>
        <style>{`
          body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f8f9fa;
          }
          .container {
            background: white;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
          }
          .header { 
            text-align: center; 
            margin-bottom: 32px;
          }
          .logo { max-width: 150px; height: auto; margin-bottom: 20px; }
          .title {
            color: ${props.primaryColor || "#007bff"};
            font-size: 24px;
            margin: 0 0 16px 0;
          }
          .verification-code {
            background: #f8f9fa;
            border: 2px solid ${props.primaryColor || "#007bff"};
            border-radius: 8px;
            padding: 20px;
            font-family: monospace;
            font-size: 24px;
            font-weight: bold;
            text-align: center;
            letter-spacing: 4px;
            margin: 24px 0;
            color: ${props.primaryColor || "#007bff"};
          }
          .button { 
            display: inline-block;
            padding: 16px 32px;
            background: ${props.primaryColor || "#007bff"};
            color: white !important;
            text-decoration: none;
            border-radius: 8px;
            margin: 24px 0;
            font-weight: 600;
            text-align: center;
          }
          .footer { 
            margin-top: 40px;
            padding-top: 24px;
            border-top: 1px solid #e0e0e0;
            font-size: 14px;
            color: #666;
            text-align: center;
          }
          .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 16px;
            margin: 24px 0;
            color: #856404;
          }
        `}</style>
      </head>
      <body>
        <div className="container">
          <div className="header">
            {props.logoUrl && (
              <img src={props.logoUrl} alt={props.appName} className="logo" />
            )}
            <h1 className="title">Verify Your Email Address</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Thank you for signing up for {props.appName}! To complete your
            registration and activate your account, please verify your email
            address.
          </p>

          <p>You can verify your email in two ways:</p>

          <p>
            <strong>Option 1:</strong> Click the verification button below:
          </p>
          <div style={{ textAlign: "center" }}>
            <a href={props.verificationUrl} className="button">
              Verify Email Address
            </a>
          </div>

          <p>
            <strong>Option 2:</strong> Use this verification code:
          </p>
          <div className="verification-code">{props.verificationCode}</div>

          <div className="warning">
            <strong>Security Note:</strong> This verification link will expire
            in 24 hours. If you didn't create an account with {props.appName},
            please ignore this email.
          </div>

          <p>
            Once verified, you'll have full access to all {props.appName}{" "}
            features.
          </p>

          <p>
            Best regards,
            <br />
            The {props.appName} Team
          </p>

          <div className="footer">
            <p>
              Need help? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
            <p>
              If you're having trouble with the button above, copy and paste
              this URL into your browser:
              <br />
              <a href={props.verificationUrl}>{props.verificationUrl}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
