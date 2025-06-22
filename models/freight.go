package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FreightStatus string

const (
	FreightStatusPending   FreightStatus = "pending"
	FreightStatusQuoted    FreightStatus = "quoted"
	FreightStatusAccepted  FreightStatus = "accepted"
	FreightStatusScheduled FreightStatus = "scheduled"
	FreightStatusLoading   FreightStatus = "loading"
	FreightStatusInTransit FreightStatus = "in_transit"
	FreightStatusUnloading FreightStatus = "unloading"
	FreightStatusDelivered FreightStatus = "delivered"
	FreightStatusCancelled FreightStatus = "cancelled"
	FreightStatusReturned  FreightStatus = "returned"
)

type CargoType string

const (
	CargoTypeGeneral       CargoType = "general"
	CargoTypeConstruction  CargoType = "construction"
	CargoTypeMachinery     CargoType = "machinery"
	CargoTypeFurniture     CargoType = "furniture"
	CargoTypeVehicle       CargoType = "vehicle"
	CargoTypeFood          CargoType = "food"
	CargoTypeChemical      CargoType = "chemical"
	CargoTypeTextile       CargoType = "textile"
	CargoTypeElectronics   CargoType = "electronics"
	CargoTypeOversized     CargoType = "oversized"
	CargoTypeHazardous     CargoType = "hazardous"
	CargoTypeLivestock     CargoType = "livestock"
	CargoTypePetrochemical CargoType = "petrochemical"
)

type FreightVehicleType string

const (
	FreightVehicleSmallTruck   FreightVehicleType = "small_truck"  // 1-3 tons
	FreightVehicleMediumTruck  FreightVehicleType = "medium_truck" // 3-7 tons
	FreightVehicleLargeTruck   FreightVehicleType = "large_truck"  // 7-15 tons
	FreightVehicleTrailer      FreightVehicleType = "trailer"      // 15+ tons
	FreightVehicleContainer    FreightVehicleType = "container"    // 20/40 ft container
	FreightVehicleFlatbed      FreightVehicleType = "flatbed"      // Open flatbed
	FreightVehicleTanker       FreightVehicleType = "tanker"       // Liquid cargo
	FreightVehicleRefrigerated FreightVehicleType = "refrigerated" // Cold chain
)

type LoadingType string

const (
	LoadingTypeManual    LoadingType = "manual"
	LoadingTypeForkLift  LoadingType = "forklift"
	LoadingTypeCrane     LoadingType = "crane"
	LoadingTypeConveyor  LoadingType = "conveyor"
	LoadingTypeHydraulic LoadingType = "hydraulic"
	LoadingTypeTailgate  LoadingType = "tailgate"
)

