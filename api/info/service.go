package infoHandler

import (
	"fmt"
	"net/http"

	evmHandler "github.com/laminafinance/crosschain-api/api/evm"
	tvmHandler "github.com/laminafinance/crosschain-api/api/tvm"
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
			return nil, utils.ErrInternal(err.Error())
		}
	}

	var err error
	params.ChainId, params.VM, err = getChainType(params.ChainId)
	if err != nil {
		return nil, utils.ErrInternal(err.Error())
	}

	switch params.VM {
	case "evm":
		return evmHandler.AssetInfoRequest(r, params)
	case "tvm":
		return tvmHandler.AssetInfoRequest(r, params)
	// case "svm":
	// 	// return svmHandler.AssetInfoRequestSvm(params)
	default:
		return nil, utils.ErrInternal(fmt.Errorf("Virtual machine %v is unsupported", params.VM).Error())
	}
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
