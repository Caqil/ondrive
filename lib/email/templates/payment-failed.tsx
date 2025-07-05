import React from "react";
import { PaymentFailedProps } from "@/types/email";

export const PaymentFailedTemplate: React.FC<PaymentFailedProps> = (props) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Payment Failed</title>
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
          .payment-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
          }
          .amount {
            font-size: 24px;
            font-weight: 600;
            color: #dc3545;
            text-align: center;
            margin: 16px 0;
          }
          .reason-box {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 16px;
            margin: 24px 0;
            color: #856404;
          }
          .button { 
            display: inline-block;
            padding: 16px 32px;
            background: #dc3545;
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
            <div className="error-icon">✗</div>
            <h1 className="title">Payment Failed</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>
            We were unable to process your payment for your {props.appName}{" "}
            subscription.
          </p>

          <div className="payment-card">
            <div className="amount">
              {props.amount} {props.currency}
            </div>
            <p style={{ textAlign: "center", margin: 0 }}>
              Payment for {props.planName}
            </p>
          </div>

          {props.reason && (
            <div className="reason-box">
              <strong>Reason:</strong> {props.reason}
            </div>
          )}

          <p>
            To continue enjoying {props.appName} without interruption, please
            update your payment method or try again.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.retryUrl} className="button">
              Retry Payment
            </a>
          </div>

          <p>
            <strong>What happens next?</strong>
          </p>
          <ul>
            <li>Your account remains active for a grace period</li>
            <li>We'll automatically retry the payment in 3 days</li>
            <li>
              If payment continues to fail, your subscription may be suspended
            </li>
          </ul>

          <p>Common reasons for payment failure:</p>
          <ul>
            <li>Insufficient funds in your account</li>
            <li>Expired or invalid payment method</li>
            <li>Bank blocking the transaction</li>
            <li>Incorrect billing information</li>
          </ul>

          <p>
            If you continue to experience issues, please contact our support
            team for assistance.
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
