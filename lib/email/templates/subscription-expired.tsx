import React from "react";
import { SubscriptionExpiredProps } from "@/types/email";

export const SubscriptionExpiredTemplate: React.FC<SubscriptionExpiredProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Subscription Expired</title>
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
          .expired-icon {
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
          .expired-card {
            background: #f8d7da;
            border: 1px solid #dc3545;
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
          .expired-date {
            font-size: 18px;
            font-weight: 600;
            color: #dc3545;
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
          .limitations {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .limitation {
            margin: 8px 0;
            padding-left: 20px;
            position: relative;
          }
          .limitation:before {
            content: "❌";
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
            <div className="expired-icon">⏰</div>
            <h1 className="title">Subscription Expired</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} subscription has expired. Your account has been
            downgraded to the free plan.
          </p>

          <div className="expired-card">
            <div className="plan-name">{props.planName}</div>
            <div>Expired on</div>
            <div className="expired-date">{props.expiredAt}</div>
          </div>

          <p>
            <strong>Don't worry!</strong> Your files and data are safe and
            secure. However, your account now has limited functionality.
          </p>

          <div className="limitations">
            <h3>Current Account Limitations:</h3>
            <div className="limitation">Reduced storage capacity</div>
            <div className="limitation">Limited file sharing options</div>
            <div className="limitation">No access to premium features</div>
            <div className="limitation">Standard support only</div>
            <div className="limitation">No API access</div>
          </div>

          <p>
            To restore full functionality and regain access to all premium
            features, please renew your subscription.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.renewUrl} className="button">
              Renew Subscription
            </a>
          </div>

          <p>
            <strong>Special Offer:</strong> Renew within 30 days to get back
            your previous plan at the same rate. After 30 days, current pricing
            will apply.
          </p>

          <p>
            We'd love to have you back as a premium member. If you have any
            questions about renewal options or need help choosing a plan, our
            team is here to assist.
          </p>

          <p>
            Best regards,
            <br />
            The {props.appName} Team
          </p>

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
