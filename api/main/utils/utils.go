package utils

import (
	"fmt"
)

func WriteJSONResponse(message string) string {
	return fmt.Sprintf("%s%s", "message: ", message)
}
