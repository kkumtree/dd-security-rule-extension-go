package extV2

import (
	_context "context"
	_encodingjson "encoding/json"
	_fmt "fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// TaggingResult represents the result of a single rule tagging operation
type TaggingResult struct {
	RuleID   string   `json:"ruleId"`
	RuleName string   `json:"ruleName"`
	Success  bool     `json:"success"`
	OldTags  []string `json:"oldTags,omitempty"`
	NewTags  []string `json:"newTags,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// BatchTaggingResult represents the result of batch tagging operation
type BatchTaggingResult struct {
	TotalRules     int             `json:"totalRules"`
	SuccessfulTags int             `json:"successfulTags"`
	FailedTags     int             `json:"failedTags"`
	Results        []TaggingResult `json:"results"`
	SkippedRules   []string        `json:"skippedRules"`
}

// MergeTags merges new tags with existing tags based on configuration
func MergeTags(existingTags []string, newTags []string, config TaggingConfig) []string {
	if config.OverwriteTags {
		// Filter out included tags from new tags
		var filteredTags []string
		for _, tag := range newTags {
			included := false
			for _, includedTag := range config.IncludedTags {
				if tag == includedTag {
					included = true
					break
				}
			}
			if !included {
				filteredTags = append(filteredTags, tag)
			}
		}
		return filteredTags
	}

	// Append mode: combine existing and new tags, removing duplicates
	tagMap := make(map[string]bool)
	var mergedTags []string

	// Add existing tags first
	for _, tag := range existingTags {
		if !tagMap[tag] {
			tagMap[tag] = true
			mergedTags = append(mergedTags, tag)
		}
	}

	// Add new tags
	for _, tag := range newTags {
		included := false
		for _, includedTag := range config.IncludedTags {
			if tag == includedTag {
				included = true
				break
			}
		}
		if !included && !tagMap[tag] {
			tagMap[tag] = true
			mergedTags = append(mergedTags, tag)
		}
	}

	return mergedTags
}

// FormatTaggingResult formats the tagging result as JSON string
func FormatTaggingResult(batchResult *BatchTaggingResult) string {
	jsonBytes, _ := _encodingjson.MarshalIndent(batchResult, "", "  ")
	return string(jsonBytes)
}

// FormatTaggingSummary formats a summary of the tagging results
func FormatTaggingSummary(batchResult *BatchTaggingResult, config TaggingConfig) string {
	mode := "LIVE"
	if config.DryRun {
		mode = "DRY RUN"
	}

	successRate := 0.0
	if batchResult.TotalRules > 0 {
		successRate = float64(batchResult.SuccessfulTags) / float64(batchResult.TotalRules) * 100
	}

	summary := _fmt.Sprintf(`
=== Rule Tagging Summary (%s) ===
Total Rules Processed: %d
Successfully Tagged: %d
Failed to Tag: %d
Skipped Rules: %d
Success Rate: %.2f%%
`,
		mode,
		batchResult.TotalRules,
		batchResult.SuccessfulTags,
		batchResult.FailedTags,
		len(batchResult.SkippedRules),
		successRate,
	)
	return summary
}

// TagSingleStandardRule tags a single security monitoring rule
func TagSingleStandardRule(ctx _context.Context, api *datadogV2.SecurityMonitoringApi, matchedRule MatchedRule, config TaggingConfig) TaggingResult {
	result := TaggingResult{
		RuleID:   matchedRule.ID,
		RuleName: matchedRule.Name,
		Success:  false,
	}

	// Get existing tags
	existingTags, err := GetExistingStandardRuleTags(ctx, api, matchedRule.ID)
	if err != nil {
		result.Error = _fmt.Sprintf("Failed to get existing tags: %v", err)
		return result
	}

	result.OldTags = existingTags

	// Merge tags
	newTags := MergeTags(existingTags, matchedRule.Tags, config)
	result.NewTags = newTags

	// If dry run, don't make actual API call
	if config.DryRun {
		result.Success = true
		return result
	}

	// Create update payload with only tags changed
	updatePayload := datadogV2.SecurityMonitoringRuleUpdatePayload{
		Tags: newTags,
	}

	// Update the rule
	_, _, err = api.UpdateSecurityMonitoringRule(ctx, matchedRule.ID, updatePayload)
	if err != nil {
		result.Error = _fmt.Sprintf("Failed to update rule: %v", err)
		return result
	}

	result.Success = true
	return result
}

// TagRulesFromMatchResult tags all rules from a MatchResult
func TagRulesFromMatchResult(ctx _context.Context, api *datadogV2.SecurityMonitoringApi, matchResult *MatchResult, config TaggingConfig) (*BatchTaggingResult, error) {
	batchResult := &BatchTaggingResult{
		TotalRules:   len(matchResult.MatchedRules),
		Results:      []TaggingResult{},
		SkippedRules: []string{},
	}

	if config.DryRun {
		_fmt.Println("🔍 DRY RUN MODE - No actual changes will be made")
	}

	_fmt.Printf("Starting to tag %d rules...\n", batchResult.TotalRules)

	// Process each matched rule
	for i, matchedRule := range matchResult.MatchedRules {
		_fmt.Printf("Processing rule %d/%d: %s (ID: %s)\n",
			i+1, batchResult.TotalRules, matchedRule.Name, matchedRule.ID)

		// Skip rules with no tags to add
		if len(matchedRule.Tags) == 0 {
			_fmt.Printf("  ⏭️  Skipping rule with no tags: %s\n", matchedRule.ID)
			batchResult.SkippedRules = append(batchResult.SkippedRules, matchedRule.ID)
			continue
		}

		// Tag the rule
		result := TagSingleStandardRule(ctx, api, matchedRule, config)
		batchResult.Results = append(batchResult.Results, result)

		if result.Success {
			batchResult.SuccessfulTags++
			if config.DryRun {
				_fmt.Printf("  ✅ Would add tags: %v\n", matchedRule.Tags)
			} else {
				_fmt.Printf("  ✅ Successfully tagged with: %v\n", matchedRule.Tags)
			}
		} else {
			batchResult.FailedTags++
			_fmt.Printf("  ❌ Failed to tag: %s\n", result.Error)
		}
	}

	return batchResult, nil
}

// ProcessRuleTagging processes the complete rule tagging workflow
func ProcessRuleTagging(ctx _context.Context, api *datadogV2.SecurityMonitoringApi, matchResult *MatchResult, config TaggingConfig) (*BatchTaggingResult, error) {
	_fmt.Println("Starting rule tagging process...")

	// Perform tagging
	batchResult, err := TagRulesFromMatchResult(ctx, api, matchResult, config)
	if err != nil {
		return nil, _fmt.Errorf("failed to tag rules: %v", err)
	}

	// Save result
	if _, err := SaveResultToFile(batchResult, "TaggingResult", "output", FormatSimplifiedResultAny); err != nil {
		_fmt.Printf("Warning: failed to save tagging result: %v\n", err)
	}

	// Display summary
	_fmt.Println(FormatTaggingSummary(batchResult, config))

	return batchResult, nil
}
