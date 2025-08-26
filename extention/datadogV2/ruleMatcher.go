package extV2

import (
	_encodingjson "encoding/json"
	_fmt "fmt"
	_os "os"
)

// InputRule represents a rule from input.json (in rules array)
type InputRule struct {
	ID        string   `json:"id,omitempty"`
	IsDefault bool     `json:"isDefault"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags,omitempty"`
}

// InputData represents the structure of input.json
type InputData struct {
	TotalRules     int         `json:"totalRules"`
	ProcessedRules int         `json:"processedRules"`
	FailedRules    []string    `json:"failedRules"`
	Rules          []InputRule `json:"rules"`
}

// MatchedRule represents the result of matching rules
type MatchedRule struct {
	ID        string   `json:"id"`        // from result
	Name      string   `json:"name"`      // from input.json
	Tags      []string `json:"tags"`      // from input.json
	IsDefault bool     `json:"isDefault"` // matched value
}

// MatchResult represents the final matching result
type MatchResult struct {
	TotalMatches     int           `json:"totalMatches"`
	TotalInputRules  int           `json:"totalInputRules"`
	TotalResultRules int           `json:"totalResultRules"`
	MatchedRules     []MatchedRule `json:"matchedRules"`
}

// LoadInputJSON loads and parses the input JSON file with better error handling
func LoadInputJSON(filename string) (*InputData, error) {
	// Check if file exists
	if _, err := _os.Stat(filename); _os.IsNotExist(err) {
		return nil, _fmt.Errorf("file %s does not exist", filename)
	}

	// Read file
	data, err := _os.ReadFile(filename)
	if err != nil {
		return nil, _fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Check if file is empty
	if len(data) == 0 {
		return nil, _fmt.Errorf("file %s is empty", filename)
	}

	// Debug: Print first 200 characters of the file
	_fmt.Printf("First 200 chars of %s: %s\n", filename, string(data[:min(200, len(data))]))

	var inputData InputData
	if err := _encodingjson.Unmarshal(data, &inputData); err != nil {
		return nil, _fmt.Errorf("failed to parse JSON from %s: %v", filename, err)
	}

	// Debug: Print parsed data summary
	_fmt.Printf("Parsed input data: totalRules=%d, processedRules=%d, rules count=%d\n",
		inputData.TotalRules, inputData.ProcessedRules, len(inputData.Rules))

	return &inputData, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MatchRules compares input.json rules with ProcessRuleListing result
func MatchRules(inputData *InputData, resultData *PaginatedResult) (*MatchResult, error) {
	matchResult := &MatchResult{
		TotalInputRules:  len(inputData.Rules),  // input.json uses "rules"
		TotalResultRules: len(resultData.Rules), // result uses "rules"
		MatchedRules:     []MatchedRule{},
	}

	// Create maps for efficient lookup
	inputRuleMap := make(map[string]InputRule)
	resultRuleMap := make(map[string]SimplifiedRule)

	// Index input rules by name+isDefault combination (from "results" array)
	for _, inputRule := range inputData.Rules {
		key := _fmt.Sprintf("%s_%t", inputRule.Name, inputRule.IsDefault)
		inputRuleMap[key] = inputRule
	}

	// Index result rules by name+isDefault combination (from "rules" array)
	for _, resultRule := range resultData.Rules {
		key := _fmt.Sprintf("%s_%t", resultRule.Name, resultRule.IsDefault)
		resultRuleMap[key] = resultRule
	}

	_fmt.Printf("Input rules indexed: %d\n", len(inputRuleMap))
	_fmt.Printf("Result rules indexed: %d\n", len(resultRuleMap))

	// Find matches
	matchedInputKeys := make(map[string]bool)
	matchedResultKeys := make(map[string]bool)

	for key, inputRule := range inputRuleMap {
		if resultRule, exists := resultRuleMap[key]; exists {
			// Found a match
			matchedRule := MatchedRule{
				ID:        resultRule.ID,       // from result
				Name:      inputRule.Name,      // from input
				Tags:      inputRule.Tags,      // from input
				IsDefault: inputRule.IsDefault, // matched value
			}
			matchResult.MatchedRules = append(matchResult.MatchedRules, matchedRule)
			matchedInputKeys[key] = true
			matchedResultKeys[key] = true
		}
	}

	matchResult.TotalMatches = len(matchResult.MatchedRules)

	return matchResult, nil
}

// ProcessRuleMatching processes the complete rule matching workflow with better error handling
func ProcessRuleMatching(inputFilename string, resultData *PaginatedResult) (*MatchResult, error) {
	// Load input JSON
	_fmt.Printf("Loading input file: %s\n", inputFilename)
	inputData, err := LoadInputJSON(inputFilename)
	if err != nil {
		return nil, _fmt.Errorf("failed to load input JSON: %v", err)
	}

	// Perform matching
	_fmt.Println("Performing rule matching...")
	matchResult, err := MatchRules(inputData, resultData)
	if err != nil {
		return nil, _fmt.Errorf("failed to match rules: %v", err)
	}

	// Save result
	if _, err := SaveResultToFile(matchResult, "MatchResult", "output", FormatSimplifiedResultAny); err != nil {
		return nil, _fmt.Errorf("failed to save match result: %v", err)
	}

	// Display summary
	_fmt.Println(FormatMatchSummary(matchResult))

	return matchResult, nil
}

// FormatMatchSummary formats a summary of the matching results
func FormatMatchSummary(matchResult *MatchResult) string {
	matchRate := 0.0
	if matchResult.TotalInputRules > 0 {
		matchRate = float64(matchResult.TotalMatches) / float64(matchResult.TotalInputRules) * 100
	}

	summary := _fmt.Sprintf(`
=== Rule Matching Summary ===
Total Input Rules: %d
Total Result Rules: %d
Total Matches: %d
Match Rate: %.2f%%
`,
		matchResult.TotalInputRules,
		matchResult.TotalResultRules,
		matchResult.TotalMatches,
		matchRate,
	)
	return summary
}
