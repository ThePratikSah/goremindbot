# GoRemindBot

A Telegram bot that helps you create and manage reminders using natural language processing and timezone support.

## Features

- 🤖 **Natural Language Processing**: Create reminders using natural language
- 🌍 **Timezone Support**: Automatic timezone handling for global users
- 💾 **SQLite Database**: Persistent storage with GORM
- 📱 **Telegram Integration**: Full Telegram Bot API support
- 🔄 **Recurring Reminders**: Support for recurring tasks
- 📋 **Task Management**: View and manage your active tasks

## Setup

1. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

2. **Set Environment Variables**:
   ```bash
   export TELEGRAM_APITOKEN="your_telegram_bot_token"
   export GEMINI_API_KEY="your_google_ai_api_key"
   ```

3. **Run the Bot**:
   ```bash
   go run .
   ```

## Usage

### Creating Reminders
Just send natural language messages to the bot:
- "Remind me to buy groceries tomorrow at 2 PM"
- "Call mom on Friday at 6 PM"
- "Submit the report by 5 PM today"
- "Take medicine every day at 9 AM"

### Commands
- `/help` - Show help message
- `/mytasks` - View your active tasks
- `/settimezone <timezone>` - Set your timezone (e.g., `/settimezone Asia/Kolkata`)

### Supported Timezones
- UTC
- America/New_York (EST/EDT)
- America/Chicago (CST/CDT)
- America/Denver (MST/MDT)
- America/Los_Angeles (PST/PDT)
- Europe/London (GMT/BST)
- Europe/Paris (CET/CEST)
- Europe/Berlin (CET/CEST)
- Asia/Tokyo (JST)
- Asia/Shanghai (CST)
- Asia/Kolkata (IST)
- Asia/Dubai (GST)
- Australia/Sydney (AEST/AEDT)
- Pacific/Auckland (NZST/NZDT)

## Database Schema

### Users Table
- `id`: Primary key
- `telegram_id`: Unique Telegram user ID
- `username`: Telegram username
- `first_name`: User's first name
- `last_name`: User's last name
- `language_code`: User's language preference
- `timezone`: User's timezone (default: Asia/Kolkata)
- `is_active`: Whether the user is active
- `created_at`, `updated_at`, `deleted_at`: Timestamps

### Tasks Table
- `id`: Primary key
- `user_id`: Foreign key to users table
- `title`: Task title
- `description`: Task description
- `due_date_time`: Due date/time (stored in UTC)
- `timezone`: User's timezone for display
- `recurrence`: Recurrence pattern (if any)
- `source_text`: Original user message
- `status`: Task status (pending, completed, cancelled)
- `is_active`: Whether the task is active
- `created_at`, `updated_at`, `deleted_at`: Timestamps

## Architecture

- **main.go**: Main application entry point and Telegram message handling
- **models.go**: Database models and data structures
- **database.go**: Database operations and GORM setup
- **llm.go**: Google AI integration for natural language processing
- **timezone.go**: Timezone handling and conversion utilities
- **commands.go**: Bot command handlers
- **helpers.go**: Utility functions

## Timezone Handling

The bot automatically handles timezone conversions:
1. New users default to Asia/Kolkata timezone
2. Users can set their timezone using `/settimezone` command
3. All task times are stored in UTC in the database
4. Times are displayed to users in their local timezone
5. The LLM processes reminders in the user's timezone context

This ensures that users in different timezones can use the bot without time conflicts, with India as the default timezone for new users.
