package main

import "strings"

// cleanJSONResponse removes markdown code blocks and extra whitespace from LLM response
func cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	response = strings.TrimSuffix(response, "```")

	// Remove extra whitespace and newlines
	response = strings.TrimSpace(response)

	return response
}
