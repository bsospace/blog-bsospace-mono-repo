package social

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type SocialMediaService struct{}

type SocialMediaProfile struct {
	Platform string `json:"platform"`
	Username string `json:"username"`
	URL      string `json:"url"`
	Valid    bool   `json:"valid"`
}

func NewSocialMediaService() *SocialMediaService {
	return &SocialMediaService{}
}

// ValidateSocialMediaLink validates and formats social media links
func (s *SocialMediaService) ValidateSocialMediaLink(platform, value string) (*SocialMediaProfile, error) {
	if value == "" {
		return nil, nil
	}

	// Clean the value
	value = strings.TrimSpace(value)

	// Handle website separately (not a social media platform)
	if platform == "website" {
		return s.validateWebsite(value)
	}

	// Remove common prefixes if user included them
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "www.")

	// Validate as username
	switch strings.ToLower(platform) {
	case "github":
		username, err := s.validateGitHub(value)
		if err != nil {
			return nil, err
		}
		return &SocialMediaProfile{
			Platform: "github",
			Username: username,
			URL:      fmt.Sprintf("https://github.com/%s", username),
			Valid:    true,
		}, nil
	case "twitter":
		return s.validateTwitter(value)
	case "linkedin":
		return s.validateLinkedIn(value)
	case "instagram":
		return s.validateInstagram(value)
	case "facebook":
		return s.validateFacebook(value)
	case "youtube":
		return s.validateYouTube(value)
	case "discord":
		return s.validateDiscord(value)
	case "telegram":
		return s.validateTelegram(value)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// validateGitHub validates GitHub username
func (s *SocialMediaService) validateGitHub(value string) (string, error) {
	// GitHub username validation: alphanumeric and hyphens, 1-39 chars, no consecutive hyphens
	// Fixed regex to be compatible with Go's regexp package
	githubRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

	// Check if it's already a GitHub URL
	if strings.Contains(value, "github.com/") {
		// Extract username from URL
		parts := strings.Split(value, "github.com/")
		if len(parts) > 1 {
			username := strings.TrimSuffix(parts[1], "/")
			// Clean username (remove query params, etc.)
			username = strings.Split(username, "?")[0]
			username = strings.Split(username, "#")[0]

			// Validate extracted username
			if !githubRegex.MatchString(username) {
				return "", fmt.Errorf("invalid GitHub username format in URL")
			}

			if len(username) < 1 || len(username) > 39 {
				return "", fmt.Errorf("GitHub username must be between 1 and 39 characters")
			}

			// Check for consecutive hyphens
			if strings.Contains(username, "--") {
				return "", fmt.Errorf("GitHub username cannot contain consecutive hyphens")
			}

			// Check if starts or ends with hyphen
			if strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") {
				return "", fmt.Errorf("GitHub username cannot start or end with hyphen")
			}

			return username, nil
		}
		return "", fmt.Errorf("invalid GitHub URL format")
	}

	// Validate as username
	if !githubRegex.MatchString(value) {
		return "", fmt.Errorf("invalid GitHub username format")
	}

	if len(value) < 1 || len(value) > 39 {
		return "", fmt.Errorf("GitHub username must be between 1 and 39 characters")
	}

	// Check for consecutive hyphens
	if strings.Contains(value, "--") {
		return "", fmt.Errorf("GitHub username cannot contain consecutive hyphens")
	}

	// Check if starts or ends with hyphen
	if strings.HasPrefix(value, "-") || strings.HasSuffix(value, "-") {
		return "", fmt.Errorf("GitHub username cannot start or end with hyphen")
	}

	return value, nil
}

