import mongoose, { Document, Schema } from 'mongoose';

export interface IInvoice extends Document {
  _id: string;
  
  // Invoice identification
  invoiceNumber: string;
  sequence: number; // Sequential number for the year
  
  // Related entities
  subscription?: mongoose.Types.ObjectId;
  user: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  
  // Invoice details
  type: 'subscription' | 'one_time' | 'credit_note' | 'proforma' | 'recurring';
  description: string;
  
  // Line items
  lineItems: {
    id: string;
    description: string;
    quantity: number;
    unitPrice: number; // in cents
    amount: number; // quantity * unitPrice
    taxRate: number;
    taxAmount: number;
    periodStart?: Date;
    periodEnd?: Date;
    planId?: mongoose.Types.ObjectId;
    metadata?: any;
  }[];
  
  // Amount breakdown
  subtotal: number; // Sum of line items before tax
  taxAmount: number;
  discountAmount: number;
  totalAmount: number; // Final amount due
  amountPaid: number;
  amountDue: number; // totalAmount - amountPaid
  
  // Currency and locale
  currency: string;
  exchangeRate?: number; // If different from base currency
  baseCurrency?: string;
  
  // Tax information
  taxDetails: {
    taxId?: string; // Business tax ID
    taxRegion: string;
    taxType: string; // VAT, GST, Sales Tax
    reverseCharge: boolean; // For B2B EU transactions
    taxExempt: boolean;
    taxExemptReason?: string;
  };
  
  // Billing information
  billingAddress: {
    name: string;
    company?: string;
    line1: string;
    line2?: string;
    city: string;
    state?: string;
    postalCode: string;
    country: string;
    taxId?: string;
  };
  
  // Important dates
  invoiceDate: Date;
  dueDate: Date;
  periodStart?: Date;
  periodEnd?: Date;
  paidAt?: Date;
  voidedAt?: Date;
  sentAt?: Date;
  
  // Status tracking
  status: 'draft' | 'open' | 'paid' | 'void' | 'uncollectible' | 'overdue';
  paymentStatus: 'not_paid' | 'partially_paid' | 'paid' | 'refunded' | 'failed';
  
  // Collection attempts
  collectionAttempts: {
    attemptNumber: number;
    attemptedAt: Date;
    method: 'auto' | 'manual';
    result: 'succeeded' | 'failed' | 'pending';
    failureReason?: string;
    nextAttempt?: Date;
  }[];
  
  // Payment information
  payments: mongoose.Types.ObjectId[]; // References to Payment documents
  
  // Provider integration
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy' | 'manual';
  providerId?: string; // External invoice ID
  providerUrl?: string; // Hosted invoice URL
  
  // Applied discounts
  appliedCoupons: {
    couponId: mongoose.Types.ObjectId;
    code: string;
    discountAmount: number;
    discountType: 'percentage' | 'fixed_amount';
  }[];
  
  // Credit notes (if this is a credit note)
  creditNote?: {
    originalInvoiceId: mongoose.Types.ObjectId;
    reason: string;
    type: 'full' | 'partial';
  };
  
  // Files and documents
  files: {
    type: 'pdf' | 'xml' | 'csv';
    url: string;
    size: number;
    generatedAt: Date;
  }[];
  
  // Communication history
  communications: {
    type: 'sent' | 'viewed' | 'downloaded' | 'reminder' | 'overdue_notice';
    channel: 'email' | 'sms' | 'postal' | 'portal';
    sentAt: Date;
    deliveredAt?: Date;
    openedAt?: Date;
    metadata?: any;
  }[];
  
  // Dunning and collection
  dunning: {
    enabled: boolean;
    level: number; // 0 = no dunning, 1-3 = escalation levels
    lastNoticeAt?: Date;
    nextNoticeAt?: Date;
    finalNoticeAt?: Date;
    collectionAgencyAt?: Date;
  };
  
  // Legal and compliance
  legalEntity?: {
    name: string;
    address: any;
    taxId: string;
    registrationNumber: string;
  };
  
  // Notes and internal tracking
  notes: string;
  internalNotes: string;
  tags: string[];
  
  // Metadata for additional information
  metadata: {
    source?: string; // web, api, manual, etc.
    campaign?: string;
    salesRep?: string;
    paymentTerms?: string;
    purchaseOrder?: string;
    [key: string]: any;
  };
  
  // Workflow and approval
  workflow: {
    requiresApproval: boolean;
    approvedBy?: mongoose.Types.ObjectId;
    approvedAt?: Date;
    rejectedBy?: mongoose.Types.ObjectId;
    rejectedAt?: Date;
    rejectionReason?: string;
  };
  
  createdAt: Date;
  updatedAt: Date;
}

