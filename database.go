package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the SQLite database connection
func InitDatabase() error {
	var err error

	// Configure GORM logger
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open("goremindbot.db"), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = DB.AutoMigrate(&User{}, &Task{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetOrCreateUser retrieves an existing user or creates a new one
func GetOrCreateUser(telegramID int64, username, firstName, lastName, languageCode *string) (*User, error) {
	var user User

	// Try to find existing user
	result := DB.Where("telegram_id = ?", telegramID).First(&user)
	if result.Error == nil {
		// User exists, update their information if needed
		updateFields := map[string]interface{}{}

		if username != nil && user.Username != username {
			updateFields["username"] = username
		}
		if firstName != nil && user.FirstName != firstName {
			updateFields["first_name"] = firstName
		}
		if lastName != nil && user.LastName != lastName {
			updateFields["last_name"] = lastName
		}
		if languageCode != nil && user.LanguageCode != languageCode {
			updateFields["language_code"] = languageCode
		}

		if len(updateFields) > 0 {
			updateFields["updated_at"] = time.Now()
			DB.Model(&user).Updates(updateFields)
		}

		return &user, nil
	}

	// User doesn't exist, create new one
	user = User{
		TelegramID:   telegramID,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		LanguageCode: languageCode,
		Timezone:     "Asia/Kolkata", // Default timezone
		IsActive:     true,
	}

	result = DB.Create(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}

	log.Printf("Created new user: %d", telegramID)
	return &user, nil
}

// CreateTask creates a new task for a user
func CreateTask(userID uint, payload *ReminderPayload) (*Task, error) {
	// Parse the datetime string and convert from user's timezone to UTC
	dueDateTime, err := ParseTaskDateTime(payload.Datetime, payload.Timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %v", err)
	}

	task := Task{
		UserID:      userID,
		Title:       payload.Title,
		Description: payload.Description,
		DueDateTime: dueDateTime,      // Store in UTC
		Timezone:    payload.Timezone, // Store user's timezone for display
		Recurrence:  payload.Recurrence,
		SourceText:  payload.SourceText,
		Status:      "pending",
		IsActive:    true,
	}

	result := DB.Create(&task)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create task: %v", result.Error)
	}

	log.Printf("Created new task: %s for user %d (UTC: %s)", task.Title, userID, dueDateTime.Format("2006-01-02 15:04:05"))
	return &task, nil
}

// GetUserTasks retrieves all active tasks for a user
func GetUserTasks(userID uint) ([]Task, error) {
	var tasks []Task
	result := DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user tasks: %v", result.Error)
	}
	return tasks, nil
}

// UpdateUserTimezone updates the user's timezone
func UpdateUserTimezone(userID uint, timezone string) error {
	result := DB.Model(&User{}).Where("id = ?", userID).Update("timezone", timezone)
	if result.Error != nil {
		return fmt.Errorf("failed to update user timezone: %v", result.Error)
	}
	return nil
}

// GetTasksDueNow retrieves all tasks that are due within the current minute and haven't sent a reminder
func GetTasksDueNow() ([]Task, error) {
	var tasks []Task
	now := time.Now().UTC()

	// Find tasks scheduled for the current minute (e.g., if now is 2:31:05, look for tasks at 2:31:00)
	startOfCurrentMinute := now.Truncate(time.Minute)
	endOfCurrentMinute := startOfCurrentMinute.Add(time.Minute) // Excludes end second for simplicity

	result := DB.Preload("User").Where(
		"due_date_time >= ? AND due_date_time < ? AND status = ? AND is_active = ? AND reminder_sent_at IS NULL",
		startOfCurrentMinute, endOfCurrentMinute, "pending", true,
	).Find(&tasks)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks due now: %v", result.Error)
	}

	return tasks, nil
}

// MarkTaskAsCompleted marks a task as completed
func MarkTaskAsCompleted(taskID uint) error {
	result := DB.Model(&Task{}).Where("id = ?", taskID).Update("status", "completed")
	if result.Error != nil {
		return fmt.Errorf("failed to mark task as completed: %v", result.Error)
	}
	return nil
}

// MarkTaskReminderSent marks a task as having its reminder sent
func MarkTaskReminderSent(taskID uint) error {
	now := time.Now().UTC()
	result := DB.Model(&Task{}).Where("id = ?", taskID).Update("reminder_sent_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to mark task reminder as sent: %v", result.Error)
	}
	return nil
}
