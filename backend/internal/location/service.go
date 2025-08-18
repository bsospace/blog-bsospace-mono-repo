package location

import (
	"fmt"
	"strings"
)

type LocationService struct{}

func NewLocationService() *LocationService {
	return &LocationService{}
}

// ValidateLocation validates and formats location input from frontend
func (s *LocationService) ValidateLocation(location string) (string, error) {
	if location == "" {
		return "", nil
	}

	// Clean the input
	location = strings.TrimSpace(location)

	// Convert to title case for consistency
	location = strings.Title(strings.ToLower(location))

	// Common location validations
	if len(location) < 2 || len(location) > 100 {
		return "", fmt.Errorf("location must be between 2 and 100 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(location, "!@#$%^&*()_+=[]{}|\\:;\"'<>?,./") {
		return "", fmt.Errorf("location contains invalid characters")
	}

	return location, nil
}

// GetSupportedRegions returns list of supported regions
func (s *LocationService) GetSupportedRegions() []string {
	return []string{
		"Thailand", "Japan", "China", "South Korea", "Vietnam",
		"Indonesia", "Malaysia", "Singapore", "Philippines", "India",
		"United States", "United Kingdom", "Germany", "France", "Canada",
		"Australia", "Brazil", "Mexico", "Russia", "South Africa",
	}
}

// IsValidRegion checks if the location is in supported regions
func (s *LocationService) IsValidRegion(location string) bool {
	supportedRegions := s.GetSupportedRegions()
	for _, region := range supportedRegions {
		if strings.EqualFold(location, region) {
			return true
		}
	}
	return false
}
