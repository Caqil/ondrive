import React from "react";
import { SubscriptionExpiringProps } from "@/types/email";

export const SubscriptionExpiringTemplate: React.FC<
  SubscriptionExpiringProps
> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Subscription Expiring Soon</title>
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
          .expiry-card {
            background: linear-gradient(135deg, #fff3cd, #ffeaa7);
            border: 1px solid #ffc107;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
            text-align: center;
          }
          .plan-name {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 8px;
          }
          .expiry-date {
            font-size: 24px;
            font-weight: 700;
            color: #ffc107;
            margin: 12px 0;
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
          .features-lost {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .feature {
            margin: 8px 0;
            padding-left: 20px;
            position: relative;
          }
          .feature:before {
            content: "⚠️";
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
            <div className="warning-icon">⏰</div>
            <h1 className="title">Subscription Expiring Soon</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} subscription is expiring soon. Don't miss out
            on your premium features!
          </p>

          <div className="expiry-card">
            <div className="plan-name">{props.planName}</div>
            <div>Expires on</div>
            <div className="expiry-date">{props.expiresAt}</div>
          </div>

          <p>
            To continue enjoying uninterrupted access to all premium features,
            please renew your subscription before it expires.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.renewUrl} className="button">
              Renew Subscription
            </a>
          </div>

          <div className="features-lost">
            <h3>What you'll lose if your subscription expires:</h3>
            <div className="feature">Access to premium storage capacity</div>
            <div className="feature">
              Advanced sharing and collaboration features
            </div>
            <div className="feature">Priority customer support</div>
            <div className="feature">Enhanced security features</div>
            <div className="feature">API access and integrations</div>
          </div>

          <p>
            <strong>Don't worry!</strong> Your files and data will remain safe.
            You'll just lose access to premium features until you renew.
          </p>

          <p>
            Renew now to maintain your current plan and pricing. Special renewal
            rates may not be available later.
          </p>

          <p>Thank you for being a valued {props.appName} member!</p>
          <p>The {props.appName} Team</p>

          <div className="footer">
            <p>
              Questions about renewal? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
