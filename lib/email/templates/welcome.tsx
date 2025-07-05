import React from "react";
import { WelcomeEmailProps } from "@/types/email";

export const WelcomeEmailTemplate: React.FC<WelcomeEmailProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Welcome to {props.appName}!</title>
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
          .welcome-title {
            color: ${props.primaryColor || "#007bff"};
            font-size: 28px;
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
            text-align: center;
          }
          .features {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .feature {
            margin: 12px 0;
            padding-left: 24px;
            position: relative;
          }
          .feature:before {
            content: "✓";
            position: absolute;
            left: 0;
            color: ${props.primaryColor || "#007bff"};
            font-weight: bold;
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
            <h1 className="welcome-title">Welcome to {props.appName}!</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Welcome to {props.appName}! We're excited to have you on board. Your
            account has been successfully created and you're ready to start
            uploading, sharing, and managing your files.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.loginUrl} className="button">
              Get Started
            </a>
          </div>

          <div className="features">
            <h3>What you can do with {props.appName}:</h3>
            <div className="feature">
              Upload and organize your files securely
            </div>
            <div className="feature">
              Share files with team members and collaborators
            </div>
            <div className="feature">
              Access your files from anywhere, anytime
            </div>
            <div className="feature">
              Collaborate with advanced sharing permissions
            </div>
            <div className="feature">
              Keep track of file versions and changes
            </div>
          </div>

          <p>
            If you have any questions or need help getting started, don't
            hesitate to reach out to our support team.
          </p>

          <p>Welcome aboard!</p>
          <p>The {props.appName} Team</p>

          <div className="footer">
            <p>
              Need help? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
            {props.unsubscribeUrl && (
              <p>
                <a href={props.unsubscribeUrl}>Unsubscribe from these emails</a>
              </p>
            )}
          </div>
        </div>
      </body>
    </html>
  );
};