type FreightRequest struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Request Information
	Status   FreightStatus `json:"status" bson:"status"`
	Type     string        `json:"type" bson:"type"`         // point_to_point, multi_stop, round_trip
	Priority string        `json:"priority" bson:"priority"` // standard, urgent, scheduled

	// Participants
	ShipperID primitive.ObjectID  `json:"shipper_id" bson:"shipper_id"`
	CarrierID *primitive.ObjectID `json:"carrier_id,omitempty" bson:"carrier_id,omitempty"`

	// Cargo Information
	Cargo FreightCargo `json:"cargo" bson:"cargo"`

	// Vehicle Requirements
	VehicleRequirements FreightVehicleRequirements `json:"vehicle_requirements" bson:"vehicle_requirements"`

	// Route Information
	PickupLocation   FreightLocation `json:"pickup_location" bson:"pickup_location"`
	DeliveryLocation FreightLocation `json:"delivery_location" bson:"delivery_location"`
	RouteStops       []FreightStop   `json:"route_stops" bson:"route_stops"`

	// Scheduling
	ScheduledPickupTime   *time.Time   `json:"scheduled_pickup_time,omitempty" bson:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time   `json:"scheduled_delivery_time,omitempty" bson:"scheduled_delivery_time,omitempty"`
	FlexibleScheduling    bool         `json:"flexible_scheduling" bson:"flexible_scheduling"`
	TimeWindows           []TimeWindow `json:"time_windows" bson:"time_windows"`

	// Actual Timing
	LoadingStartedAt   *time.Time `json:"loading_started_at,omitempty" bson:"loading_started_at,omitempty"`
	LoadingCompletedAt *time.Time `json:"loading_completed_at,omitempty" bson:"loading_completed_at,omitempty"`
	DepartedAt         *time.Time `json:"departed_at,omitempty" bson:"departed_at,omitempty"`
	ArrivedAt          *time.Time `json:"arrived_at,omitempty" bson:"arrived_at,omitempty"`
	UnloadingStartedAt *time.Time `json:"unloading_started_at,omitempty" bson:"unloading_started_at,omitempty"`
	DeliveredAt        *time.Time `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`

	// Pricing
	Fare            FreightFare `json:"fare" bson:"fare"`
	PaymentTerms    string      `json:"payment_terms" bson:"payment_terms"`
	PaymentMethodID string      `json:"payment_method_id" bson:"payment_method_id"`
	PaymentStatus   string      `json:"payment_status" bson:"payment_status"`

	// Loading & Unloading
	LoadingRequirements   LoadingRequirements `json:"loading_requirements" bson:"loading_requirements"`
	UnloadingRequirements LoadingRequirements `json:"unloading_requirements" bson:"unloading_requirements"`

	// Documentation
	Documents   []FreightDocument `json:"documents" bson:"documents"`
	CustomsInfo *CustomsInfo      `json:"customs_info,omitempty" bson:"customs_info,omitempty"`

	// Tracking
	TrackingCode     string          `json:"tracking_code" bson:"tracking_code"`
	TrackingHistory  []TrackingEvent `json:"tracking_history" bson:"tracking_history"`
	CurrentLocation  *Location       `json:"current_location,omitempty" bson:"current_location,omitempty"`
	Route            []Location      `json:"route" bson:"route"`
	EstimatedArrival *time.Time      `json:"estimated_arrival,omitempty" bson:"estimated_arrival,omitempty"`

	// Distance & Duration
	EstimatedDistance float64 `json:"estimated_distance" bson:"estimated_distance"` // km
	ActualDistance    float64 `json:"actual_distance" bson:"actual_distance"`       // km
	EstimatedDuration int     `json:"estimated_duration" bson:"estimated_duration"` // hours
	ActualDuration    int     `json:"actual_duration" bson:"actual_duration"`       // hours

	// Communication
	SpecialInstructions string              `json:"special_instructions" bson:"special_instructions"`
	ConversationID      *primitive.ObjectID `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`

	// Insurance & Safety
	Insurance          *FreightInsurance `json:"insurance,omitempty" bson:"insurance,omitempty"`
	SafetyRequirements []string          `json:"safety_requirements" bson:"safety_requirements"`

	// Delivery Confirmation
	DeliveryProof *FreightDeliveryProof `json:"delivery_proof,omitempty" bson:"delivery_proof,omitempty"`

	// Cancellation
	CancellationReason string              `json:"cancellation_reason" bson:"cancellation_reason"`
	CancelledBy        *primitive.ObjectID `json:"cancelled_by,omitempty" bson:"cancelled_by,omitempty"`
	CancelledAt        *time.Time          `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`

	// Rating & Feedback
	ShipperRating *Rating `json:"shipper_rating,omitempty" bson:"shipper_rating,omitempty"`
	CarrierRating *Rating `json:"carrier_rating,omitempty" bson:"carrier_rating,omitempty"`

	// Metadata
	Platform  string    `json:"platform" bson:"platform"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type FreightCargo struct {
	// Basic Information
	Description string    `json:"description" bson:"description" validate:"required"`
	Type        CargoType `json:"type" bson:"type"`
	Category    string    `json:"category" bson:"category"`
	Quantity    int       `json:"quantity" bson:"quantity" validate:"gte=1"`

	// Physical Properties
	TotalWeight float64         `json:"total_weight" bson:"total_weight" validate:"gte=20,lte=50000"` // kg (min 20kg for freight)
	TotalVolume float64         `json:"total_volume" bson:"total_volume"`                             // m³
	Dimensions  CargoDimensions `json:"dimensions" bson:"dimensions"`
	Items       []CargoItem     `json:"items" bson:"items"`

	// Value & Insurance
	TotalValue float64 `json:"total_value" bson:"total_value"`
	Currency   string  `json:"currency" bson:"currency"`

	// Special Properties
	IsHazardous           bool              `json:"is_hazardous" bson:"is_hazardous"`
	IsFragile             bool              `json:"is_fragile" bson:"is_fragile"`
	IsPerishable          bool              `json:"is_perishable" bson:"is_perishable"`
	IsOversized           bool              `json:"is_oversized" bson:"is_oversized"`
	IsHighValue           bool              `json:"is_high_value" bson:"is_high_value"`
	RequiresRefrigeration bool              `json:"requires_refrigeration" bson:"requires_refrigeration"`
	TemperatureRange      *TemperatureRange `json:"temperature_range,omitempty" bson:"temperature_range,omitempty"`

	// Handling Requirements
	HandlingInstructions []string    `json:"handling_instructions" bson:"handling_instructions"`
	SpecialEquipment     []string    `json:"special_equipment" bson:"special_equipment"`
	LoadingMethod        LoadingType `json:"loading_method" bson:"loading_method"`
	UnloadingMethod      LoadingType `json:"unloading_method" bson:"unloading_method"`

	// Packaging
	PackagingType      string `json:"packaging_type" bson:"packaging_type"`
	PackagingDetails   string `json:"packaging_details" bson:"packaging_details"`
	NumberOfPallets    int    `json:"number_of_pallets" bson:"number_of_pallets"`
	NumberOfBoxes      int    `json:"number_of_boxes" bson:"number_of_boxes"`
	NumberOfContainers int    `json:"number_of_containers" bson:"number_of_containers"`

	// Documentation
	Photos             []CargoPhoto `json:"photos" bson:"photos"`
	ManifestNumber     string       `json:"manifest_number" bson:"manifest_number"`
	CustomsDeclaration string       `json:"customs_declaration" bson:"customs_declaration"`

	// Compliance & Safety
	HazmatClass     string   `json:"hazmat_class" bson:"hazmat_class"`
	UNNumber        string   `json:"un_number" bson:"un_number"`
	SafetyDataSheet string   `json:"safety_data_sheet" bson:"safety_data_sheet"`
	Certifications  []string `json:"certifications" bson:"certifications"`
}

