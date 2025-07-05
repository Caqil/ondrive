import React from "react";
import { ShareNotificationProps } from "@/types/email";

export const ShareNotificationTemplate: React.FC<ShareNotificationProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>File Shared With You</title>
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
          .share-icon {
            background: #e3f2fd;
            border: 2px solid ${props.primaryColor || "#007bff"};
            border-radius: 50%;
            width: 60px;
            height: 60px;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 30px;
            color: ${props.primaryColor || "#007bff"};
          }
          .file-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            margin: 24px 0;
            text-align: center;
          }
          .file-name {
            font-size: 18px;
            font-weight: 600;
            color: ${props.primaryColor || "#007bff"};
            margin-bottom: 8px;
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
          .message-box {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 16px;
            margin: 24px 0;
            font-style: italic;
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
            <div className="share-icon">📁</div>
            <h1 className="title">File Shared With You</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            <strong>{props.senderName}</strong> has shared a file with you on{" "}
            {props.appName}.
          </p>

          <div className="file-card">
            <div className="file-name">{props.fileName}</div>
            <p>Click the button below to view and download the file.</p>
            <a href={props.shareUrl} className="button">
              View File
            </a>
          </div>

          {props.message && (
            <div className="message-box">
              <strong>Message from {props.senderName}:</strong>
              <br />"{props.message}"
            </div>
          )}

          {props.expiresAt && (
            <p>
              <strong>Note:</strong> This shared file will expire on{" "}
              {props.expiresAt}.
            </p>
          )}

          <p>
            You can view, download, and collaborate on this file through your{" "}
            {props.appName} account.
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
              <a href={props.shareUrl}>{props.shareUrl}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