// validateTwitter validates Twitter/X username
func (s *SocialMediaService) validateTwitter(value string) (*SocialMediaProfile, error) {
	// Twitter username validation: alphanumeric and underscores, 1-15 chars
	twitterRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{1,15}$`)

	// Check if it's already a Twitter URL
	if strings.Contains(value, "twitter.com/") || strings.Contains(value, "x.com/") {
		// Extract username from URL
		var username string
		if strings.Contains(value, "twitter.com/") {
			parts := strings.Split(value, "twitter.com/")
			if len(parts) > 1 {
				username = strings.TrimSuffix(parts[1], "/")
			}
		} else if strings.Contains(value, "x.com/") {
			parts := strings.Split(value, "x.com/")
			if len(parts) > 1 {
				username = strings.TrimSuffix(parts[1], "/")
			}
		}

		if username != "" {
			// Clean username (remove query params, etc.)
			username = strings.Split(username, "?")[0]
			username = strings.Split(username, "#")[0]

			// Validate extracted username
			if !twitterRegex.MatchString(username) {
				return &SocialMediaProfile{
					Platform: "twitter",
					Username: username,
					URL:      "",
					Valid:    false,
				}, fmt.Errorf("invalid Twitter username format in URL")
			}

			return &SocialMediaProfile{
				Platform: "twitter",
				Username: username,
				URL:      fmt.Sprintf("https://twitter.com/%s", username),
				Valid:    true,
			}, nil
		}
		return nil, fmt.Errorf("invalid Twitter URL format")
	}

	// Validate as username
	if !twitterRegex.MatchString(value) {
		return &SocialMediaProfile{
			Platform: "twitter",
			Username: value,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid Twitter username format")
	}

	return &SocialMediaProfile{
		Platform: "twitter",
		Username: value,
		URL:      fmt.Sprintf("https://twitter.com/%s", value),
		Valid:    true,
	}, nil
}

// validateLinkedIn validates LinkedIn profile URL
func (s *SocialMediaService) validateLinkedIn(value string) (*SocialMediaProfile, error) {
	// LinkedIn username validation: alphanumeric, hyphens, and underscores, 3-100 chars
	linkedinRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,100}$`)

	// Check if it's already a LinkedIn URL
	if strings.Contains(value, "linkedin.com/in/") {
		// Extract username from URL
		parts := strings.Split(value, "linkedin.com/in/")
		if len(parts) > 1 {
			username := strings.TrimSuffix(parts[1], "/")
			// Clean username (remove query params, etc.)
			username = strings.Split(username, "?")[0]
			username = strings.Split(username, "#")[0]

			// Validate extracted username
			if !linkedinRegex.MatchString(username) {
				return &SocialMediaProfile{
					Platform: "linkedin",
					Username: username,
					URL:      "",
					Valid:    false,
				}, fmt.Errorf("invalid LinkedIn username format in URL")
			}

			return &SocialMediaProfile{
				Platform: "linkedin",
				Username: username,
				URL:      fmt.Sprintf("https://linkedin.com/in/%s", username),
				Valid:    true,
			}, nil
		}
		return nil, fmt.Errorf("invalid LinkedIn URL format")
	}

	// Validate as username
	if !linkedinRegex.MatchString(value) {
		return &SocialMediaProfile{
			Platform: "linkedin",
			Username: value,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid LinkedIn username format")
	}

	return &SocialMediaProfile{
		Platform: "linkedin",
		Username: value,
		URL:      fmt.Sprintf("https://linkedin.com/in/%s", value),
		Valid:    true,
	}, nil
}

