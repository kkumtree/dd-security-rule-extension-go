package extV2

import (
	_encodingjson "encoding/json"
	_fmt "fmt"
	_os "os"
	_pathfilepath "path/filepath"
	_time "time"
)

// SaveToJSONFile saves the formatted result to a JSON file
func SaveToJSONFile(data string, filename string) error {
	// Create directory if it doesn't exist
	dir := _pathfilepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := _os.MkdirAll(dir, 0755); err != nil {
			return _fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Write data to file using _os.WriteFile (Go 1.16+)
	if err := _os.WriteFile(filename, []byte(data), 0644); err != nil {
		return _fmt.Errorf("failed to write file %s: %v", filename, err)
	}

	return nil
}

// GenerateTimestampedFilename generates a filename with timestamp
func GenerateTimestampedFilename(prefix string, extension string) string {
	timestamp := _time.Now().Local().Format("2006-01-02_15-04-05")
	return _fmt.Sprintf("%s_%s.%s", timestamp, prefix, extension)
}

func FormatSimplifiedResultAny(result any) (string, error) {
	jsonBytes, err := _encodingjson.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func SaveResultToFile(
	batchResult any,
	prefix string,
	outputDir string,
	formatter func(any) (string, error),
) (string, error) {
	// Generate filename with timestamp
	filename := GenerateTimestampedFilename(prefix, "json")

	// Add output directory if specified
	if outputDir != "" {
		filename = _pathfilepath.Join(outputDir, filename)
	}

	// Format the batch result
	formattedResult, err := formatter(batchResult)
	if err != nil {
		return "", _fmt.Errorf("failed to format batch result: %v", err)
	}

	// Save to file
	if err := SaveToJSONFile(formattedResult, filename); err != nil {
		return "", err
	}

	return filename, nil
}
