package jettonWallet

import (
	"fmt"

	tvmUtils "github.com/crosscall-labs/crosschain-api/api/tvm/utils"
	"github.com/crosscall-labs/crosschain-api/pkg/utils"
)

func GetWalletData(userJettonWalletRaw string) (GetWalletDataResponse, error) {
	api := "https://testnet.tonapi.io/v2/blockchain/accounts/"
	response, err := tvmUtils.CallViewFunction(api, userJettonWalletRaw, "get_wallet_data", []string{})
	if err != nil {
		return GetWalletDataResponse{}, err
	}

	stack, err := tvmUtils.ParseViewResponse(response)
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

	utils.LogInfo("Formatted get_wallet_data response", tvmUtils.FormatKeyValueLogs(stack))
	return result, nil
}
