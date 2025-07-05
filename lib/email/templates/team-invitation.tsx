import React from "react";
import { TeamInvitationProps } from "@/types/email";

export const TeamInvitationTemplate: React.FC<TeamInvitationProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Team Invitation</title>
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
          .team-icon {
            background: #e8f5e8;
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
          .team-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
            text-align: center;
          }
          .team-name {
            font-size: 20px;
            font-weight: 600;
            color: ${props.primaryColor || "#007bff"};
            margin-bottom: 12px;
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
          .benefits {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .benefit {
            margin: 8px 0;
            padding-left: 20px;
            position: relative;
          }
          .benefit:before {
            content: "✓";
            position: absolute;
            left: 0;
            color: #28a745;
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
            <div className="team-icon">👥</div>
            <h1 className="title">You're Invited to Join a Team!</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            <strong>{props.inviterName}</strong> has invited you to join their
            team on {props.appName}.
          </p>

          <div className="team-card">
            <div className="team-name">{props.teamName}</div>
            <p>Join this team to collaborate and share files together.</p>
            <a href={props.inviteUrl} className="button">
              Accept Invitation
            </a>
          </div>

          {props.message && (
            <div className="message-box">
              <strong>Personal message from {props.inviterName}:</strong>
              <br />"{props.message}"
            </div>
          )}

          <div className="benefits">
            <h3>Team collaboration benefits:</h3>
            <div className="benefit">
              Share files and folders with team members
            </div>
            <div className="benefit">Collaborate on projects in real-time</div>
            <div className="benefit">Manage team permissions and access</div>
            <div className="benefit">Track team activity and file changes</div>
            <div className="benefit">
              Centralized team storage and organization
            </div>
          </div>

          <p>
            Click the "Accept Invitation" button above to join the team. If you
            don't have a {props.appName} account yet, you'll be able to create
            one during the process.
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
              <a href={props.inviteUrl}>{props.inviteUrl}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
