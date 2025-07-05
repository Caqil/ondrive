import React from "react";
import { SubscriptionCancelledProps } from "@/types/email";

export const SubscriptionCancelledTemplate: React.FC<
  SubscriptionCancelledProps
> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Subscription Cancelled</title>
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
            color: #6c757d;
            font-size: 24px;
            margin: 0 0 16px 0;
          }
          .cancelled-icon {
            background: #f8f9fa;
            border: 2px solid #6c757d;
            border-radius: 50%;
            width: 60px;
            height: 60px;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 30px;
            color: #6c757d;
          }
          .cancellation-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
          }
          .details {
            margin: 16px 0;
          }
          .detail-row {
            display: flex;
            justify-content: space-between;
            margin: 8px 0;
            padding: 4px 0;
            border-bottom: 1px solid #e9ecef;
          }
          .detail-row:last-child {
            border-bottom: none;
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
          .access-info {
            background: #d1ecf1;
            border: 1px solid #bee5eb;
            border-radius: 6px;
            padding: 20px;
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
            <div className="cancelled-icon">✓</div>
            <h1 className="title">Subscription Cancelled</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            We've successfully cancelled your {props.appName} subscription as
            requested. We're sorry to see you go!
          </p>

          <div className="cancellation-card">
            <h3>Cancellation Details:</h3>
            <div className="details">
              <div className="detail-row">
                <span>Plan:</span>
                <span>{props.planName}</span>
              </div>
              <div className="detail-row">
                <span>Cancelled On:</span>
                <span>{props.cancelledAt}</span>
              </div>
              <div className="detail-row">
                <span>Access Until:</span>
                <span>{props.accessUntil}</span>
              </div>
            </div>
          </div>

          <div className="access-info">
            <h3>What happens now?</h3>
            <ul>
              <li>
                Your subscription is cancelled, but you'll keep access to
                premium features until {props.accessUntil}
              </li>
              <li>No further charges will be made to your payment method</li>
              <li>All your files and data remain safe and accessible</li>
              <li>
                After {props.accessUntil}, your account will be downgraded to
                the free plan
              </li>
            </ul>
          </div>

          <p>
            <strong>Changed your mind?</strong> You can reactivate your
            subscription at any time before {props.accessUntil} to continue with
            uninterrupted service.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.reactivateUrl} className="button">
              Reactivate Subscription
            </a>
          </div>

          <p>
            We'd love to know how we can improve. If you have a moment, please
            let us know why you cancelled so we can make {props.appName} better
            for everyone.
          </p>

          <p>
            Thank you for being part of the {props.appName} community. We hope
            to see you again soon!
          </p>

          <p>
            Best regards,
            <br />
            The {props.appName} Team
          </p>

          <div className="footer">
            <p>
              Questions? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