type CargoDimensions struct {
	Length float64 `json:"length" bson:"length"` // meters
	Width  float64 `json:"width" bson:"width"`   // meters
	Height float64 `json:"height" bson:"height"` // meters
	Volume float64 `json:"volume" bson:"volume"` // calculated volume in m³
}

type CargoItem struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Quantity      int                `json:"quantity" bson:"quantity"`
	Weight        float64            `json:"weight" bson:"weight"`
	Dimensions    CargoDimensions    `json:"dimensions" bson:"dimensions"`
	Value         float64            `json:"value" bson:"value"`
	HSCode        string             `json:"hs_code" bson:"hs_code"`
	SKU           string             `json:"sku" bson:"sku"`
	SerialNumbers []string           `json:"serial_numbers" bson:"serial_numbers"`

	// Special Properties
	IsFragile    bool `json:"is_fragile" bson:"is_fragile"`
	IsHazardous  bool `json:"is_hazardous" bson:"is_hazardous"`
	RequiresCare bool `json:"requires_care" bson:"requires_care"`

	// Packaging
	PackagingType string `json:"packaging_type" bson:"packaging_type"`
	ContainerID   string `json:"container_id" bson:"container_id"`
	PalletID      string `json:"pallet_id" bson:"pallet_id"`
}

