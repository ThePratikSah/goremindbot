package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ExtractTimezoneFromMessage attempts to extract timezone information from Telegram message
// This is a simplified approach - in a real-world scenario, you might want to use
// more sophisticated methods like asking users to set their timezone or using location data
func ExtractTimezoneFromMessage(update *tgbotapi.Update) string {
	// For now, we'll use a default timezone approach
	// In the future, you could:
	// 1. Ask users to set their timezone via a command
	// 2. Use Telegram's location sharing feature
	// 3. Parse timezone from the message content
	// 4. Use a timezone detection service

	// Default to India timezone
	defaultTimezone := "Asia/Kolkata"

	// If the user has shared their location, we could potentially use that
	// to determine their timezone, but that's complex and requires additional setup

	// For now, we'll return Asia/Kolkata as default and let users set their timezone manually
	// You can implement a /settimezone command later
	return defaultTimezone
}

// ConvertToUserTimezone converts a UTC time to the user's timezone
func ConvertToUserTimezone(utcTime time.Time, userTimezone string) (time.Time, error) {
	// Load the user's timezone
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		// If timezone is invalid, fall back to UTC
		log.Printf("Invalid timezone %s, falling back to UTC: %v", userTimezone, err)
		return utcTime, nil
	}

	// Convert to user's timezone
	return utcTime.In(loc), nil
}

// ConvertFromUserTimezone converts a time from user's timezone to UTC
func ConvertFromUserTimezone(userTime time.Time, userTimezone string) (time.Time, error) {
	// Load the user's timezone
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		// If timezone is invalid, assume it's already in UTC
		log.Printf("Invalid timezone %s, assuming UTC: %v", userTimezone, err)
		return userTime, nil
	}

	// Parse the time as if it's in the user's timezone
	userTimeInTz := time.Date(
		userTime.Year(),
		userTime.Month(),
		userTime.Day(),
		userTime.Hour(),
		userTime.Minute(),
		userTime.Second(),
		userTime.Nanosecond(),
		loc,
	)

	// Convert to UTC
	return userTimeInTz.UTC(), nil
}

// ParseTaskDateTime parses a datetime string and converts it to UTC based on user's timezone
func ParseTaskDateTime(datetimeStr, userTimezone string) (time.Time, error) {
	// Parse the datetime string (assuming it's in user's timezone)
	userTime, err := time.Parse("2006-01-02T15:04:05", datetimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime: %v", err)
	}

	// Truncate to minute precision as requested
	userTime = userTime.Truncate(time.Minute)

	// Convert from user's timezone to UTC
	utcTime, err := ConvertFromUserTimezone(userTime, userTimezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to convert timezone: %v", err)
	}

	return utcTime, nil
}

// FormatTaskDateTime formats a UTC time to user's timezone for display
func FormatTaskDateTime(utcTime time.Time, userTimezone string) string {
	userTime, err := ConvertToUserTimezone(utcTime, userTimezone)
	if err != nil {
		// Fall back to UTC if conversion fails
		userTime = utcTime
	}

	// Format to exclude seconds
	return userTime.Format("2006-01-02 15:04 MST")
}

// GetCommonTimezones returns a list of common timezones for user selection
func GetCommonTimezones() []string {
	return []string{
		"UTC",
		"America/New_York",    // EST/EDT
		"America/Chicago",     // CST/CDT
		"America/Denver",      // MST/MDT
		"America/Los_Angeles", // PST/PDT
		"Europe/London",       // GMT/BST
		"Europe/Paris",        // CET/CEST
		"Europe/Berlin",       // CET/CEST
		"Asia/Tokyo",          // JST
		"Asia/Shanghai",       // CST
		"Asia/Kolkata",        // IST
		"Asia/Dubai",          // GST
		"Australia/Sydney",    // AEST/AEDT
		"Pacific/Auckland",    // NZST/NZDT
	}
}
