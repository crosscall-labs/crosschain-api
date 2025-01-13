package tvmUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/crosscall-labs/crosschain-api/pkg/utils"
)

func CallViewFunction(api string, contractAddress string, method string, args []string) ([]byte, error) {
	baseURL := fmt.Sprintf("%s%s/methods/%s", api, contractAddress, method)
	query := url.Values{}
	for _, arg := range args {
		query.Add("args", arg)
	}
	fullURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())
	resp, err := http.Get(fullURL)

	utils.LogInfo("Viewcall URL", fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("received status: %d", resp.StatusCode)
	}
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return responseBytes, nil
}

func ParseViewResponse(decodedBytes []byte) ([]StackItem, error) {
	type RawResponse struct {
		Success  bool        `json:"success"`
		ExitCode int         `json:"exit_code"`
		Stack    []StackItem `json:"stack"`
	}

	var rawResponse RawResponse
	if err := json.Unmarshal(decodedBytes, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if !rawResponse.Success {
		return nil, fmt.Errorf("response indicates failure: exit_code %d", rawResponse.ExitCode)
	}

	return rawResponse.Stack, nil
}

func FormatKeyValueLogs(data []StackItem) string {
	var builder strings.Builder
	builder.Grow(len(data) * 20) // Adjust the growth to account for larger data.

	for _, entry := range data {
		builder.WriteString(fmt.Sprintf("  Type: %s", entry.Type))
		if entry.Num != "" {
			builder.WriteString(fmt.Sprintf(", Num: %s", entry.Num))
		}
		if entry.Cell != "" {
			builder.WriteString(fmt.Sprintf(", Cell: %s", entry.Cell))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
