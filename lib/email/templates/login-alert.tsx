import React from "react";
import { LoginAlertProps } from "@/types/email";

export const LoginAlertTemplate: React.FC<LoginAlertProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>New Sign-in to Your Account</title>
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
            color: #ffc107;
            font-size: 24px;
            margin: 0 0 16px 0;
          }
          .alert-icon {
            background: #fff3cd;
            border: 2px solid #ffc107;
            border-radius: 50%;
            width: 60px;
            height: 60px;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 30px;
            color: #ffc107;
          }
          .details {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .detail-row {
            margin: 8px 0;
            font-size: 14px;
          }
          .warning {
            background: #f8d7da;
            border: 1px solid #dc3545;
            border-radius: 6px;
            padding: 16px;
            margin: 24px 0;
            color: #721c24;
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
            <div className="alert-icon">🔒</div>
            <h1 className="title">New Sign-in Detected</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>We detected a new sign-in to your {props.appName} account.</p>

          <div className="details">
            <h3>Sign-in Details:</h3>
            <div className="detail-row">
              <strong>Date & Time:</strong> {props.loginTime}
            </div>
            <div className="detail-row">
              <strong>IP Address:</strong> {props.ipAddress}
            </div>
            <div className="detail-row">
              <strong>Browser:</strong> {props.userAgent}
            </div>
            {props.location && (
              <div className="detail-row">
                <strong>Location:</strong> {props.location}
              </div>
            )}
          </div>

          <p>If this was you, you can safely ignore this email.</p>

          <div className="warning">
            <strong>Suspicious activity?</strong> If you don't recognize this
            sign-in, please:
            <ul>
              <li>Change your password immediately</li>
              <li>Enable two-factor authentication</li>
              <li>
                Contact our support team at{" "}
                <a href={`mailto:${props.supportEmail}`}>
                  {props.supportEmail}
                </a>
              </li>
            </ul>
          </div>

          <p>To help keep your account secure, we recommend:</p>
          <ul>
            <li>Using a strong, unique password</li>
            <li>Enabling two-factor authentication</li>
            <li>Regularly reviewing your account activity</li>
            <li>Signing out of shared or public devices</li>
          </ul>

          <p>
            Best regards,
            <br />
            The {props.appName} Security Team
          </p>

          <div className="footer">
            <p>
              Need help? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
