package requestHandler

import (
	"fmt"
	"net/http"

	evmHandler "github.com/laminafinance/crosschain-api/api/evm"
	tvmHandler "github.com/laminafinance/crosschain-api/api/tvm"
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func AssetMintRequest(r *http.Request, parameters ...*utils.AssetMintRequestParams) (interface{}, error) {
	var params *utils.AssetMintRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &utils.AssetMintRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	var err error
	_, params.VM, err = utils.GetChainType(params.ChainId)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}

	switch params.VM {
	case "evm":
		return evmHandler.AssetMintRequest(r, params)
	case "tvm":
		return tvmHandler.AssetMintRequest(r, params)
	// case "svm":
	// 	// return svmHandler.AssetInfoRequestSvm(params)
	default:
		return nil, fmt.Errorf("Virtual machine %v is unsupported", params.VM)
	}
}
