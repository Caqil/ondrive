package utils

import (
	"math"
)

const (
	// EarthRadiusKM is the radius of the Earth in kilometers
	EarthRadiusKM = 6371.0
	// EarthRadiusMiles is the radius of the Earth in miles
	EarthRadiusMiles = 3959.0
)

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// CalculateDistance calculates the distance between two points using Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Haversine formula
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	distance := EarthRadiusKM * c
	return distance
}

// CalculateDistanceInMiles calculates distance in miles
func CalculateDistanceInMiles(lat1, lon1, lat2, lon2 float64) float64 {
	distanceKM := CalculateDistance(lat1, lon1, lat2, lon2)
	return distanceKM * 0.621371 // Convert KM to miles
}

// CalculateBearing calculates the initial bearing from point 1 to point 2
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLonRad := degreesToRadians(lon2 - lon1)

	x := math.Sin(deltaLonRad) * math.Cos(lat2Rad)
	y := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLonRad)

	bearingRad := math.Atan2(x, y)
	bearingDeg := radiansToDegrees(bearingRad)

	// Normalize to 0-360 degrees
	return math.Mod(bearingDeg+360, 360)
}

// IsWithinRadius checks if a point is within a certain radius of another point
func IsWithinRadius(lat1, lon1, lat2, lon2, radiusKM float64) bool {
	distance := CalculateDistance(lat1, lon1, lat2, lon2)
	return distance <= radiusKM
}

// GetBoundingBox returns the bounding box for a given center point and radius
func GetBoundingBox(centerLat, centerLon, radiusKM float64) (minLat, maxLat, minLon, maxLon float64) {
	// Approximate degrees per kilometer
	latDegPerKM := 1.0 / 111.0
	lonDegPerKM := 1.0 / (111.0 * math.Cos(degreesToRadians(centerLat)))

	deltaLat := radiusKM * latDegPerKM
	deltaLon := radiusKM * lonDegPerKM

	minLat = centerLat - deltaLat
	maxLat = centerLat + deltaLat
	minLon = centerLon - deltaLon
	maxLon = centerLon + deltaLon

	return minLat, maxLat, minLon, maxLon
}

// CalculateETA estimates arrival time based on distance and average speed
func CalculateETA(distanceKM, averageSpeedKMH float64) float64 {
	if averageSpeedKMH <= 0 {
		return 0
	}
	return (distanceKM / averageSpeedKMH) * 60 // Return in minutes
}

// GetMidpoint calculates the midpoint between two locations
func GetMidpoint(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	deltaLon := lon2Rad - lon1Rad

	bx := math.Cos(lat2Rad) * math.Cos(deltaLon)
	by := math.Cos(lat2Rad) * math.Sin(deltaLon)

	midLat := math.Atan2(
		math.Sin(lat1Rad)+math.Sin(lat2Rad),
		math.Sqrt((math.Cos(lat1Rad)+bx)*(math.Cos(lat1Rad)+bx)+by*by),
	)

	midLon := lon1Rad + math.Atan2(by, math.Cos(lat1Rad)+bx)

	return radiansToDegrees(midLat), radiansToDegrees(midLon)
}

// NormalizePhoneNumber formats phone number to international format
func NormalizePhoneNumber(phone string) string {
	// Remove all non-digit characters
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}

	// Add country code if missing (assuming US/international format)
	if len(cleaned) == 10 {
		cleaned = "1" + cleaned
	}

	return "+" + cleaned
}

// ValidateCoordinates checks if latitude and longitude are valid
func ValidateCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// Helper functions
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// CalculateRouteDistance calculates total distance for multiple waypoints
func CalculateRouteDistance(locations []Location) float64 {
	if len(locations) < 2 {
		return 0
	}

	totalDistance := 0.0
	for i := 0; i < len(locations)-1; i++ {
		distance := CalculateDistance(
			locations[i].Latitude, locations[i].Longitude,
			locations[i+1].Latitude, locations[i+1].Longitude,
		)
		totalDistance += distance
	}

	return totalDistance
}
