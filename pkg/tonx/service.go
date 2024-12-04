package tonx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// generic request method for TonX
func SendTonXRequest(url, apiKey, jsonrpc string, id int, method string, params interface{}) (string, error) {
	requestBody := TonXRequest{
		Jsonrpc: jsonrpc,
		Id:      id,
		Method:  method,
		Params:  params,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling request data: %v", err)
	}

	fullURL := fmt.Sprintf("%s/%s", url, apiKey)
	fmt.Printf("\nrequest url: %s\n\n", fullURL)

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}
