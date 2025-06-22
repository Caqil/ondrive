package controllers

// Controllers struct holds all controller instances
type Controllers struct {
	AuthController         *AuthController
	UserController         *UserController
	RideController         *RideController
	DriverController       *DriverController
	PaymentController      *PaymentController
	FareController         *FareController
	ChatController         *ChatController
	RatingController       *RatingController
	CourierController      *CourierController
	FreightController      *FreightController
	NotificationController *NotificationController
	WebSocketController    *WebSocketController
	PublicController       *PublicController
	UploadController       *UploadController
}

// NewControllers creates and returns all controller instances
func NewControllers(services *services.Services) *Controllers {
	return &Controllers{
		AuthController:         NewAuthController(services.AuthService),
		UserController:         NewUserController(services.UserService),
		RideController:         NewRideController(services.RideService),
		DriverController:       NewDriverController(services.DriverService),
		PaymentController:      NewPaymentController(services.PaymentService),
		FareController:         NewFareController(services.FareService),
		ChatController:         NewChatController(services.ChatService),
		RatingController:       NewRatingController(services.RatingService),
		CourierController:      NewCourierController(services.CourierService),
		FreightController:      NewFreightController(services.FreightService),
		NotificationController: NewNotificationController(services.NotificationService),
		WebSocketController:    NewWebSocketController(services.WebSocketService),
		PublicController:       NewPublicController(services.PublicService),
		UploadController:       NewUploadController(services.UploadService),
	}
}
