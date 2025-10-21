package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// ParseReminder takes a user message and returns a structured ReminderPayload
func ParseReminder(ctx context.Context, client *genai.Client, message string, userTimezone string) (*ReminderPayload, error) {
	// Get current time in user's timezone
	now := time.Now().UTC()
	userTime, err := ConvertToUserTimezone(now, userTimezone)
	if err != nil {
		// Fall back to UTC if timezone conversion fails
		userTime = now
	}

	nowStr := userTime.Format("2006-01-02T15:04:05")

	prompt := fmt.Sprintf(`
		You are an AI that extracts reminders or tasks from user messages and also generates a friendly confirmation message for the user.
		Current date and time: %s %s
		User timezone: %s

		Rules:
		- If the message contains a task/reminder, return JSON strictly as:
		{
			"type": "task",
			"title": string,
			"description": string,
			"datetime": string,
			"timezone": string,
			"recurrence": string|null,
			"source_text": string,
			"llm_message": string
		}
		- If the message does NOT contain a task/reminder, return:
		{
			"type": "not_task",
			"llm_message": "I don't see any task or reminder in your message. If you have any task or reminder, please let me know."
		}
		- Resolve relative dates like "tomorrow", "next Friday", or "in 3 hours" using the current date/time above.
		- The "llm_message" field should be a friendly confirmation, e.g., "Sure, I'll remind you to buy medicine tomorrow at 9 AM"
		- IMPORTANT: Return ONLY valid JSON. Do not wrap in markdown code blocks or add any extra text.

		Examples:

		1) Message: "Remind me to buy medicine tomorrow at 9 AM"
		Response:
		{
			"type": "task",
			"title": "Buy medicine",
			"description": "Reminder to buy medicine",
			"datetime": "2025-10-23T09:00:00",
			"timezone": "UTC",
			"recurrence": null,
			"source_text": "Remind me to buy medicine tomorrow at 9 AM",
			"llm_message": "Sure, I'll remind you to buy medicine tomorrow at 9 AM"
		}

		2) Message: "Hey, how are you?"
		Response:
		{
			"type": "not_task",
			"llm_message": "I don't see any task or reminder in your message. If you have any task or reminder, please let me know."
		}

		3) Message: "Submit the report by 5 PM today"
		Response:
		{
			"type": "task",
			"title": "Submit the report",
			"description": "Submit the report by 5 PM today",
			"datetime": "2025-10-22T17:00:00",
			"timezone": "UTC",
			"recurrence": null,
			"source_text": "Submit the report by 5 PM today",
			"llm_message": "Got it! I will remind you to submit the report by 5 PM today"
		}

		Now analyze this user message and respond:
		Message: "%s"
	`, nowStr, userTimezone, userTimezone, message)

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Clean the response to remove markdown formatting
	cleanedResponse := cleanJSONResponse(result.Text())

	var payload ReminderPayload
	if err := json.Unmarshal([]byte(cleanedResponse), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v, raw response: %s", err, result.Text())
	}

	return &payload, nil
}
