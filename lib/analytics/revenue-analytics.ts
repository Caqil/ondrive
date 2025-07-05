import { connectDB } from '@/lib/db';
import { Subscription, Plan } from '@/models';
import { Payment } from '@/models/Payment';
import { AnalyticsDateRange, AnalyticsQuery } from '@/types';

export class RevenueAnalytics {
  /**
   * Get comprehensive revenue metrics
   */
  static async getRevenueMetrics(query: AnalyticsQuery): Promise<RevenueMetrics> {
    await connectDB();

    const { dateRange, compareWith } = query;

    // Total revenue in period
    const revenueData = await Payment.aggregate([
      {
        $match: {
          status: 'succeeded',
          createdAt: { $gte: dateRange.start, $lte: dateRange.end }
        }
      },
      {
        $group: {
          _id: null,
          totalRevenue: { $sum: '$amount' },
          paymentCount: { $sum: 1 },
          averageOrderValue: { $avg: '$amount' }
        }
      }
    ]);

    const totalRevenue = revenueData[0]?.totalRevenue || 0;
    const averageOrderValue = revenueData[0]?.averageOrderValue || 0;

    // Previous period for comparison
    let previousRevenue = 0;
    if (compareWith) {
      const prevRevenueData = await Payment.aggregate([
        {
          $match: {
            status: 'succeeded',
            createdAt: { $gte: compareWith.start, $lte: compareWith.end }
          }
        },
        {
          $group: {
            _id: null,
            totalRevenue: { $sum: '$amount' }
          }
        }
      ]);
      previousRevenue = prevRevenueData[0]?.totalRevenue || 0;
    }

    // Monthly Recurring Revenue (MRR)
    const mrrData = await Subscription.aggregate([
      {
        $match: {
          status: { $in: ['active', 'trial'] },
          currentPeriodEnd: { $gte: dateRange.end }
        }
      },
      {
        $addFields: {
          monthlyAmount: {
            $cond: {
              if: { $eq: ['$interval', 'year'] },
              then: { $divide: ['$amount', 12] },
              else: '$amount'
            }
          }
        }
      },
      {
        $group: {
          _id: null,
          mrr: { $sum: '$monthlyAmount' }
        }
      }
    ]);

    const monthlyRecurringRevenue = mrrData[0]?.mrr || 0;
    const annualRecurringRevenue = monthlyRecurringRevenue * 12;

    // One-time revenue (non-subscription payments)
    const oneTimeRevenueData = await Payment.aggregate([
      {
        $match: {
          status: 'succeeded',
          createdAt: { $gte: dateRange.start, $lte: dateRange.end },
          'metadata.isRenewal': { $ne: true }
        }
      },
      {
        $group: {
          _id: null,
          oneTimeRevenue: { $sum: '$amount' }
        }
      }
    ]);

    const oneTimeRevenue = oneTimeRevenueData[0]?.oneTimeRevenue || 0;

    // Revenue growth rate
    const revenueGrowthRate = previousRevenue > 0 
      ? ((totalRevenue - previousRevenue) / previousRevenue) * 100 
      : 0;

    // Revenue per customer
    const uniqueCustomers = await Payment.distinct('user', {
      status: 'succeeded',
      createdAt: { $gte: dateRange.start, $lte: dateRange.end }
    });
    const revenuePerCustomer = uniqueCustomers.length > 0 ? totalRevenue / uniqueCustomers.length : 0;

    // Conversion rate (trials to paid)
    const conversionData = await this.calculateConversionRate(dateRange);

    // Refund rate
    const refundData = await Payment.aggregate([
      {
        $match: {
          createdAt: { $gte: dateRange.start, $lte: dateRange.end }
        }
      },
      {
        $group: {
          _id: null,
          totalPayments: { $sum: 1 },
          refundedPayments: {
            $sum: {
              $cond: [{ $eq: ['$status', 'refunded'] }, 1, 0]
            }
          },
          refundedAmount: {
            $sum: {
              $cond: [{ $eq: ['$status', 'refunded'] }, '$amount', 0]
            }
          }
        }
      }
    ]);

    const refundRate = refundData[0]?.totalPayments > 0 
      ? (refundData[0].refundedPayments / refundData[0].totalPayments) * 100 
      : 0;

    // Revenue by plan
    const revenueByPlan = await this.getRevenueByPlan(query);

    // Revenue trends
    const revenueTrends = await this.getRevenueTrends(query);

    // Plan distribution
    const planDistribution = await this.getPlanDistribution(query);

    return {
      totalRevenue: {
        value: totalRevenue,
        previousValue: previousRevenue,
        change: totalRevenue - previousRevenue,
        changePercent: revenueGrowthRate
      },
      monthlyRecurringRevenue: {
        value: monthlyRecurringRevenue
      },
      annualRecurringRevenue: {
        value: annualRecurringRevenue
      },
      oneTimeRevenue: {
        value: oneTimeRevenue
      },
      revenueGrowthRate: {
        value: revenueGrowthRate
      },
      averageOrderValue: {
        value: averageOrderValue
      },
      revenuePerCustomer: {
        value: revenuePerCustomer
      },
      conversionRate: {
        value: conversionData.conversionRate
      },
      refundRate: {
        value: refundRate
      },
      revenueByPlan,
      revenueTrends,
      planDistribution
    };
  }