type CargoPhoto struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	URL         string             `json:"url" bson:"url"`
	Type        string             `json:"type" bson:"type"` // loading, loaded, unloading, delivered, damage
	Description string             `json:"description" bson:"description"`
	TakenAt     time.Time          `json:"taken_at" bson:"taken_at"`
	TakenBy     primitive.ObjectID `json:"taken_by" bson:"taken_by"`
	Location    *Location          `json:"location,omitempty" bson:"location,omitempty"`
}

type TemperatureRange struct {
	MinTemperature float64 `json:"min_temperature" bson:"min_temperature"` // Celsius
	MaxTemperature float64 `json:"max_temperature" bson:"max_temperature"` // Celsius
	Unit           string  `json:"unit" bson:"unit"`                       // celsius, fahrenheit
}

type FreightVehicleRequirements struct {
	VehicleType       FreightVehicleType `json:"vehicle_type" bson:"vehicle_type"`
	MinLoadCapacity   float64            `json:"min_load_capacity" bson:"min_load_capacity"`     // tons
	MinVolumeCapacity float64            `json:"min_volume_capacity" bson:"min_volume_capacity"` // m³

	// Dimensions
	MinCargoLength float64 `json:"min_cargo_length" bson:"min_cargo_length"` // meters
	MinCargoWidth  float64 `json:"min_cargo_width" bson:"min_cargo_width"`   // meters
	MinCargoHeight float64 `json:"min_cargo_height" bson:"min_cargo_height"` // meters

	// Special Features
	RequiresRefrigeration bool `json:"requires_refrigeration" bson:"requires_refrigeration"`
	RequiresTailgate      bool `json:"requires_tailgate" bson:"requires_tailgate"`
	RequiresCrane         bool `json:"requires_crane" bson:"requires_crane"`
	RequiresSideLoader    bool `json:"requires_side_loader" bson:"requires_side_loader"`
	RequiresSecureStorage bool `json:"requires_secure_storage" bson:"requires_secure_storage"`
	RequiresGPS           bool `json:"requires_gps" bson:"requires_gps"`

	// Certifications
	RequiredCertifications []string `json:"required_certifications" bson:"required_certifications"`
	RequiredLicenses       []string `json:"required_licenses" bson:"required_licenses"`

	// Age & Condition
	MaxVehicleAge      int     `json:"max_vehicle_age" bson:"max_vehicle_age"` // years
	MinConditionRating float64 `json:"min_condition_rating" bson:"min_condition_rating"`

	// Driver Requirements
	MinDriverExperience   int      `json:"min_driver_experience" bson:"min_driver_experience"` // years
	RequiredDriverLicense []string `json:"required_driver_license" bson:"required_driver_license"`
	DriverLanguages       []string `json:"driver_languages" bson:"driver_languages"`
}

