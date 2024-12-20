package infoHandler

import (
	"fmt"
	"net/http"

	"github.com/laminafinance/crosschain-api/pkg/utils"
)

func VersionRequest(r *http.Request, parameters ...interface{}) (interface{}, error) {
	return utils.VersionResponse{
		Version: Version,
	}, nil
}

// func ChainInfoRequest(r *http.Request, parameters ...*UnsignedEscrowRequestParams) (interface{}, error) {
// 	// migrate from sdk api
// }

func AssetInfoRequest(r *http.Request, parameters ...*utils.AssetInfoRequestParams) (interface{}, error) {
	var params *utils.AssetInfoRequestParams

	if len(parameters) > 0 {
		params = parameters[0]
	} else {
		params = &utils.AssetInfoRequestParams{}
	}

	if r != nil {
		if err := utils.ParseAndValidateParams(r, &params); err != nil {
			return nil, err
		}
	}

	var err error
	params.ChainId, params.VM, err = getChainType(params.ChainId)
	if err != nil {
		return nil, err
	}

	switch params.VM {
	case "evm":
		return evmHandler.AssetInfoRequest(params)
	case "tvm":
		// return tvmHandler.AssetInfoRequest(params)
	// case "svm":
	// 	// return svmHandler.AssetInfoRequestSvm(params)
	default:
		return nil, fmt.Errorf("Virtual machine %v is unsupported", params.VM)
	}
	return nil, nil
}

func UserInfoRequest(r *http.Request, parameters ...*interface{}) (interface{}, error) {
	// var params *UnsignedEscrowRequestParams

	// if len(parameters) > 0 {
	// 	params = parameters[0]
	// } else {
	// 	params = &UnsignedEscrowRequestParams{}
	// }

	// if r != nil {
	// 	if err := utils.ParseAndValidateParams(r, &params); err != nil {
	// 		return nil, err
	// 	}
	// }

	return nil, nil
}

func UnsignedEscrowRequest(r *http.Request, parameters ...*interface{}) (interface{}, error) {
	// var params *UnsignedEscrowRequestParams

	// if len(parameters) > 0 {
	// 	params = parameters[0]
	// } else {
	// 	params = &UnsignedEscrowRequestParams{}
	// }

	// if r != nil {
	// 	if err := utils.ParseAndValidateParams(r, &params); err != nil {
	// 		return nil, err
	// 	}
	// }

	return nil, nil
}
