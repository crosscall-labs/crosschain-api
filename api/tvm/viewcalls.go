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

type GetJettonDataResponse struct {
	TotalSupply      string `json:"total_supply"`
	Mintable         bool   `json:"mintable"`
	AdminAddress     string `json:"admin_address"`
	JettonContent    string `json:"jetton_content"`
	JettonWalletCode string `json:"jetton_wallet_code"`
}

func CallViewFunction(api string, contractAddress string, method string, args []string) (interface{}, error) {
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var result ViewFunctionResult
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("the error: %v", err)
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	return result, nil
}

func ParseViewResponse(response interface{}, targetStruct interface{}) (interface{}, error) {
	decodedBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response map: %v", err)
	}

	if err := json.Unmarshal(decodedBytes, targetStruct); err != nil {
		return nil, fmt.Errorf("parsing failed: %v", err)
	}

	return targetStruct, nil
}

func getUserJettonWallet(userAddressRaw string, assetAddressRaw string) (interface{}, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, assetAddressRaw, "get_wallet_address", []string{userAddressRaw})
	if err != nil {
		return nil, err
	}

	parsedResponse, err := ParseViewResponse(response.(ViewFunctionResult).Decoded, &GetUserJettonWalletResponse{})
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

func getJettonData(jettonAddressRaw string) (interface{}, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, jettonAddressRaw, "get_jetton_data", []string{})
	if err != nil {
		return nil, err
	}

	parsedResponse, err := ParseViewResponse(response.(ViewFunctionResult).Decoded, &GetJettonDataResponse{})
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

func getWalletData(userJettonWalletRaw string) (interface{}, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, userJettonWalletRaw, "get_wallet_data", []string{})
	if err != nil {
		return nil, err
	}

	parsedResponse, err := ParseViewResponse(response.(ViewFunctionResult).Decoded, &GetWalletDataResponse{})
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

func getWalletInfo(proxyWalletAddressRaw string) (interface{}, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := CallViewFunction(api, proxyWalletAddressRaw, "get_wallet_info", []string{})
	if err != nil {
		return nil, err
	}

	parsedResponse, err := ParseViewResponse(response.(ViewFunctionResult).Decoded, &GetWalletDataResponse{})
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}
