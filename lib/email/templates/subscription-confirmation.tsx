import React from "react";
import { SubscriptionConfirmationProps } from "@/types/email";

export const SubscriptionConfirmationTemplate: React.FC<
  SubscriptionConfirmationProps
> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Subscription Confirmed</title>
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
          .plan-card {
            background: linear-gradient(135deg, ${
              props.primaryColor || "#007bff"
            }, ${props.primaryColor ? `${props.primaryColor}dd` : "#0056b3"});
            color: white;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
            text-align: center;
          }
          .plan-name {
            font-size: 24px;
            font-weight: 700;
            margin-bottom: 8px;
          }
          .plan-price {
            font-size: 32px;
            font-weight: 600;
            margin: 8px 0;
          }
          .plan-cycle {
            opacity: 0.9;
            font-size: 16px;
          }
          .details {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
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
            font-weight: 600;
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
            <h1 className="title">Subscription Confirmed!</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Thank you for subscribing to {props.appName}! Your subscription has
            been confirmed and is now active.
          </p>

          <div className="plan-card">
            <div className="plan-name">{props.planName}</div>
            <div className="plan-price">
              {props.amount} {props.currency}
            </div>
            <div className="plan-cycle">per {props.billingCycle}</div>
          </div>

          <div className="details">
            <h3>Subscription Details:</h3>
            <div className="detail-row">
              <span>Plan:</span>
              <span>{props.planName}</span>
            </div>
            <div className="detail-row">
              <span>Amount:</span>
              <span>
                {props.amount} {props.currency}
              </span>
            </div>
            <div className="detail-row">
              <span>Billing Cycle:</span>
              <span>{props.billingCycle}</span>
            </div>
            <div className="detail-row">
              <span>Next Billing Date:</span>
              <span>{props.nextBillingDate}</span>
            </div>
          </div>

          <p>
            You now have access to all premium features included in your plan.
            You can manage your subscription, update payment methods, or view
            billing history at any time.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.manageUrl} className="button">
              Manage Subscription
            </a>
          </div>

          <p>
            If you have any questions about your subscription or need help
            getting started, don't hesitate to reach out to our support team.
          </p>

          <p>
            Welcome to {props.planName}!<br />
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