type FreightLocation struct {
	// Geographic Information
	Coordinates []float64 `json:"coordinates" bson:"coordinates" validate:"coordinates"`
	Address     string    `json:"address" bson:"address" validate:"required"`
	Name        string    `json:"name" bson:"name"`
	PlaceID     string    `json:"place_id" bson:"place_id"`

	// Detailed Address
	CompanyName     string `json:"company_name" bson:"company_name"`
	BuildingNumber  string `json:"building_number" bson:"building_number"`
	WarehouseNumber string `json:"warehouse_number" bson:"warehouse_number"`
	DockNumber      string `json:"dock_number" bson:"dock_number"`
	Gate            string `json:"gate" bson:"gate"`
	Yard            string `json:"yard" bson:"yard"`

	// Contact Information
	ContactPerson    string `json:"contact_person" bson:"contact_person"`
	ContactPhone     string `json:"contact_phone" bson:"contact_phone" validate:"phone"`
	ContactEmail     string `json:"contact_email" bson:"contact_email" validate:"email"`
	AlternateContact string `json:"alternate_contact" bson:"alternate_contact"`

	// Operating Hours
	OperatingHours   string `json:"operating_hours" bson:"operating_hours"`
	TimeZone         string `json:"time_zone" bson:"time_zone"`
	WeekendAvailable bool   `json:"weekend_available" bson:"weekend_available"`
	HolidayAvailable bool   `json:"holiday_available" bson:"holiday_available"`

	// Facility Information
	FacilityType       string   `json:"facility_type" bson:"facility_type"`       // warehouse, factory, port, construction_site
	LoadingCapacity    int      `json:"loading_capacity" bson:"loading_capacity"` // number of trucks
	AvailableEquipment []string `json:"available_equipment" bson:"available_equipment"`
	HasScale           bool     `json:"has_scale" bson:"has_scale"`
	HasForkLift        bool     `json:"has_forklift" bson:"has_forklift"`
	HasCrane           bool     `json:"has_crane" bson:"has_crane"`
	HasDock            bool     `json:"has_dock" bson:"has_dock"`

	// Access Information
	AccessRestrictions  string `json:"access_restrictions" bson:"access_restrictions"`
	RequiresAppointment bool   `json:"requires_appointment" bson:"requires_appointment"`
	SecurityClearance   bool   `json:"security_clearance" bson:"security_clearance"`
	TruckRouteAccess    bool   `json:"truck_route_access" bson:"truck_route_access"`

	// Parking & Maneuvering
	ParkingAvailable   bool    `json:"parking_available" bson:"parking_available"`
	ManeuveringSpace   string  `json:"maneuvering_space" bson:"maneuvering_space"`     // tight, moderate, ample
	MaxVehicleLength   float64 `json:"max_vehicle_length" bson:"max_vehicle_length"`   // meters
	MaxVehicleHeight   float64 `json:"max_vehicle_height" bson:"max_vehicle_height"`   // meters
	WeightRestrictions float64 `json:"weight_restrictions" bson:"weight_restrictions"` // tons

	// Special Instructions
	LoadingInstructions    string `json:"loading_instructions" bson:"loading_instructions"`
	UnloadingInstructions  string `json:"unloading_instructions" bson:"unloading_instructions"`
	SafetyRequirements     string `json:"safety_requirements" bson:"safety_requirements"`
	SpecialEquipmentNeeded string `json:"special_equipment_needed" bson:"special_equipment_needed"`

	// Costs
	LoadingFee     float64 `json:"loading_fee" bson:"loading_fee"`
	UnloadingFee   float64 `json:"unloading_fee" bson:"unloading_fee"`
	StorageFee     float64 `json:"storage_fee" bson:"storage_fee"`
	DemurrageFee   float64 `json:"demurrage_fee" bson:"demurrage_fee"`
	WaitingTimeFee float64 `json:"waiting_time_fee" bson:"waiting_time_fee"`
}

type FreightStop struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Location         FreightLocation    `json:"location" bson:"location"`
	StopType         string             `json:"stop_type" bson:"stop_type"` // pickup, delivery, fuel, rest, inspection
	SequenceNumber   int                `json:"sequence_number" bson:"sequence_number"`
	EstimatedArrival *time.Time         `json:"estimated_arrival,omitempty" bson:"estimated_arrival,omitempty"`
	ActualArrival    *time.Time         `json:"actual_arrival,omitempty" bson:"actual_arrival,omitempty"`
	EstimatedDepart  *time.Time         `json:"estimated_depart,omitempty" bson:"estimated_depart,omitempty"`
	ActualDepart     *time.Time         `json:"actual_depart,omitempty" bson:"actual_depart,omitempty"`
	Duration         int                `json:"duration" bson:"duration"` // minutes
	IsCompleted      bool               `json:"is_completed" bson:"is_completed"`
	Notes            string             `json:"notes" bson:"notes"`
	CargoAction      string             `json:"cargo_action" bson:"cargo_action"` // load, unload, inspect, none
	CargoItems       []string           `json:"cargo_items" bson:"cargo_items"`
}

