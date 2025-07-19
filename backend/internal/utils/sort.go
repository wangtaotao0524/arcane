package utils

import (
	"sort"
	"strings"
	"time"

	"github.com/ofkm/arcane-backend/internal/dto"
)

// SortUserResponses sorts a slice of UserResponseDto structs by the specified column and direction
func SortUserResponses(users []dto.UserResponseDto, column, direction string) {
	sort.Slice(users, func(i, j int) bool {
		var aVal, bVal interface{}

		switch column {
		case "username":
			aVal, bVal = users[i].Username, users[j].Username
		case "displayName":
			aVal, bVal = getStringPtrValue(users[i].DisplayName), getStringPtrValue(users[j].DisplayName)
		case "email":
			aVal, bVal = getStringPtrValue(users[i].Email), getStringPtrValue(users[j].Email)
		case "roles":
			aVal, bVal = strings.Join(users[i].Roles, ","), strings.Join(users[j].Roles, ",")
		case "createdAt", "updatedAt":
			aTime, _ := time.Parse("2006-01-02T15:04:05.999999Z", users[i].CreatedAt)
			bTime, _ := time.Parse("2006-01-02T15:04:05.999999Z", users[j].CreatedAt)
			if column == "updatedAt" {
				aTime, _ = time.Parse("2006-01-02T15:04:05.999999Z", users[i].UpdatedAt)
				bTime, _ = time.Parse("2006-01-02T15:04:05.999999Z", users[j].UpdatedAt)
			}
			if direction == "desc" {
				return aTime.After(bTime)
			}
			return aTime.Before(bTime)
		default:
			return false
		}

		// Handle string comparison
		aStr := strings.ToLower(aVal.(string))
		bStr := strings.ToLower(bVal.(string))

		if direction == "desc" {
			return aStr > bStr
		}
		return aStr < bStr
	})
}

// getStringPtrValue safely gets the value from a string pointer
func getStringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