  /**
   * Get revenue trends over time
   */
  static async getRevenueTrends(query: AnalyticsQuery): Promise<TimeSeriesData[]> {
    await connectDB();

    const groupByStage = this.getGroupByStage(query.groupBy || 'day');

    const trends = await Payment.aggregate([
      {
        $match: {
          status: 'succeeded',
          createdAt: { $gte: query.dateRange.start, $lte: query.dateRange.end }
        }
      },
      {
        $group: {
          _id: groupByStage,
          revenue: { $sum: '$amount' },
          paymentCount: { $sum: 1 }
        }
      },
      {
        $sort: { _id: 1 }
      }
    ]);

    return trends.map(trend => ({
      date: new Date(this.dateFromGroupStage(trend._id, query.groupBy || 'day')),
      value: trend.revenue,
      label: `${(trend.revenue / 100).toFixed(2)}`
    }));
  }

  /**
   * Get revenue breakdown by plan
   */
  static async getRevenueByPlan(query: AnalyticsQuery) {
    await connectDB();

    const revenueByPlan = await Payment.aggregate([
      {
        $match: {
          status: 'succeeded',
          createdAt: { $gte: query.dateRange.start, $lte: query.dateRange.end }
        }
      },
      {
        $lookup: {
          from: 'subscriptions',
          localField: 'subscription',
          foreignField: '_id',
          as: 'subscription'
        }
      },
      {
        $unwind: { 
          path: '$subscription',
          preserveNullAndEmptyArrays: true
        }
      },
      {
        $lookup: {
          from: 'plans',
          localField: 'subscription.plan',
          foreignField: '_id',
          as: 'plan'
        }
      },
      {
        $unwind: { 
          path: '$plan',
          preserveNullAndEmptyArrays: true
        }
      },
      {
        $group: {
          _id: {
            planId: '$plan._id',
            planName: { $ifNull: ['$plan.name', 'One-time Payment'] }
          },
          revenue: { $sum: '$amount' },
          customerCount: { $addToSet: '$user' },
          paymentCount: { $sum: 1 }
        }
      },
      {
        $addFields: {
          customerCount: { $size: '$customerCount' }
        }
      },
      {
        $sort: { revenue: -1 }
      }
    ]);

    const totalRevenue = revenueByPlan.reduce((sum, plan) => sum + plan.revenue, 0);

    return revenueByPlan.map(plan => ({
      planName: plan._id.planName,
      revenue: plan.revenue,
      percentage: totalRevenue > 0 ? (plan.revenue / totalRevenue) * 100 : 0,
      customerCount: plan.customerCount
    }));
  }

  /**
   * Get plan distribution (active subscriptions)
   */
  static async getPlanDistribution(query: AnalyticsQuery) {
    await connectDB();

    const distribution = await Subscription.aggregate([
      {
        $match: {
          status: { $in: ['active', 'trial'] },
          currentPeriodEnd: { $gte: query.dateRange.end }
        }
      },
      {
        $lookup: {
          from: 'plans',
          localField: 'plan',
          foreignField: '_id',
          as: 'plan'
        }
      },
      {
        $unwind: '$plan'
      },
      {
        $group: {
          _id: {
            planId: '$plan._id',
            planName: '$plan.name'
          },
          count: { $sum: 1 }
        }
      },
      {
        $sort: { count: -1 }
      }
    ]);

    const totalSubscriptions = distribution.reduce((sum, plan) => sum + plan.count, 0);

    return distribution.map(plan => ({
      plan: plan._id.planName,
      count: plan.count,
      percentage: totalSubscriptions > 0 ? (plan.count / totalSubscriptions) * 100 : 0
    }));
  }

  /**
   * Calculate trial to paid conversion rate
   */
  private static async calculateConversionRate(dateRange: AnalyticsDateRange) {
    // Trials that started in this period
    const trialSubscriptions = await Subscription.countDocuments({
      status: 'trial',
      trialStart: { $gte: dateRange.start, $lte: dateRange.end }
    });

    // Trials that converted to paid in this period
    const convertedSubscriptions = await Subscription.countDocuments({
      status: 'active',
      trialStart: { $gte: dateRange.start, $lte: dateRange.end },
      trialEnd: { $lte: dateRange.end }
    });

    const conversionRate = trialSubscriptions > 0 
      ? (convertedSubscriptions / trialSubscriptions) * 100 
      : 0;

    return {
      trialCount: trialSubscriptions,
      convertedCount: convertedSubscriptions,
      conversionRate
    };
  }

