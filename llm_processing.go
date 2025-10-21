package main

import (
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/genai"
)

// processUserReminder handles LLM parsing and task creation in a goroutine
func processUserReminder(ctx context.Context, bot *tgbotapi.BotAPI, aiClient *genai.Client, user *User, messageText string, chatID int64) {
	// Add a small delay to ensure the immediate response is sent first, if needed, though 'go' keyword handles this
	time.Sleep(50 * time.Millisecond)

	payload, err := ParseReminder(ctx, aiClient, messageText, user.Timezone)
	if err != nil {
		log.Printf("Error parsing reminder for user %d: %v", user.TelegramID, err)
		// Optionally, send a follow-up error message if LLM completely failed
		// sendErrorMessage(bot, chatID, "Sorry, I had trouble understanding your message. Please try again.", nil)
		return
	}

	log.Printf("Parsed reminder payload for user %d: %+v", user.TelegramID, payload)

	if payload.Type == "task" {
		task, err := CreateTask(user.ID, payload)
		if err != nil {
			log.Printf("Error creating task for user %d: %v", user.TelegramID, err)
			// Optionally, send a follow-up error message if DB saving failed
			// sendErrorMessage(bot, chatID, "I understood your reminder, but had trouble saving it. Please try again.", nil)
			return
		}

		// At this point, the task is scheduled. The immediate response already told the user.
		// We could send a more detailed confirmation, or rely on the /mytasks command.
		// For now, we'll just log and not send an additional message to avoid spamming the user
		// after the immediate confirmation.
		log.Printf("Task '%s' created for user %d. Due: %s", task.Title, user.TelegramID, task.DueDateTime.Format(time.RFC3339))

	} else {
		// If it's not a task, the LLM usually provides a conversational response.
		// Since we sent an immediate generic message, this LLM conversational response
		// will currently be lost. If you want to send it, you'd need to edit the previous message
		// or send a follow-up. For a simple "Task Scheduled" response, we ignore it.
		log.Printf("LLM determined message for user %d was not a task: %s", user.TelegramID, payload.LLMMessage)
	}
}
