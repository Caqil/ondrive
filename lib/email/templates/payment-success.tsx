import React from "react";
import { PaymentSuccessProps } from "@/types/email";

export const PaymentSuccessTemplate: React.FC<PaymentSuccessProps> = (
  props
) => {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Payment Successful</title>
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
          .payment-card {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 24px;
            margin: 24px 0;
          }
          .amount {
            font-size: 32px;
            font-weight: 700;
            color: #28a745;
            text-align: center;
            margin: 16px 0;
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
            <h1 className="title">Payment Successful!</h1>
          </div>

          <p>Hi {props.userName},</p>

          <p>Thank you! Your payment has been processed successfully.</p>

          <div className="payment-card">
            <div className="amount">
              {props.amount} {props.currency}
            </div>
            <p style={{ textAlign: "center", margin: 0 }}>
              Payment for {props.planName}
            </p>
          </div>

          <div className="details">
            <h3>Payment Details:</h3>
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
              <span>Invoice Number:</span>
              <span>{props.invoiceNumber}</span>
            </div>
            <div className="detail-row">
              <span>Date:</span>
              <span>{props.date}</span>
            </div>
          </div>

          <p>
            Your subscription is now active and you have full access to all
            features included in your plan.
          </p>

          <div style={{ textAlign: "center" }}>
            <a href={props.receiptUrl} className="button">
              Download Receipt
            </a>
          </div>

          <p>
            You can view your billing history and manage your subscription at
            any time through your account settings.
          </p>

          <p>Thank you for choosing {props.appName}!</p>
          <p>The {props.appName} Team</p>

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