  /**
   * Get revenue cohort analysis
   */
  static async getRevenueCohortAnalysis(query: AnalyticsQuery) {
    await connectDB();

    // Group customers by first payment month
    const cohorts = await Payment.aggregate([
      {
        $match: {
          status: 'succeeded'
        }
      },
      {
        $sort: { user: 1, createdAt: 1 }
      },
      {
        $group: {
          _id: '$user',
          firstPayment: { $first: '$createdAt' },
          payments: { $push: { amount: '$amount', date: '$createdAt' } }
        }
      },
      {
        $group: {
          _id: {
            year: { $year: '$firstPayment' },
            month: { $month: '$firstPayment' }
          },
          customers: { $push: { userId: '$_id', payments: '$payments' } },
          cohortSize: { $sum: 1 }
        }
      }
    ]);

    const cohortAnalysis: Array<{
      cohort: string;
      size: number;
      revenueByMonth: Array<{
        month: number;
        totalRevenue: number;
        activeCustomers: number;
        averageRevenuePerCustomer: number;
      }>;
    }> = [];

    for (const cohort of cohorts) {
      const cohortDate = new Date(cohort._id.year, cohort._id.month - 1, 1);
      const revenueByMonth = await this.calculateCohortRevenue(cohort.customers, cohortDate);
      
      cohortAnalysis.push({
        cohort: `${cohort._id.year}-${cohort._id.month.toString().padStart(2, '0')}`,
        size: cohort.cohortSize,
        revenueByMonth
      });
    }

    return cohortAnalysis;
  }

  /**
   * Calculate cohort revenue over time
   */
  private static async calculateCohortRevenue(customers: any[], cohortStart: Date) {
    const revenueByMonth: Array<{
      month: number;
      totalRevenue: number;
      activeCustomers: number;
      averageRevenuePerCustomer: number;
    }> = [];
    const monthsToTrack = 12;

    for (let month = 0; month < monthsToTrack; month++) {
      const periodStart = new Date(cohortStart.getFullYear(), cohortStart.getMonth() + month, 1);
      const periodEnd = new Date(cohortStart.getFullYear(), cohortStart.getMonth() + month + 1, 0);

      let totalRevenue = 0;
      let activeCustomers = 0;

      customers.forEach(customer => {
        const monthlyRevenue = customer.payments
          .filter((payment: any) => {
            const paymentDate = new Date(payment.date);
            return paymentDate >= periodStart && paymentDate <= periodEnd;
          })
          .reduce((sum: number, payment: any) => sum + payment.amount, 0);

        if (monthlyRevenue > 0) {
          totalRevenue += monthlyRevenue;
          activeCustomers++;
        }
      });

      revenueByMonth.push({
        month,
        totalRevenue,
        activeCustomers,
        averageRevenuePerCustomer: activeCustomers > 0 ? totalRevenue / activeCustomers : 0
      });
    }

    return revenueByMonth;
  }

  /**
   * Helper method to get MongoDB group stage for different time periods
   */
  private static getGroupByStage(groupBy: string) {
    switch (groupBy) {
      case 'hour':
        return {
          year: { $year: '$createdAt' },
          month: { $month: '$createdAt' },
          day: { $dayOfMonth: '$createdAt' },
          hour: { $hour: '$createdAt' }
        };
      case 'day':
        return {
          year: { $year: '$createdAt' },
          month: { $month: '$createdAt' },
          day: { $dayOfMonth: '$createdAt' }
        };
      case 'week':
        return {
          year: { $year: '$createdAt' },
          week: { $week: '$createdAt' }
        };
      case 'month':
        return {
          year: { $year: '$createdAt' },
          month: { $month: '$createdAt' }
        };
      default:
        return {
          year: { $year: '$createdAt' },
          month: { $month: '$createdAt' },
          day: { $dayOfMonth: '$createdAt' }
        };
    }
  }

  /**
   * Helper to convert group stage back to date
   */
  private static dateFromGroupStage(groupStage: any, groupBy: string): string {
    switch (groupBy) {
      case 'hour':
        return new Date(groupStage.year, groupStage.month - 1, groupStage.day, groupStage.hour).toISOString();
      case 'day':
        return new Date(groupStage.year, groupStage.month - 1, groupStage.day).toISOString();
      case 'week':
        const firstDay = new Date(groupStage.year, 0, 1);
        const weekStart = new Date(firstDay.getTime() + (groupStage.week - 1) * 7 * 24 * 60 * 60 * 1000);
        return weekStart.toISOString();
      case 'month':
        return new Date(groupStage.year, groupStage.month - 1, 1).toISOString();
      default:
        return new Date(groupStage.year, groupStage.month - 1, groupStage.day).toISOString();
    }
  }
}