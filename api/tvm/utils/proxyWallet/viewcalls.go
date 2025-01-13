package proxyWallet

import (
	"fmt"

	tvmUtils "github.com/crosscall-labs/crosschain-api/api/tvm/utils"
	"github.com/crosscall-labs/crosschain-api/pkg/utils"
)

func GetWalletInfo(proxyWalletAddressRaw string) (GetWalletInfoResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := tvmUtils.CallViewFunction(api, proxyWalletAddressRaw, "get_wallet_info", []string{})
	if err != nil {
		return GetWalletInfoResponse{}, err
	}

	stack, err := tvmUtils.ParseViewResponse(response)
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

	utils.LogInfo("Formatted get_wallet_info response", tvmUtils.FormatKeyValueLogs(stack))
	return result, nil
}
