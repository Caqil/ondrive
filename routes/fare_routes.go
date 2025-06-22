package routes

import (
	"indrive-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupFareRoutes(rg *gin.RouterGroup, controller *controllers.FareController) {
	fares := rg.Group("/fares")
	{
		// Fare Calculation
		fares.POST("/estimate", controller.EstimateFare)
		fares.POST("/calculate", controller.CalculateFare)
		fares.GET("/base-rates", controller.GetBaseRates)
		fares.GET("/surge-info", controller.GetSurgeInfo)

		// Fare Negotiation (InDrive's core feature)
		fares.POST("/propose", controller.ProposeFare)
		fares.POST("/counter-offer", controller.CounterOffer)
		fares.POST("/accept", controller.AcceptFare)
		fares.POST("/reject", controller.RejectFare)
		fares.GET("/negotiations/:ride_id", controller.GetNegotiationHistory)

		// Fare Comparison
		fares.GET("/compare", controller.CompareFares)
		fares.GET("/market-rates", controller.GetMarketRates)
		fares.GET("/suggested-fare", controller.GetSuggestedFare)

		// Fare Rules & Settings
		fares.GET("/rules", controller.GetFareRules)
		fares.GET("/minimum-fare", controller.GetMinimumFare)
		fares.GET("/maximum-fare", controller.GetMaximumFare)

		// Fare History & Analytics
		fares.GET("/history", controller.GetFareHistory)
		fares.GET("/statistics", controller.GetFareStatistics)
		fares.GET("/trends", controller.GetFareTrends)

		// Special Pricing
		fares.GET("/promotional", controller.GetPromotionalPricing)
		fares.GET("/discount-codes", controller.GetDiscountCodes)
		fares.POST("/apply-discount", controller.ApplyDiscountCode)

		// Commission & Fees
		fares.GET("/commission-rates", controller.GetCommissionRates)
		fares.GET("/service-fees", controller.GetServiceFees)
		fares.GET("/breakdown/:ride_id", controller.GetFareBreakdown)
	}
}
