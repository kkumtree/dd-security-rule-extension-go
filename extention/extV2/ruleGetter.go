package extV2

import (
	_context "context"
	_encodingjson "encoding/json"
	_fmt "fmt"
	_nethttp "net/http"
	_strings "strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// extractSimplifiedRule extracts only id, isDefault, and name from a rule object
func extractSimplifiedRule(ruleData interface{}) (*SimplifiedRule, error) {
	// Convert to JSON and back to map for easier field extraction
	jsonBytes, err := _encodingjson.Marshal(ruleData)
	if err != nil {
		return nil, _fmt.Errorf("failed to marshal rule data: %v", err)
	}

	var ruleMap map[string]interface{}
	if err := _encodingjson.Unmarshal(jsonBytes, &ruleMap); err != nil {
		return nil, _fmt.Errorf("failed to unmarshal rule data: %v", err)
	}

	rule := &SimplifiedRule{}

	// Extract ID
	if id, ok := ruleMap["id"].(string); ok {
		rule.ID = id
	}

	// Extract IsDefault
	if isDefault, ok := ruleMap["isDefault"].(bool); ok {
		rule.IsDefault = isDefault
	}

	// Extract Name
	if name, ok := ruleMap["name"].(string); ok {
		rule.Name = name
	}

	return rule, nil
}

// extractTagsFromRule extracts tags from a rule object
func extractTagsFromRule(ruleData interface{}) ([]string, error) {
	jsonBytes, err := _encodingjson.Marshal(ruleData)
	if err != nil {
		return nil, _fmt.Errorf("failed to marshal rule data: %v", err)
	}

	var ruleMap map[string]interface{}
	if err := _encodingjson.Unmarshal(jsonBytes, &ruleMap); err != nil {
		return nil, _fmt.Errorf("failed to unmarshal rule data: %v", err)
	}

	var tags []string
	if tagArray, ok := ruleMap["tags"].([]interface{}); ok {
		for _, tag := range tagArray {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	return tags, nil
}

// matchesTagFilters checks if any of the rule's tags match the configured filters (case-insensitive)
func matchesTagFilters(ruleTags []string, tagFilters []string) bool {
	// If no filters are configured, include all rules
	if len(tagFilters) == 0 {
		return true
	}

	// Check if any rule tag contains any of the filter keywords (case-insensitive)
	for _, ruleTag := range ruleTags {
		lowerRuleTag := _strings.ToLower(ruleTag)
		for _, filter := range tagFilters {
			lowerFilter := _strings.ToLower(filter)
			if _strings.Contains(lowerRuleTag, lowerFilter) {
				return true
			}
		}
	}

	return false
}

// FormatSimplifiedResult formats the simplified result as JSON string
func FormatSimplifiedResult(result *PaginatedResult) (string, error) {
	jsonBytes, err := _encodingjson.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", _fmt.Errorf("failed to format simplified result: %v", err)
	}
	return string(jsonBytes), nil
}

// GetExistingStandardRuleTags fetches existing tags for a security monitoring rule
func GetExistingStandardRuleTags(ctx _context.Context, api *datadogV2.SecurityMonitoringApi, ruleID string) ([]string, error) {
	rule, _, err := api.GetSecurityMonitoringRule(ctx, ruleID)
	if err != nil {
		return nil, _fmt.Errorf("failed to get rule %s: %v", ruleID, err)
	}

	// Extract tags from the rule
	if rule.SecurityMonitoringStandardRuleResponse.Tags != nil {
		return rule.SecurityMonitoringStandardRuleResponse.Tags, nil
	}

	return []string{}, nil
}

// ProcessRuleListing fetches all rules with pagination
func ProcessRuleListing(ctx _context.Context, api *datadogV2.SecurityMonitoringApi, config PaginationConfig) (*PaginatedResult, error) {
	result := &PaginatedResult{
		Rules: make([]SimplifiedRule, 0),
		// Rules: make([]interface{}, 0),
	}

	pageNumber := int64(0)
	ruleCondition := len(config.TagFilters)
	ruleCounter := 0
	totalProcessedRules := 0

	for {
		_fmt.Printf("Fetching page %d (size: %d)...\n", pageNumber+1, config.PageSize)

		// Create pagination parameters
		params := datadogV2.NewListSecurityMonitoringRulesOptionalParameters()
		params.PageSize = &config.PageSize
		params.PageNumber = &pageNumber

		// Make API call
		apiCall := NewAPICall("SecurityMonitoringApi", api.ListSecurityMonitoringRules)
		resp, _, err := apiCall.CallWithErrorHandling(func() (interface{}, *_nethttp.Response, error) {
			return api.ListSecurityMonitoringRules(ctx, *params)
		})

		if err != nil {
			return nil, _fmt.Errorf("failed to fetch page %d: %v", pageNumber, err)
		}

		// Parse response to extract rules
		respBytes, err := _encodingjson.Marshal(resp)
		if err != nil {
			return nil, _fmt.Errorf("failed to marshal response: %v", err)
		}

		var responseData map[string]interface{}
		if err := _encodingjson.Unmarshal(respBytes, &responseData); err != nil {
			return nil, _fmt.Errorf("failed to unmarshal response: %v", err)
		}

		// Extract data array
		if data, ok := responseData["data"].([]interface{}); ok {
			if len(data) == 0 {
				_fmt.Println("No more data found. Stopping pagination.")
				break
			}

			// Process each rule and extract only required fields
			pageRules := make([]SimplifiedRule, 0, len(data))
			filteredCount := 0

			for _, ruleData := range data {
				totalProcessedRules++

				ruleTags, err := extractTagsFromRule(ruleData)
				if err != nil || !matchesTagFilters(ruleTags, config.TagFilters) {
					continue
				}

				filteredCount++

				simplifiedRule, err := extractSimplifiedRule(ruleData)
				if err != nil {
					_fmt.Printf("Warning: failed to extract rule data: %v\n", err)
					continue
				}
				pageRules = append(pageRules, *simplifiedRule)
			}

			result.Rules = append(result.Rules, pageRules...)
			result.TotalRules += len(data)
			ruleCounter += filteredCount
			_fmt.Printf("Fetched %d rules from page %d\n", len(data), pageNumber+1)

			// Check if we got fewer rules than requested (last page)
			if int64(len(data)) < config.PageSize {
				_fmt.Println("Last page reached (fewer rules than page size).")
				break
			}
		} else {
			_fmt.Println("No data array found in response. Stopping pagination.")
			break
		}

		pageNumber++
		result.TotalPages++

		// Check max pages limit
		if config.MaxPages > 0 && pageNumber >= config.MaxPages {
			_fmt.Printf("Reached maximum pages limit (%d). Stopping.\n", config.MaxPages)
			break
		}
	}
	if ruleCondition > 0 {
		result.TotalRules = ruleCounter
	}

	// Save result
	if _, err := SaveResultToFile(result, "ListRulesResult", "output", FormatSimplifiedResultAny); err != nil {
		_fmt.Printf("Warning: failed to save listing result: %v\n", err)
	}

	return result, nil
}