const invoiceSchema = new Schema<IInvoice>({
  invoiceNumber: {
    type: String,
    required: true,
    unique: true,
    index: true
  },
  sequence: {
    type: Number,
    required: true,
    index: true
  },
  
  subscription: {
    type: Schema.Types.ObjectId,
    ref: 'Subscription',
    index: true
  },
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  
  type: {
    type: String,
    enum: ['subscription', 'one_time', 'credit_note', 'proforma', 'recurring'],
    required: true,
    index: true
  },
  description: {
    type: String,
    required: true,
    trim: true,
    maxlength: 1000
  },
  
  lineItems: [{
    id: {
      type: String,
      required: true
    },
    description: {
      type: String,
      required: true,
      trim: true,
      maxlength: 500
    },
    quantity: {
      type: Number,
      required: true,
      min: 0
    },
    unitPrice: {
      type: Number,
      required: true,
      min: 0
    },
    amount: {
      type: Number,
      required: true,
      min: 0
    },
    taxRate: {
      type: Number,
      default: 0,
      min: 0,
      max: 100
    },
    taxAmount: {
      type: Number,
      default: 0,
      min: 0
    },
    periodStart: Date,
    periodEnd: Date,
    planId: {
      type: Schema.Types.ObjectId,
      ref: 'Plan'
    },
    metadata: Schema.Types.Mixed
  }],
  
  subtotal: {
    type: Number,
    required: true,
    min: 0
  },
  taxAmount: {
    type: Number,
    default: 0,
    min: 0
  },
  discountAmount: {
    type: Number,
    default: 0,
    min: 0
  },
  totalAmount: {
    type: Number,
    required: true,
    min: 0
  },
  amountPaid: {
    type: Number,
    default: 0,
    min: 0
  },
  amountDue: {
    type: Number,
    required: true,
    min: 0
  },
  
  currency: {
    type: String,
    required: true,
    uppercase: true,
    length: 3
  },
  exchangeRate: Number,
  baseCurrency: {
    type: String,
    uppercase: true,
    length: 3
  },
  
  taxDetails: {
    taxId: String,
    taxRegion: {
      type: String,
      required: true
    },
    taxType: {
      type: String,
      required: true
    },
    reverseCharge: {
      type: Boolean,
      default: false
    },
    taxExempt: {
      type: Boolean,
      default: false
    },
    taxExemptReason: String
  },
  
  billingAddress: {
    name: {
      type: String,
      required: true,
      trim: true
    },
    company: {
      type: String,
      trim: true
    },
    line1: {
      type: String,
      required: true,
      trim: true
    },
    line2: {
      type: String,
      trim: true
    },
    city: {
      type: String,
      required: true,
      trim: true
    },
    state: {
      type: String,
      trim: true
    },
    postalCode: {
      type: String,
      required: true,
      trim: true
    },
    country: {
      type: String,
      required: true,
      uppercase: true,
      length: 2
    },
    taxId: String
  },
  
  invoiceDate: {
    type: Date,
    required: true,
    index: true
  },
  dueDate: {
    type: Date,
    required: true,
    index: true
  },
  periodStart: Date,
  periodEnd: Date,
  paidAt: Date,
  voidedAt: Date,
  sentAt: Date,
  
  status: {
    type: String,
    enum: ['draft', 'open', 'paid', 'void', 'uncollectible', 'overdue'],
    required: true,
    index: true
  },
  paymentStatus: {
    type: String,
    enum: ['not_paid', 'partially_paid', 'paid', 'refunded', 'failed'],
    required: true,
    index: true
  },
  
  collectionAttempts: [{
    attemptNumber: {
      type: Number,
      required: true
    },
    attemptedAt: {
      type: Date,
      required: true
    },
    method: {
      type: String,
      enum: ['auto', 'manual'],
      required: true
    },
    result: {
      type: String,
      enum: ['succeeded', 'failed', 'pending'],
      required: true
    },
    failureReason: String,
    nextAttempt: Date
  }],
  
  payments: [{
    type: Schema.Types.ObjectId,
    ref: 'Payment'
  }],
  
  provider: {
    type: String,
    enum: ['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'manual'],
    required: true,
    index: true
  },
  providerId: {
    type: String,
    index: true
  },
  providerUrl: String,
  
  appliedCoupons: [{
    couponId: {
      type: Schema.Types.ObjectId,
      ref: 'Coupon',
      required: true
    },
    code: {
      type: String,
      required: true,
      uppercase: true
    },
    discountAmount: {
      type: Number,
      required: true,
      min: 0
    },
    discountType: {
      type: String,
      enum: ['percentage', 'fixed_amount'],
      required: true
    }
  }],
  
  creditNote: {
    originalInvoiceId: {
      type: Schema.Types.ObjectId,
      ref: 'Invoice'
    },
    reason: String,
    type: {
      type: String,
      enum: ['full', 'partial']
    }
  },
  
  files: [{
    type: {
      type: String,
      enum: ['pdf', 'xml', 'csv'],
      required: true
    },
    url: {
      type: String,
      required: true
    },
    size: {
      type: Number,
      required: true,
      min: 0
    },
    generatedAt: {
      type: Date,
      required: true,
      default: Date.now
    }
  }],
  
  communications: [{
    type: {
      type: String,
      enum: ['sent', 'viewed', 'downloaded', 'reminder', 'overdue_notice'],
      required: true
    },
    channel: {
      type: String,
      enum: ['email', 'sms', 'postal', 'portal'],
      required: true
    },
    sentAt: {
      type: Date,
      required: true
    },
    deliveredAt: Date,
    openedAt: Date,
    metadata: Schema.Types.Mixed
  }],
  
  dunning: {
    enabled: {
      type: Boolean,
      default: true
    },
    level: {
      type: Number,
      default: 0,
      min: 0,
      max: 3
    },
    lastNoticeAt: Date,
    nextNoticeAt: Date,
    finalNoticeAt: Date,
    collectionAgencyAt: Date
  },
  
  legalEntity: {
    name: String,
    address: Schema.Types.Mixed,
    taxId: String,
    registrationNumber: String
  },
  
  notes: {
    type: String,
    trim: true,
    maxlength: 2000
  },
  internalNotes: {
    type: String,
    trim: true,
    maxlength: 2000
  },
  tags: [{
    type: String,
    trim: true,
    maxlength: 50
  }],
  
  metadata: {
    type: Schema.Types.Mixed,
    default: {}
  },
  
  workflow: {
    requiresApproval: {
      type: Boolean,
      default: false
    },
    approvedBy: {
      type: Schema.Types.ObjectId,
      ref: 'User'
    },
    approvedAt: Date,
    rejectedBy: {
      type: Schema.Types.ObjectId,
      ref: 'User'
    },
    rejectedAt: Date,
    rejectionReason: String
  }
}, {
  timestamps: true,
  collection: 'invoices'
});

// Compound indexes for efficient queries
invoiceSchema.index({ user: 1, status: 1, invoiceDate: -1 });
invoiceSchema.index({ subscription: 1, status: 1, invoiceDate: -1 });
invoiceSchema.index({ status: 1, dueDate: 1 }); // For overdue tracking
invoiceSchema.index({ provider: 1, providerId: 1 });
invoiceSchema.index({ currency: 1, totalAmount: 1, invoiceDate: -1 });
invoiceSchema.index({ 'billingAddress.country': 1, 'taxDetails.taxType': 1 });
invoiceSchema.index({ type: 1, status: 1, invoiceDate: -1 });
invoiceSchema.index({ sequence: 1, invoiceDate: -1 });

// Text search index
invoiceSchema.index({
  invoiceNumber: 'text',
  description: 'text',
  'billingAddress.name': 'text',
  'billingAddress.company': 'text',
  notes: 'text'
});

// Pre-save middleware for auto-generating invoice number and calculations
invoiceSchema.pre('save', async function(next) {
  if (this.isNew && !this.invoiceNumber) {
    const year = new Date().getFullYear();
    const yearStart = new Date(year, 0, 1);
    const yearEnd = new Date(year + 1, 0, 1);
    
    const sequence = await mongoose.models.Invoice.countDocuments({
      invoiceDate: { $gte: yearStart, $lt: yearEnd }
    }) + 1;
    
    this.sequence = sequence;
    this.invoiceNumber = `INV-${year}-${sequence.toString().padStart(6, '0')}`;
  }
  
  // Calculate amounts
  this.subtotal = this.lineItems.reduce((sum, item) => sum + item.amount, 0);
  this.taxAmount = this.lineItems.reduce((sum, item) => sum + item.taxAmount, 0);
  this.totalAmount = this.subtotal + this.taxAmount - this.discountAmount;
  this.amountDue = this.totalAmount - this.amountPaid;
  
  // Update payment status
  if (this.amountPaid === 0) {
    this.paymentStatus = 'not_paid';
  } else if (this.amountPaid >= this.totalAmount) {
    this.paymentStatus = 'paid';
    if (!this.paidAt) this.paidAt = new Date();
  } else {
    this.paymentStatus = 'partially_paid';
  }
  
  // Update status based on payment and due date
  if (this.paymentStatus === 'paid') {
    this.status = 'paid';
  } else if (this.status !== 'draft' && this.status !== 'void' && this.dueDate < new Date()) {
    this.status = 'overdue';
  }
  
  next();
});

export const Invoice = mongoose.models.Invoice || mongoose.model<IInvoice>('Invoice', invoiceSchema);