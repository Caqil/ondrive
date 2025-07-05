import React from "react";
import { PasswordChangedProps } from "@/types/email";

export const PasswordChangedTemplate: React.FC<PasswordChangedProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Password Changed Successfully</title>
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
            color: #28a745;
            font-size: 24px;
            margin: 0 0 16px 0;
          }
          .success-icon {
            background: #d4edda;
            border: 2px solid #28a745;
            border-radius: 50%;
            width: 60px;
            height: 60px;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 30px;
            color: #28a745;
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
            <div className="success-icon">✓</div>
            <h1 className="title">Password Changed Successfully</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} account password has been successfully changed.
          </p>

          <div className="details">
            <h3>Change Details:</h3>
            <div className="detail-row">
              <strong>Date & Time:</strong> {props.changeTime}
            </div>
            <div className="detail-row">
              <strong>IP Address:</strong> {props.ipAddress}
            </div>
            <div className="detail-row">
              <strong>Browser:</strong> {props.userAgent}
            </div>
          </div>

          <div className="warning">
            <strong>Didn't make this change?</strong> If you didn't change your
            password, please contact our support team immediately at{" "}
            <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a> as
            your account may be compromised.
          </div>

          <p>For your security, we recommend:</p>
          <ul>
            <li>Using a strong, unique password</li>
            <li>Enabling two-factor authentication</li>
            <li>Not sharing your password with anyone</li>
            <li>Logging out of shared devices</li>
          </ul>

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
          </div>
        </div>
      </body>
    </html>
  );
};
