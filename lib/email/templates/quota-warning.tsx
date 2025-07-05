import React from "react";
import { QuotaWarningProps } from "@/types/email";

export const QuotaWarningTemplate: React.FC<QuotaWarningProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Storage Quota Warning</title>
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
          .warning-icon {
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
          .usage-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
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
            background: linear-gradient(90deg, #ffc107, #e55353);
            height: 100%;
            width: ${props.usagePercentage}%;
            border-radius: 12px;
          }
          .usage-text {
            text-align: center;
            font-weight: 600;
            margin: 12px 0;
            color: #ffc107;
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
          .tips {
            background: #e8f5e8;
            border: 1px solid #c3e6c3;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .tip {
            margin: 8px 0;
            padding-left: 20px;
            position: relative;
          }
          .tip:before {
            content: "💡";
            position: absolute;
            left: 0;
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
            <div className="warning-icon">⚠️</div>
            <h1 className="title">Storage Quota Warning</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} storage is getting full. You're currently using{" "}
            {props.usagePercentage}% of your available storage space.
          </p>

          <div className="usage-card">
            <h3>Storage Usage</h3>
            <div className="usage-bar">
              <div className="usage-fill"></div>
            </div>
            <div className="usage-text">
              {props.usedStorage} of {props.totalStorage} used (
              {props.usagePercentage}%)
            </div>
          </div>

          <p>
            When your storage is full, you won't be able to upload new files. To
            avoid any interruption:
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.upgradeUrl} className="button">
              Upgrade Storage
            </a>
          </div>

          <div className="tips">
            <h3>Tips to free up space:</h3>
            <div className="tip">Delete files you no longer need</div>
            <div className="tip">Empty your trash/deleted files</div>
            <div className="tip">Remove old file versions</div>
            <div className="tip">Compress large files before uploading</div>
            <div className="tip">Move files to external storage</div>
          </div>

          <p>
            You can also upgrade to a plan with more storage to accommodate your
            growing needs.
          </p>

          <p>
            Best regards,
            <br />
            The {props.appName} Team
          </p>

          <div className="footer">
            <p>
              Need help managing your storage? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
