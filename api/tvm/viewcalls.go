package tvmHandler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GetUserJettonWalletResponse struct {
	JettonWalletAddress string `json:"jetton_wallet_address"`
}

type GetWalletDataResponse struct {
	Balance          string `json:"balance"`
	Owner            string `json:"owner"`
	Jetton           string `json:"jetton"`
	JettonWalletCode string `json:"jetton_wallet_address"`
}

type GetWalletInfoResponse struct {
	Nonce             string `json:"nonce"`
	EntrypointAddress string `json:"entrypoint-address"`
	EvmAddress        string `json:"evm-address"`
	TvmAddress        string `json:"tvm-address"`
}

// RawResponse represents the raw JSON structure of the API response.
type RawResponse struct {
	Success  bool        `json:"success"`
	ExitCode int         `json:"exit_code"`
	Stack    []StackItem `json:"stack"`
}

// StackItem represents individual items in the stack array with the custom format.
type StackItem struct {
	Type string `json:"type"`
	Num  string `json:"num,omitempty"`
	Cell string `json:"cell,omitempty"`
}

// (int, slice, int, slice) get_wallet_info() method_id {
// 	load_data();
// 	return (storage::nonce, storage::entrypoint_address, storage::owner_evm_address, storage::owner_ton_address);
// }

// using canonical naming convention
type GetJettonDataResponse struct {
	TotalSupply      string `json:"total_supply"`
	Mintable         string `json:"mintable"`
	AdminAddress     string `json:"admin_address"`
	JettonContent    string `json:"jetton_content"`
	JettonWalletCode string `json:"jetton_wallet_code"`
}

func CallViewFunction(api string, contractAddress string, method string, args []string) ([]byte, error) {
	baseURL := fmt.Sprintf("%s%s/methods/%s", api, contractAddress, method)
	query := url.Values{}
	for _, arg := range args {
		query.Add("args", arg)
	}
	fullURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())
	resp, err := http.Get(fullURL)
	fmt.Printf("\nfullurl: %v", fullURL)
	if err != nil {
		fmt.Print(fmt.Errorf("failed to make GET request: %v", err).Error())
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

// func ParseViewResponse(response interface{}, targetStruct interface{}) (interface{}, error) {
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal response map: %v", err)
// 	}

// 	if err := json.Unmarshal(decodedBytes, targetStruct); err != nil {
// 		return nil, fmt.Errorf("parsing failed: %v", err)
// 	}

// 	return targetStruct, nil
// }

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

func getUserJettonWallet(userAddressRaw string, assetAddressRaw string) (GetUserJettonWalletResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, assetAddressRaw, "get_wallet_address", []string{userAddressRaw})
	if err != nil {
		return GetUserJettonWalletResponse{}, err
	}

	stack, err := ParseViewResponse(response)
	if err != nil {
		return GetUserJettonWalletResponse{}, err
	}

	if len(stack) < 1 {
		return GetUserJettonWalletResponse{}, fmt.Errorf("stack has insufficient items to map to GetWalletInfoResponse")
	}

	result := GetUserJettonWalletResponse{
		JettonWalletAddress: stack[0].Cell,
	}

	fmt.Printf("\nformatted get_wallet_address response: %+v", result)
	return result, nil
}

func getJettonData(jettonAddressRaw string) (GetJettonDataResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, jettonAddressRaw, "get_jetton_data", []string{})
	if err != nil {
		return GetJettonDataResponse{}, err
	}

	stack, err := ParseViewResponse(response)
	if err != nil {
		return GetJettonDataResponse{}, err
	}

	if len(stack) < 5 {
		return GetJettonDataResponse{}, fmt.Errorf("stack has insufficient items to map to GetWalletInfoResponse")
	}

	result := GetJettonDataResponse{
		TotalSupply:      stack[0].Num,
		Mintable:         stack[1].Num,
		AdminAddress:     stack[2].Cell,
		JettonContent:    stack[3].Cell,
		JettonWalletCode: stack[4].Cell,
	}

	fmt.Printf("\nformatted get_jetton_data response: %+v", result)
	return result, nil
}

func getWalletData(userJettonWalletRaw string) (GetWalletDataResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, userJettonWalletRaw, "get_wallet_data", []string{})
	if err != nil {
		return GetWalletDataResponse{}, err
	}

	stack, err := ParseViewResponse(response)
	if err != nil {
		return GetWalletDataResponse{}, err
	}

	if len(stack) < 4 {
		return GetWalletDataResponse{}, fmt.Errorf("stack has insufficient items to map to GetWalletInfoResponse")
	}

	result := GetWalletDataResponse{
		Balance:          stack[0].Num,
		Owner:            stack[1].Cell,
		Jetton:           stack[2].Num,
		JettonWalletCode: stack[3].Cell,
	}

	fmt.Printf("\nformatted get_wallet_data response: %+v", result)
	return result, nil
}

func getWalletInfo(proxyWalletAddressRaw string) (GetWalletInfoResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, proxyWalletAddressRaw, "get_wallet_info", []string{})
	if err != nil {
		return GetWalletInfoResponse{}, err
	}

	stack, err := ParseViewResponse(response)
	if err != nil {
		return GetWalletInfoResponse{}, err
	}

	if len(stack) < 4 {
		return GetWalletInfoResponse{}, fmt.Errorf("stack has insufficient items to map to GetWalletInfoResponse")
	}

	result := GetWalletInfoResponse{
		Nonce:             stack[0].Num,
		EntrypointAddress: stack[1].Cell,
		EvmAddress:        stack[2].Num,
		TvmAddress:        stack[3].Cell,
	}

	fmt.Printf("\nformatted get_wallet_info response: %+v", result)
	return result, nil
}
