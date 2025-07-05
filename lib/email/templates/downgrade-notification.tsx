import React from "react";
import { DowngradeNotificationProps } from "@/types/email";

export const DowngradeNotificationTemplate: React.FC<
  DowngradeNotificationProps
> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Plan Downgraded</title>
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
          .info-icon {
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
          .downgrade-card {
            background: #fff3cd;
            border: 1px solid #ffc107;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
            text-align: center;
          }
          .plan-transition {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 16px;
            margin: 16px 0;
          }
          .plan {
            padding: 12px 20px;
            border-radius: 6px;
            font-weight: 600;
          }
          .old-plan {
            background: #6c757d;
            color: white;
          }
          .new-plan {
            background: #ffc107;
            color: #333;
          }
          .arrow {
            font-size: 20px;
            color: #ffc107;
          }
          .removed-features {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 20px;
            margin: 24px 0;
          }
          .removed-feature {
            margin: 12px 0;
            padding-left: 24px;
            position: relative;
            color: #856404;
          }
          .removed-feature:before {
            content: "⚠️";
            position: absolute;
            left: 0;
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
            <div className="info-icon">📋</div>
            <h1 className="title">Plan Downgraded</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Your {props.appName} plan has been downgraded as requested. The
            changes have taken effect immediately.
          </p>

          <div className="downgrade-card">
            <h3>Plan Change</h3>
            <div className="plan-transition">
              <div className="plan old-plan">{props.oldPlan}</div>
              <div className="arrow">→</div>
              <div className="plan new-plan">{props.newPlan}</div>
            </div>
            <p style={{ margin: "12px 0 0 0" }}>
              Downgraded on {props.downgradeDate}
            </p>
          </div>

          <div className="removed-features">
            <h3>⚠️ Features No Longer Available:</h3>
            {props.removedFeatures.map((feature, index) => (
              <div key={index} className="removed-feature">
                {feature}
              </div>
            ))}
          </div>

          <p>
            <strong>Important Notes:</strong>
          </p>
          <ul>
            <li>Your files and data remain safe and accessible</li>
            <li>
              You may need to manage your storage if it exceeds the new limit
            </li>
            <li>Your next billing cycle will reflect the new plan pricing</li>
            <li>You can upgrade again at any time</li>
          </ul>

          <p>
            If you're finding that you need more features than your current plan
            offers, you can upgrade at any time to regain access to premium
            functionality.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.manageUrl} className="button">
              Manage Plan
            </a>
          </div>

          <p>
            Thank you for continuing to use {props.appName}. If you have any
            questions about your downgraded plan or need assistance, please
            don't hesitate to contact our support team.
          </p>

          <p>
            Best regards,
            <br />
            The {props.appName} Team
          </p>

          <div className="footer">
            <p>
              Questions about your plan? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
