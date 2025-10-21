package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TaskChecker runs as a background goroutine to check for due tasks and send reminders
func TaskChecker(bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("Task checker started - checking for due tasks every second")

	for range ticker.C {
		// Check for tasks that are due now
		tasks, err := GetTasksDueNow()
		if err != nil {
			log.Printf("Error getting tasks due now: %v", err)
			continue
		}

		// Send reminders for each due task
		for _, task := range tasks {
			err := sendTaskReminder(bot, &task)
			if err != nil {
				log.Printf("Error sending reminder for task %d: %v", task.ID, err)
				continue
			}

			// Mark task as reminder sent (but keep it pending so user can mark as completed)
			err = MarkTaskReminderSent(task.ID)
			if err != nil {
				log.Printf("Error marking task %d reminder as sent: %v", task.ID, err)
			} else {
				log.Printf("Sent reminder for task %d: %s", task.ID, task.Title)
			}
		}
	}
}

// sendTaskReminder sends a reminder message to the user for a specific task
func sendTaskReminder(bot *tgbotapi.BotAPI, task *Task) error {
	// Format the reminder message
	formattedTime := FormatTaskDateTime(task.DueDateTime, task.User.Timezone)

	message := fmt.Sprintf("🔔 **Reminder: %s**\n\n📝 %s\n\n⏰ Scheduled for: %s (%s)\n\n✅ Reply with 'done' to mark as completed",
		task.Title,
		task.Description,
		formattedTime,
		task.User.Timezone,
	)

	// Create the message
	msg := tgbotapi.NewMessage(int64(task.User.TelegramID), message)
	msg.ParseMode = "Markdown"

	// Send the message
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send reminder message: %v", err)
	}

	return nil
}
