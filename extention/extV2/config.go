package extV2

import (
	_bufio "bufio"
	_fmt "fmt"
	_os "os"
	_strconv "strconv"
	_strings "strings"
)

// SimplifiedRule represents a simplified security monitoring rule with only essential fields
type SimplifiedRule struct {
	ID        string `json:"id"`
	IsDefault bool   `json:"isDefault"`
	Name      string `json:"name"`
}

// PaginatedResult holds the results from all pages
type PaginatedResult struct {
	TotalRules int              `json:"totalRules"`
	TotalPages int              `json:"totalPages"`
	Rules      []SimplifiedRule `json:"rules"`
}

// PaginationConfig holds pagination settings
type PaginationConfig struct {
	PageSize   int64
	MaxPages   int64    // 0 means no limit
	TagFilters []string // Optional tag filters (case-insenitive)
}

// TaggingConfig holds configuration for rule tagging
type TaggingConfig struct {
	DryRun         bool     // If true, only simulate tagging without actual API calls
	OverwriteTags  bool     // If true, replace existing tags; if false, append to existing tags
	IncludedTags   []string // Tags to exclude from tagging (e.g., system tags)
	MaxConcurrency int      // Maximum number of concurrent API calls
}

// Config holds application configuration from environment variables
type Config struct {
	DDSite            string
	DDAPIKey          string
	DDAppKey          string
	InputRuleFilename string
	Pagination        PaginationConfig
	Tagging           TaggingConfig
}

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(filename string) error {
	file, err := _os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := _bufio.NewScanner(file)
	for scanner.Scan() {
		line := _strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || _strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first '=' character
		parts := _strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := _strings.TrimSpace(parts[0])
		value := _strings.TrimSpace(parts[1])

		// Only set if environment variable is not already set
		if _os.Getenv(key) == "" {
			_os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadConfig loads configuration with .env file support
func LoadConfig() (*Config, error) {
	// Try to load .env file (optional)
	if err := LoadEnvFile(".env"); err != nil {
		_fmt.Printf("Note: .env file not found or error loading: %v\n", err)
	}
	// Parse PageSize with default value
	pageSize := int64(100) // Default value
	if pageSizeStr := _os.Getenv("PAGE_SIZE"); pageSizeStr != "" {
		if parsed, err := _strconv.ParseInt(pageSizeStr, 10, 64); err == nil {
			pageSize = parsed
		}
	}

	// Parse MaxPages with default value (0 means no limit)
	maxPages := int64(0) // Default value
	if maxPagesStr := _os.Getenv("MAX_PAGES"); maxPagesStr != "" {
		if parsed, err := _strconv.ParseInt(maxPagesStr, 10, 64); err == nil {
			maxPages = parsed
		}
	}

	// Parse TagFilters from environment variable (comma-separated)
	var tagFilters []string
	if tagFiltersStr := _os.Getenv("TAG_FILTERS"); tagFiltersStr != "" {
		filters := _strings.Split(tagFiltersStr, ",")
		for _, filter := range filters {
			trimmed := _strings.TrimSpace(filter)
			if trimmed != "" {
				tagFilters = append(tagFilters, trimmed)
			}
		}
	}

	// Parse DryRun setting from environment variable
	dryRun := false // default value
	if dryRunStr := _os.Getenv("DRYRUN"); dryRunStr != "" {
		if parsed, err := _strconv.ParseBool(dryRunStr); err == nil {
			dryRun = parsed
		}
	}

	// Parse other tagging configurations
	overwriteTags := false // default value (append mode)
	if overwriteStr := _os.Getenv("OVERWRITE_TAGS"); overwriteStr != "" {
		if parsed, err := _strconv.ParseBool(overwriteStr); err == nil {
			overwriteTags = parsed
		}
	}

	var includedTags []string
	if includedTagsStr := _os.Getenv("INCLUDED_TAGS"); includedTagsStr != "" {
		tags := _strings.Split(includedTagsStr, ",")
		for _, tag := range tags {
			trimmed := _strings.TrimSpace(tag)
			if trimmed != "" {
				includedTags = append(includedTags, trimmed)
			}
		}
	}

	maxConcurrency := 5 // default value
	if concurrencyStr := _os.Getenv("MAX_CONCURRENCY"); concurrencyStr != "" {
		if parsed, err := _strconv.Atoi(concurrencyStr); err == nil && parsed > 0 {
			maxConcurrency = parsed
		}
	}

	inputRuleFilename := "input.json"
	if inputrulefilenameStr := _os.Getenv("INPUT"); inputrulefilenameStr != "" {
		inputRuleFilename = inputrulefilenameStr
	}

	config := &Config{
		DDSite:            _os.Getenv("DD_SITE"),
		DDAPIKey:          _os.Getenv("DD_API_KEY"),
		DDAppKey:          _os.Getenv("DD_APP_KEY"),
		InputRuleFilename: inputRuleFilename,
		Pagination: PaginationConfig{
			PageSize:   pageSize,
			MaxPages:   maxPages,
			TagFilters: tagFilters,
		},
		Tagging: TaggingConfig{
			DryRun:         dryRun,
			OverwriteTags:  overwriteTags,
			IncludedTags:   includedTags,
			MaxConcurrency: maxConcurrency,
		},
	}

	// Validate required environment variables
	if config.DDAPIKey == "" {
		return nil, _fmt.Errorf("DD_API_KEY environment variable is required")
	}

	if config.DDAppKey == "" {
		return nil, _fmt.Errorf("DD_APP_KEY environment variable is required")
	}

	// Set default DD_SITE if not provided
	if config.DDSite == "" {
		config.DDSite = "datadoghq.com"
		_fmt.Printf("DD_SITE not set, using default: %s\n", config.DDSite)
	}

	return config, nil
}

// SetDatadogEnvironment sets Datadog environment variables for the API client
func (c *Config) SetDatadogEnvironment() {
	_os.Setenv("DD_SITE", c.DDSite)
	_os.Setenv("DD_API_KEY", c.DDAPIKey)
	_os.Setenv("DD_APP_KEY", c.DDAppKey)
}