type LoadingRequirements struct {
	LoadingMethod       LoadingType `json:"loading_method" bson:"loading_method"`
	RequiredEquipment   []string    `json:"required_equipment" bson:"required_equipment"`
	RequiredPersonnel   int         `json:"required_personnel" bson:"required_personnel"`
	EstimatedTime       int         `json:"estimated_time" bson:"estimated_time"` // minutes
	SpecialInstructions string      `json:"special_instructions" bson:"special_instructions"`
	SafetyRequirements  []string    `json:"safety_requirements" bson:"safety_requirements"`
	RequiresSupervision bool        `json:"requires_supervision" bson:"requires_supervision"`
	WeightDistribution  string      `json:"weight_distribution" bson:"weight_distribution"`
	SecuringMethod      string      `json:"securing_method" bson:"securing_method"`
	LoadingSequence     []string    `json:"loading_sequence" bson:"loading_sequence"`
}

type FreightFare struct {
	// Base Pricing
	BaseFare     float64 `json:"base_fare" bson:"base_fare"`
	WeightRate   float64 `json:"weight_rate" bson:"weight_rate"`     // per kg
	DistanceRate float64 `json:"distance_rate" bson:"distance_rate"` // per km
	VolumeRate   float64 `json:"volume_rate" bson:"volume_rate"`     // per m³
	TimeRate     float64 `json:"time_rate" bson:"time_rate"`         // per hour

	// Special Service Fees
	LoadingFee         float64 `json:"loading_fee" bson:"loading_fee"`
	UnloadingFee       float64 `json:"unloading_fee" bson:"unloading_fee"`
	HazmatFee          float64 `json:"hazmat_fee" bson:"hazmat_fee"`
	OversizeFee        float64 `json:"oversize_fee" bson:"oversize_fee"`
	FragileHandlingFee float64 `json:"fragile_handling_fee" bson:"fragile_handling_fee"`
	RefrigerationFee   float64 `json:"refrigeration_fee" bson:"refrigeration_fee"`
	ExpressFee         float64 `json:"express_fee" bson:"express_fee"`

	// Additional Charges
	FuelSurcharge  float64 `json:"fuel_surcharge" bson:"fuel_surcharge"`
	TollCharges    float64 `json:"toll_charges" bson:"toll_charges"`
	ParkingFees    float64 `json:"parking_fees" bson:"parking_fees"`
	WaitingTimeFee float64 `json:"waiting_time_fee" bson:"waiting_time_fee"`
	DemurrageFee   float64 `json:"demurrage_fee" bson:"demurrage_fee"`
	DetentionFee   float64 `json:"detention_fee" bson:"detention_fee"`

	// Insurance & Protection
	InsuranceFee float64 `json:"insurance_fee" bson:"insurance_fee"`
	SecurityFee  float64 `json:"security_fee" bson:"security_fee"`

	// Route-specific
	MultiplStopFee   float64 `json:"multiple_stop_fee" bson:"multiple_stop_fee"`
	BackhaulDiscount float64 `json:"backhaul_discount" bson:"backhaul_discount"`

	// Timing Adjustments
	PeakHourSurcharge float64 `json:"peak_hour_surcharge" bson:"peak_hour_surcharge"`
	WeekendSurcharge  float64 `json:"weekend_surcharge" bson:"weekend_surcharge"`
	HolidaySurcharge  float64 `json:"holiday_surcharge" bson:"holiday_surcharge"`
	UrgentServiceFee  float64 `json:"urgent_service_fee" bson:"urgent_service_fee"`

	// Discounts
	VolumeDiscount          float64 `json:"volume_discount" bson:"volume_discount"`
	RegularCustomerDiscount float64 `json:"regular_customer_discount" bson:"regular_customer_discount"`
	SeasonalDiscount        float64 `json:"seasonal_discount" bson:"seasonal_discount"`
	PromoDiscount           float64 `json:"promo_discount" bson:"promo_discount"`
	PromoCode               string  `json:"promo_code" bson:"promo_code"`

	// Totals & Commission
	SubTotal        float64 `json:"sub_total" bson:"sub_total"`
	TaxAmount       float64 `json:"tax_amount" bson:"tax_amount"`
	TotalAmount     float64 `json:"total_amount" bson:"total_amount"`
	CarrierEarnings float64 `json:"carrier_earnings" bson:"carrier_earnings"`
	Commission      float64 `json:"commission" bson:"commission"`
	CommissionRate  float64 `json:"commission_rate" bson:"commission_rate"`

	Currency string `json:"currency" bson:"currency"`

	// Payment Terms
	PaymentTerms   string  `json:"payment_terms" bson:"payment_terms"` // immediate, net_15, net_30, cod
	AdvancePayment float64 `json:"advance_payment" bson:"advance_payment"`
	BalancePayment float64 `json:"balance_payment" bson:"balance_payment"`
}

