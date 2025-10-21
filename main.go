package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/genai"
)

func main() {
	// Initialize database
	err := InitDatabase()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	// Initialize Google AI client
	aiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Google AI client: %v", err))
	}

	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start the background task checker
	go TaskChecker(bot)

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		// Get or create user in database
		var username, languageCode *string
		if update.Message.From.UserName != "" {
			username = &update.Message.From.UserName
		}
		if update.Message.From.LanguageCode != "" {
			languageCode = &update.Message.From.LanguageCode
		}

		user, err := GetOrCreateUser(
			int64(update.Message.From.ID),
			username,
			&update.Message.From.FirstName,
			&update.Message.From.LastName,
			languageCode,
		)
		if err != nil {
			log.Printf("Error handling user: %v", err)
			continue
		}

		// Handle different types of messages
		var responseText string

		if update.Message.Text != "" {
			// Check for special commands
			text := strings.TrimSpace(update.Message.Text)

			if strings.HasPrefix(text, "/settimezone") {
				responseText = handleSetTimezoneCommand(text, user)
			} else if strings.HasPrefix(text, "/mytasks") {
				responseText = handleMyTasksCommand(user)
			} else if strings.HasPrefix(text, "/help") {
				responseText = handleHelpCommand()
			} else if strings.ToLower(strings.TrimSpace(text)) == "done" {
				responseText = handleDoneCommand(user)
			} else {
				// // Regular text message - parse with LLM
				// ctx := context.Background()
				// payload, err := ParseReminder(ctx, aiClient, update.Message.Text, user.Timezone)
				// if err != nil {
				// 	log.Printf("Error parsing reminder: %v", err)
				// 	responseText = "Sorry, I had trouble understanding your message. Please try again."
				// } else {
				// 	// Log the parsed payload
				// 	log.Printf("Parsed reminder payload: %+v", payload)

				// 	if payload.Type == "task" {
				// 		// Create task in database
				// 		task, err := CreateTask(user.ID, payload)
				// 		if err != nil {
				// 			log.Printf("Error creating task: %v", err)
				// 			responseText = "I understood your reminder, but had trouble saving it. Please try again."
				// 		} else {
				// 			// Format the response with timezone information
				// 			formattedTime := FormatTaskDateTime(task.DueDateTime, user.Timezone)
				// 			responseText = fmt.Sprintf("%s\n\n📅 Scheduled for: %s (%s)",
				// 				payload.LLMMessage, formattedTime, user.Timezone)
				// 		}
				// 	} else {
				// 		responseText = payload.LLMMessage
				// 	}
				// }
				// Immediate generic response for tasks
				initialResponse := fmt.Sprintf("Task Scheduled: \"%s\" (processing in background...)", update.Message.Text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, initialResponse)
				msg.ReplyToMessageID = update.Message.MessageID
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending immediate response: %v", err)
				}

				// Process the reminder in a separate goroutine
				go processUserReminder(context.Background(), bot, aiClient, user, update.Message.Text, update.Message.Chat.ID)
				continue // Skip the rest of the loop for this message
			}
		} else if update.Message.Voice != nil {
			// Audio/voice message
			responseText = "🎵 I received your audio message! I can only process text messages for now."
		} else if update.Message.Audio != nil {
			// Audio file
			responseText = "🎶 I received your audio file! I can only process text messages for now."
		} else if update.Message.Photo != nil {
			// Photo message
			responseText = "📸 I received your photo! I can only process text messages for now."
		} else if update.Message.Video != nil {
			// Video message
			responseText = "🎥 I received your video! I can only process text messages for now."
		} else if update.Message.Document != nil {
			// Document message
			responseText = "📄 I received your document! I can only process text messages for now."
		} else {
			// Other message types
			responseText = "I received your message, but I can only process text messages for now."
		}

		// Create a reply message
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		msg.ReplyToMessageID = update.Message.MessageID

		// Send the message
		if _, err := bot.Send(msg); err != nil {
			// Note that panics are a bad way to handle errors. Telegram can
			// have service outages or network errors, you should retry sending
			// messages or more gracefully handle failures.
			log.Printf("Error sending message: %v", err)
		}
	}
}