// validateInstagram validates Instagram username
func (s *SocialMediaService) validateInstagram(username string) (*SocialMediaProfile, error) {
	// Instagram username rules: alphanumeric + dots + underscores, 1-30 chars
	instagramRegex := regexp.MustCompile(`^[a-zA-Z0-9._]{1,30}$`)

	if !instagramRegex.MatchString(username) {
		return &SocialMediaProfile{
			Platform: "instagram",
			Username: username,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid Instagram username format")
	}

	return &SocialMediaProfile{
		Platform: "instagram",
		Username: username,
		URL:      fmt.Sprintf("https://instagram.com/%s", username),
		Valid:    true,
	}, nil
}

// validateFacebook validates Facebook profile URL
func (s *SocialMediaService) validateFacebook(profile string) (*SocialMediaProfile, error) {
	// Facebook profile URL or username
	if strings.Contains(profile, "facebook.com") {
		// Full URL provided
		if _, err := url.Parse(profile); err != nil {
			return &SocialMediaProfile{
				Platform: "facebook",
				Username: profile,
				URL:      "",
				Valid:    false,
			}, fmt.Errorf("invalid Facebook URL")
		}
		return &SocialMediaProfile{
			Platform: "facebook",
			Username: profile,
			URL:      profile,
			Valid:    true,
		}, nil
	}

	// Username provided
	facebookRegex := regexp.MustCompile(`^[a-zA-Z0-9.]{5,50}$`)
	if !facebookRegex.MatchString(profile) {
		return &SocialMediaProfile{
			Platform: "facebook",
			Username: profile,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid Facebook username format")
	}

	return &SocialMediaProfile{
		Platform: "facebook",
		Username: profile,
		URL:      fmt.Sprintf("https://facebook.com/%s", profile),
		Valid:    true,
	}, nil
}

// validateYouTube validates YouTube channel URL
func (s *SocialMediaService) validateYouTube(channel string) (*SocialMediaProfile, error) {
	// YouTube channel URL or username
	if strings.Contains(channel, "youtube.com") || strings.Contains(channel, "youtu.be") {
		// Full URL provided
		if _, err := url.Parse(channel); err != nil {
			return &SocialMediaProfile{
				Platform: "youtube",
				Username: channel,
				URL:      "",
				Valid:    false,
			}, fmt.Errorf("invalid YouTube URL")
		}
		return &SocialMediaProfile{
			Platform: "youtube",
			Username: channel,
			URL:      channel,
			Valid:    true,
		}, nil
	}

	// Username provided
	youtubeRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	if !youtubeRegex.MatchString(channel) {
		return &SocialMediaProfile{
			Platform: "youtube",
			Username: channel,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid YouTube username format")
	}

	return &SocialMediaProfile{
		Platform: "youtube",
		Username: channel,
		URL:      fmt.Sprintf("https://youtube.com/@%s", channel),
		Valid:    true,
	}, nil
}

// validateDiscord validates Discord username
func (s *SocialMediaService) validateDiscord(username string) (*SocialMediaProfile, error) {
	// Discord username format: username#discriminator (old) or username (new)
	discordRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{2,32}(?:#[0-9]{4})?$`)

	if !discordRegex.MatchString(username) {
		return &SocialMediaProfile{
			Platform: "discord",
			Username: username,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid Discord username format")
	}

	return &SocialMediaProfile{
		Platform: "discord",
		Username: username,
		URL:      username, // Discord doesn't have public profile URLs
		Valid:    true,
	}, nil
}

// validateTelegram validates Telegram username
func (s *SocialMediaService) validateTelegram(username string) (*SocialMediaProfile, error) {
	// Telegram username rules: alphanumeric + underscore, 5-32 chars, must start with letter
	telegramRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{4,31}$`)

	if !telegramRegex.MatchString(username) {
		return &SocialMediaProfile{
			Platform: "telegram",
			Username: username,
			URL:      "",
			Valid:    false,
		}, fmt.Errorf("invalid Telegram username format")
	}

	return &SocialMediaProfile{
		Platform: "telegram",
		Username: username,
		URL:      fmt.Sprintf("https://t.me/%s", username),
		Valid:    true,
	}, nil
}

// validateWebsite validates website URL
func (s *SocialMediaService) validateWebsite(value string) (*SocialMediaProfile, error) {
	// Basic website validation
	if value == "" {
		return nil, nil
	}

	// Add protocol if missing
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		value = "https://" + value
	}

	// Basic URL validation
	if !strings.Contains(value, ".") {
		return nil, fmt.Errorf("invalid website URL format")
	}

	return &SocialMediaProfile{
		Platform: "website",
		Username: value,
		URL:      value,
		Valid:    true,
	}, nil
}

// ValidateAllSocialMedia validates all social media fields
func (s *SocialMediaService) ValidateAllSocialMedia(data map[string]string) (map[string]*SocialMediaProfile, []string) {
	profiles := make(map[string]*SocialMediaProfile)
	var errors []string

	// Handle website separately
	if website, exists := data["website"]; exists && website != "" {
		profile, err := s.validateWebsite(website)
		if err != nil {
			errors = append(errors, fmt.Sprintf("website: %s", err.Error()))
		} else if profile != nil {
			profiles["website"] = profile
		}
	}

	// Handle social media platforms
	for platform, value := range data {
		if platform == "website" || value == "" {
			continue // Skip website (already handled) and empty values
		}

		profile, err := s.ValidateSocialMediaLink(platform, value)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", platform, err.Error()))
		} else if profile != nil {
			profiles[platform] = profile
		}
	}

	return profiles, errors
}
