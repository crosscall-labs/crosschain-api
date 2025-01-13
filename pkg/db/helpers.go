package db

import (
	"log"

	"github.com/crosscall-labs/crosschain-api/pkg/utils"
	"github.com/supabase-community/supabase-go"
)

func LogError(client *supabase.Client, err error, message string, context interface{}) error {
	logData := map[string]interface{}{
		"log_level": "ERROR",
		"error":     err,
		"message":   message,
		"context":   context,
	}

	_, _, dbErr := client.From("debug_logs").Insert(logData, false, "", "minimal", "").Execute()
	if dbErr != nil {
		log.Printf("Failed to insert log: %v", dbErr)
		return dbErr
	}
	return nil
}

func LogPanic(client *supabase.Client, message string, context interface{}) error {
	panicErr := utils.Error{
		Code:    500,
		Message: "Server exited",
		Details: message,
		Origin:  utils.GetOrigin(),
	}

	if err := LogError(client, panicErr, message, context); err != nil {
		log.Printf("Failed to log panic: %v", err)
		return err
	}

	return nil
}
