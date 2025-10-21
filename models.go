package main

import (
	"time"

	"gorm.io/gorm"
)

// ReminderPayload represents the parsed reminder from LLM
type ReminderPayload struct {
	Type        string  `json:"type"` // "task" or "not_task"
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Datetime    string  `json:"datetime,omitempty"` // ISO 8601
	Timezone    string  `json:"timezone,omitempty"`
	Recurrence  *string `json:"recurrence,omitempty"` // null if not recurring
	SourceText  string  `json:"source_text"`
	LLMMessage  string  `json:"llm_message,omitempty"` // personal touch message from LLM
}

// User represents a Telegram user in the database
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TelegramID   int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username     *string        `gorm:"index" json:"username,omitempty"`
	FirstName    *string        `json:"first_name,omitempty"`
	LastName     *string        `json:"last_name,omitempty"`
	LanguageCode *string        `json:"language_code,omitempty"`
	Timezone     string         `gorm:"default:'Asia/Kolkata'" json:"timezone"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Tasks []Task `gorm:"foreignKey:UserID" json:"tasks,omitempty"`
}

// Task represents a reminder/task in the database
type Task struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `json:"description"`
	DueDateTime    time.Time      `gorm:"not null;index" json:"due_date_time"`
	Timezone       string         `gorm:"not null" json:"timezone"`
	Recurrence     *string        `json:"recurrence,omitempty"`            // null if not recurring
	SourceText     string         `json:"source_text"`                     // original message from user
	Status         string         `gorm:"default:'pending'" json:"status"` // pending, completed, cancelled
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	ReminderSentAt *time.Time     `json:"reminder_sent_at,omitempty"` // when reminder was sent
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