type FreightDocument struct {
	ID             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Type           string              `json:"type" bson:"type"` // bill_of_lading, manifest, invoice, customs, insurance
	Name           string              `json:"name" bson:"name"`
	DocumentNumber string              `json:"document_number" bson:"document_number"`
	URL            string              `json:"url" bson:"url"`
	IsRequired     bool                `json:"is_required" bson:"is_required"`
	IsUploaded     bool                `json:"is_uploaded" bson:"is_uploaded"`
	UploadedAt     *time.Time          `json:"uploaded_at,omitempty" bson:"uploaded_at,omitempty"`
	UploadedBy     *primitive.ObjectID `json:"uploaded_by,omitempty" bson:"uploaded_by,omitempty"`
	IsVerified     bool                `json:"is_verified" bson:"is_verified"`
	VerifiedAt     *time.Time          `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	VerifiedBy     *primitive.ObjectID `json:"verified_by,omitempty" bson:"verified_by,omitempty"`
	ExpiryDate     *time.Time          `json:"expiry_date,omitempty" bson:"expiry_date,omitempty"`
	Notes          string              `json:"notes" bson:"notes"`
}

type CustomsInfo struct {
	CountryOfOrigin      string         `json:"country_of_origin" bson:"country_of_origin"`
	CountryOfDestination string         `json:"country_of_destination" bson:"country_of_destination"`
	HSCode               string         `json:"hs_code" bson:"hs_code"`
	CustomsValue         float64        `json:"customs_value" bson:"customs_value"`
	Currency             string         `json:"currency" bson:"currency"`
	DutyRate             float64        `json:"duty_rate" bson:"duty_rate"`
	DutyAmount           float64        `json:"duty_amount" bson:"duty_amount"`
	TaxAmount            float64        `json:"tax_amount" bson:"tax_amount"`
	DeclarationNumber    string         `json:"declaration_number" bson:"declaration_number"`
	BrokerInfo           *CustomsBroker `json:"broker_info,omitempty" bson:"broker_info,omitempty"`
	RequiredDocuments    []string       `json:"required_documents" bson:"required_documents"`
	SpecialPermits       []string       `json:"special_permits" bson:"special_permits"`
	RestrictedItems      []string       `json:"restricted_items" bson:"restricted_items"`
	IncoTerms            string         `json:"inco_terms" bson:"inco_terms"`
	ExportLicense        string         `json:"export_license" bson:"export_license"`
	ImportLicense        string         `json:"import_license" bson:"import_license"`
}

type CustomsBroker struct {
	Name          string `json:"name" bson:"name"`
	LicenseNumber string `json:"license_number" bson:"license_number"`
	Phone         string `json:"phone" bson:"phone"`
	Email         string `json:"email" bson:"email"`
	Address       string `json:"address" bson:"address"`
}

type FreightInsurance struct {
	IsRequired     bool      `json:"is_required" bson:"is_required"`
	CoverageType   string    `json:"coverage_type" bson:"coverage_type"` // cargo, liability, comprehensive
	CoverageAmount float64   `json:"coverage_amount" bson:"coverage_amount"`
	Premium        float64   `json:"premium" bson:"premium"`
	Currency       string    `json:"currency" bson:"currency"`
	PolicyNumber   string    `json:"policy_number" bson:"policy_number"`
	ProviderName   string    `json:"provider_name" bson:"provider_name"`
	Deductible     float64   `json:"deductible" bson:"deductible"`
	ValidFrom      time.Time `json:"valid_from" bson:"valid_from"`
	ValidTo        time.Time `json:"valid_to" bson:"valid_to"`
	SupportContact string    `json:"support_contact" bson:"support_contact"`
	ClaimProcess   string    `json:"claim_process" bson:"claim_process"`
	Exclusions     []string  `json:"exclusions" bson:"exclusions"`
	Terms          string    `json:"terms" bson:"terms"`
}

type FreightDeliveryProof struct {
	// Delivery Information
	DeliveredTo      string             `json:"delivered_to" bson:"delivered_to"`
	DeliveredBy      primitive.ObjectID `json:"delivered_by" bson:"delivered_by"`
	DeliveryTime     time.Time          `json:"delivery_time" bson:"delivery_time"`
	DeliveryLocation FreightLocation    `json:"delivery_location" bson:"delivery_location"`

	// Proof Documentation
	SignatureURL    string   `json:"signature_url" bson:"signature_url"`
	DeliveryPhotos  []string `json:"delivery_photos" bson:"delivery_photos"`
	WeighBridgeSlip string   `json:"weigh_bridge_slip" bson:"weigh_bridge_slip"`
	UnloadingPhotos []string `json:"unloading_photos" bson:"unloading_photos"`

	// Condition Assessment
	CargoCondition    string        `json:"cargo_condition" bson:"cargo_condition"` // good, damaged, partial, complete_loss
	DamageReport      *DamageReport `json:"damage_report,omitempty" bson:"damage_report,omitempty"`
	QuantityDelivered float64       `json:"quantity_delivered" bson:"quantity_delivered"`
	WeightDelivered   float64       `json:"weight_delivered" bson:"weight_delivered"`

	// Bill of Lading
	BOLNumber    string `json:"bol_number" bson:"bol_number"`
	BOLSignedURL string `json:"bol_signed_url" bson:"bol_signed_url"`

	// Recipient Information
	RecipientName  string `json:"recipient_name" bson:"recipient_name"`
	RecipientID    string `json:"recipient_id" bson:"recipient_id"`
	RecipientTitle string `json:"recipient_title" bson:"recipient_title"`

	// Additional Documentation
	CustomsClearance string `json:"customs_clearance" bson:"customs_clearance"`
	InspectionReport string `json:"inspection_report" bson:"inspection_report"`
	QualityReport    string `json:"quality_report" bson:"quality_report"`

	// Feedback & Notes
	RecipientFeedback string `json:"recipient_feedback" bson:"recipient_feedback"`
	DeliveryNotes     string `json:"delivery_notes" bson:"delivery_notes"`
	SpecialRemarks    string `json:"special_remarks" bson:"special_remarks"`

	// Verification
	IsVerified bool                `json:"is_verified" bson:"is_verified"`
	VerifiedAt *time.Time          `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	VerifiedBy *primitive.ObjectID `json:"verified_by,omitempty" bson:"verified_by,omitempty"`
}

type DamageReport struct {
	ReportNumber     string             `json:"report_number" bson:"report_number"`
	DamageType       string             `json:"damage_type" bson:"damage_type"`     // physical, water, theft, contamination
	DamageExtent     string             `json:"damage_extent" bson:"damage_extent"` // minor, major, total
	DamagePhotos     []string           `json:"damage_photos" bson:"damage_photos"`
	Description      string             `json:"description" bson:"description"`
	CauseOfDamage    string             `json:"cause_of_damage" bson:"cause_of_damage"`
	EstimatedValue   float64            `json:"estimated_value" bson:"estimated_value"`
	ReportedBy       primitive.ObjectID `json:"reported_by" bson:"reported_by"`
	ReportedAt       time.Time          `json:"reported_at" bson:"reported_at"`
	WitnessInfo      string             `json:"witness_info" bson:"witness_info"`
	RepairEstimate   float64            `json:"repair_estimate" bson:"repair_estimate"`
	IsInsuranceClaim bool               `json:"is_insurance_claim" bson:"is_insurance_claim"`
	ClaimNumber      string             `json:"claim_number" bson:"claim_number"`
}
