import React from "react";
import { UpgradeNotificationProps } from "@/types/email";

export const UpgradeNotificationTemplate: React.FC<UpgradeNotificationProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Plan Upgraded Successfully</title>
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
          .upgrade-card {
            background: linear-gradient(135deg, #d4edda, #c3e6cb);
            border: 1px solid #28a745;
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
            background: #f8f9fa;
            color: #6c757d;
          }
          .new-plan {
            background: #28a745;
            color: white;
          }
          .arrow {
            font-size: 20px;
            color: #28a745;
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
            content: "✨";
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
            <div className="success-icon">🚀</div>
            <h1 className="title">Plan Upgraded Successfully!</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            Congratulations! Your {props.appName} plan has been successfully
            upgraded. You now have access to enhanced features and capabilities.
          </p>

          <div className="upgrade-card">
            <h3>Plan Upgrade</h3>
            <div className="plan-transition">
              <div className="plan old-plan">{props.oldPlan}</div>
              <div className="arrow">→</div>
              <div className="plan new-plan">{props.newPlan}</div>
            </div>
            <p style={{ margin: "12px 0 0 0" }}>
              Upgraded on {props.upgradeDate}
            </p>
          </div>

          <div className="features">
            <h3>🎉 New Features Now Available:</h3>
            {props.newFeatures.map((feature, index) => (
              <div key={index} className="feature">
                {feature}
              </div>
            ))}
          </div>

          <p>
            Your upgrade is effective immediately, and you can start using all
            the new features right away. Your next billing cycle will reflect
            the new plan pricing.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.manageUrl} className="button">
              Explore New Features
            </a>
          </div>

          <p>
            Thank you for choosing to upgrade! We're excited to help you
            accomplish even more with {props.appName}.
          </p>

          <p>
            If you have any questions about your new features or need help
            getting started, our support team is here to help.
          </p>

          <p>
            Welcome to {props.newPlan}!<br />
            The {props.appName} Team
          </p>

          <div className="footer">
            <p>
              Need help with your new features? Contact us at{" "}
              <a href={`mailto:${props.supportEmail}`}>{props.supportEmail}</a>
            </p>
          </div>
        </div>
      </body>
    </html>
  );
};
