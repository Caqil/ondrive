import React from "react";
import { PasswordResetProps } from "@/types/email";

export const PasswordResetTemplate: React.FC<PasswordResetProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Reset your password</title>
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
          .button { 
            display: inline-block;
            padding: 16px 32px;
            background: ${props.primaryColor || "#007bff"};
            color: white !important;
            text-decoration: none;
            border-radius: 8px;
            margin: 24px 0;
            font-weight: 600;
          }
          .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 16px;
            margin: 24px 0;
            color: #856404;
          }
          .footer { 
            margin-top: 40px;
            padding-top: 24px;
            border-top: 1px solid #e0e0e0;
            font-size: 14px;
            color: #666;
            text-align: center;
          }
        `}</style>
      </head>
      <body>
        <div className="container">
          <div className="header">
            {props.logoUrl && (
              <img src={props.logoUrl} alt={props.appName} className="logo" />
            )}
            <h1 className="title">Reset Your Password</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            We received a request to reset the password for your {props.appName}{" "}
            account.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.resetUrl} className="button">
              Reset Password
            </a>
          </div>

          <div className="warning">
            <strong>Important:</strong> This reset link will expire in{" "}
            {props.expiresIn}. If you didn't request a password reset, please
            ignore this email and your password will remain unchanged.
          </div>

          <p>
            For security reasons, this link can only be used once. If you need
            to reset your password again, please request a new reset link.
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
              <a href={props.resetUrl}>{props.resetUrl}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
