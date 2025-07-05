import React from "react";
import { QuotaExceededProps } from "@/types/email";

export const QuotaExceededTemplate: React.FC<QuotaExceededProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Storage Quota Exceeded</title>
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
            color: #dc3545;
            font-size: 24px;
            margin: 0 0 16px 0;
          }
          .error-icon {
            background: #f8d7da;
            border: 2px solid #dc3545;
            border-radius: 50%;
            width: 60px;
            height: 60px;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 30px;
            color: #dc3545;
          }
          .usage-card {
            background: #f8d7da;
            border: 1px solid #dc3545;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
          }
          .usage-bar {
            background: #e9ecef;
            border-radius: 12px;
            height: 24px;
            margin: 16px 0;
            overflow: hidden;
          }
          .usage-fill {
            background: #dc3545;
            height: 100%;
            width: 100%;
            border-radius: 12px;
          }
          .usage-text {
            text-align: center;
            font-weight: 600;
            margin: 12px 0;
            color: #dc3545;
          }
          .button-primary { 
            display: inline-block;
            padding: 16px 32px;
            background: ${props.primaryColor || "#007bff"};
            color: white !important;
            text-decoration: none;
            border-radius: 8px;
            margin: 12px 8px;
            font-weight: 600;
          }
          .button-secondary { 
            display: inline-block;
            padding: 16px 32px;
            background: #6c757d;
            color: white !important;
            text-decoration: none;
            border-radius: 8px;
            margin: 12px 8px;
            font-weight: 600;
          }
          .actions {
            text-align: center;
            margin: 24px 0;
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
            <div className="error-icon">🚫</div>
            <h1 className="title">Storage Quota Exceeded</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} storage quota has been exceeded. You cannot
            upload new files until you free up space or upgrade your plan.
          </p>

          <div className="usage-card">
            <h3>Storage Usage</h3>
            <div className="usage-bar">
              <div className="usage-fill"></div>
            </div>
            <div className="usage-text">
              {props.usedStorage} of {props.totalStorage} used (100%+)
            </div>
          </div>

          <p>
            <strong>What this means:</strong>
          </p>
          <ul>
            <li>You cannot upload new files</li>
            <li>File sharing may be limited</li>
            <li>Existing files remain accessible</li>
            <li>Account functionality is otherwise normal</li>
          </ul>

          <p>
            <strong>Choose an option to resolve this:</strong>
          </p>

          <div className="actions">
            <a href={props.upgradeUrl} className="button-primary">
              Upgrade Plan
            </a>
            <a href={props.cleanupUrl} className="button-secondary">
              Free Up Space
            </a>
          </div>

          <p>
            <strong>Quick ways to free up space:</strong>
          </p>
          <ul>
            <li>Delete unnecessary files and folders</li>
            <li>Empty your trash completely</li>
            <li>Remove old file versions</li>
            <li>Download and delete large files you don't need online</li>
          </ul>

          <p>
            Upgrading your plan will give you more storage immediately and
            prevent this issue in the future.
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
          </div>
        </div>
      </body>
    </html>
  );
};
