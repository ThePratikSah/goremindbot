package main

import (
	"fmt"
	"strings"
	"time"
)

// handleSetTimezoneCommand handles the /settimezone command
func handleSetTimezoneCommand(text string, user *User) string {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		// Show available timezones
		timezones := GetCommonTimezones()
		response := "Please specify your timezone. Here are some common ones:\n\n"
		for i, tz := range timezones {
			if i > 0 && i%3 == 0 {
				response += "\n"
			}
			response += fmt.Sprintf("`%s` ", tz)
		}
		response += "\n\nExample: `/settimezone Asia/Kolkata`"
		return response
	}

	timezone := parts[1]

	// Validate timezone
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Sprintf("❌ Invalid timezone: %s\n\nPlease use a valid timezone like `Asia/Kolkata` or `America/New_York`", timezone)
	}

	// Update user's timezone
	err = UpdateUserTimezone(user.ID, timezone)
	if err != nil {
		return "❌ Failed to update timezone. Please try again."
	}

	// Show current time in user's timezone
	now := time.Now().UTC()
	userTime, _ := ConvertToUserTimezone(now, timezone)

	return fmt.Sprintf("✅ Timezone updated to %s\n\nCurrent time: %s",
		timezone, userTime.Format("2006-01-02 15:04:05 MST"))
}

// handleMyTasksCommand handles the /mytasks command
func handleMyTasksCommand(user *User) string {
	tasks, err := GetUserTasks(user.ID)
	if err != nil {
		return "❌ Failed to retrieve your tasks. Please try again."
	}

	if len(tasks) == 0 {
		return "📝 You don't have any active tasks yet.\n\nSend me a message like 'Remind me to buy groceries tomorrow at 2 PM' to create your first task!"
	}

	response := "📋 Your active tasks:\n\n"
	for i, task := range tasks {
		formattedTime := FormatTaskDateTime(task.DueDateTime, user.Timezone)
		status := "⏰"
		switch task.Status {
		case "completed":
			status = "✅"
		case "cancelled":
			status = "❌"
		}

		response += fmt.Sprintf("%d. %s %s - 📝 %s\n   at 📅 %s\n\n",
			i+1, status, task.Title, task.Description, formattedTime)
	}

	return response
}

// handleDoneCommand handles the "done" response to mark the most recent reminder as completed
func handleDoneCommand(user *User) string {
	// Find the most recent task that had a reminder sent but is still pending
	var task Task
	result := DB.Where(
		"user_id = ? AND status = ? AND is_active = ? AND reminder_sent_at IS NOT NULL",
		user.ID, "pending", true,
	).Order("reminder_sent_at DESC").First(&task)

	if result.Error != nil {
		return "❌ No recent reminder found to mark as done. You can use /mytasks to see your active tasks."
	}

	// Mark the task as completed
	err := MarkTaskAsCompleted(task.ID)
	if err != nil {
		return "❌ Failed to mark task as completed. Please try again."
	}

	return fmt.Sprintf("✅ Great! I've marked '%s' as completed. 🎉", task.Title)
}

// handleHelpCommand handles the /help command
func handleHelpCommand() string {
	return `🤖 **GoRemindBot Help**
		I can help you create and manage reminders! Here's how to use me:

		**Creating Reminders:**
		Just send me a message like:
		• "Remind me to buy groceries tomorrow at 2 PM"
		• "Call mom on Friday at 6 PM"
		• "Submit the report by 5 PM today"
		• "Take medicine every day at 9 AM"

		**Commands:**
		• /help - Show this help message
		• /mytasks - View your active tasks  
		• /settimezone <timezone> - Set your timezone (e.g., /settimezone Asia/Kolkata)

		**Supported Timezones:**
	` + strings.Join(GetCommonTimezones(), ", ") + `
		**Features:**
		• Natural language processing
		• Timezone support
		• Recurring reminders
		• Task management
		• Reply "done" to mark reminders as completed

		Just start chatting with me naturally! 🚀
	`
}
