// package main

// import (
// 	"context"
// 	"fmt"
// 	"os"

// 	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
// 	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
// )

// func main() {
// 	// Load configuration from environment variables
// 	config, err := LoadConfig()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
// 		os.Exit(1)
// 	}
// 	// Set Datadog environment variables for the API client
// 	config.SetDatadogEnvironment()

// 	ctx := datadog.NewDefaultContext(context.Background())
// 	configuration := datadog.NewConfiguration()
// 	apiClient := datadog.NewAPIClient(configuration)
// 	api := datadogV2.NewSecurityMonitoringApi(apiClient)

// 	fmt.Println("Processing paginated lists of security monitoring rules...")

// 	listResult, err := ProcessRuleListing(ctx, api, config.Pagination)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Listing error: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Process rule matching with input.json
// 	matchResult, err := ProcessRuleMatching(
// 		"input.json", // input file
// 		listResult,   // result from ProcessRuleListing
// 	)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Rule matching error: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Process rule tagging
// 	fmt.Printf("\n=== Starting Rule Tagging ===\n")
// 	taggingResult, err := ProcessRuleTagging(
// 		ctx,
// 		api,
// 		matchResult,
// 		config.Tagging,
// 	)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Rule tagging error: %v\n", err)
// 		os.Exit(1)
// 	}

// 	fmt.Printf("Tagging process for %d rules completed! Check for details.\n", taggingResult.SuccessfulTags)
// }