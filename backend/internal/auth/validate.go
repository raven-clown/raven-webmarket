package auth

import (
	"fmt"
	"strings"
)

func discordKey(discordID string) string {
	discordID = strings.TrimSpace(discordID)
	if strings.HasPrefix(discordID, "discord:") {
		return discordID
	}
	return "discord:" + discordID
}

func isValidSteamIdentifier(identifier string) bool {
	id := strings.ToLower(strings.TrimSpace(identifier))
	if id == "" {
		return false
	}
	if strings.HasPrefix(id, "discord:") {
		return false
	}
	if !strings.HasPrefix(id, "steam:") {
		return false
	}
	hex := strings.TrimPrefix(id, "steam:")
	if len(hex) < 1 {
		return false
	}
	for _, c := range hex {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return false
	}
	return true
}

func rejectInvalidIdentifier(identifier string) error {
	id := strings.ToLower(strings.TrimSpace(identifier))
	if strings.HasPrefix(id, "discord:") {
		return fmt.Errorf("account not eligible: identifier must be steam hex, not discord id")
	}
	return fmt.Errorf("account not eligible: identifier must be steam hex format")
}

func isUnknownColumn(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown column") || strings.Contains(msg, "doesn't exist")
}
