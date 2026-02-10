package utils

import "github.com/google/uuid"

func NewUUID() string {
	if uuid, err := uuid.NewV7(); err == nil {
		return uuid.String()
	}
	panic("failed to generate UUID")
}
