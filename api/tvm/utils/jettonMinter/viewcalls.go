package jettonMinter

import (
	"fmt"

	tvmUtils "github.com/laminafinance/crosschain-api/api/tvm/utils"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func GetWalletAddress(userAddressRaw string, assetAddressRaw string) (GetWalletAddressResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := tvmUtils.CallViewFunction(api, assetAddressRaw, "get_wallet_address", []string{userAddressRaw})
	if err != nil {
		return GetWalletAddressResponse{}, err
	}

	stack, err := tvmUtils.ParseViewResponse(response)
	if err != nil {
		return GetWalletAddressResponse{}, err
	}

	if len(stack) < 1 {
		return GetWalletAddressResponse{}, fmt.Errorf("stack has insufficient items to map to GetWalletInfoResponse")
	}

	result := GetWalletAddressResponse{
		WalletAddress: stack[0].Cell,
	}

	utils.LogInfo("Formatted get_wallet_address response", tvmUtils.FormatKeyValueLogs(stack))
	return result, nil
}

func GetJettonData(jettonAddressRaw string) (GetJettonDataResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := tvmUtils.CallViewFunction(api, jettonAddressRaw, "get_jetton_data", []string{})
	if err != nil {
		return GetJettonDataResponse{}, err
	}

	stack, err := tvmUtils.ParseViewResponse(response)
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

	utils.LogInfo("Formatted get_jetton_data response", tvmUtils.FormatKeyValueLogs(stack))
	return result, nil
}
