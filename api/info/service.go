package infoHandler

import (
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

func AssetInfoRequest(r *http.Request, parameters ...*interface{}) (interface{}, error) {
	// var params *AssetInfoRequestParams

	// if len(parameters) > 0 {
	// 	params = parameters[0]
	// } else {
	// 	params = &AssetInfoRequestParams{}
	// }

	// if r != nil {
	// 	if err := utils.ParseAndValidateParams(r, &params); err != nil {
	// 		return nil, err
	// 	}
	// }

	// chainInfo, err := utils.GetChainInfo(params.ChainId)
	// if err != nil {
	// 	return nil, err
	// }

	// switch chainInfo.VM {
	// case "evm":
	// 	// return evmHandler.AssetInfoRequestEvm(params)
	// case "tvm":
	// 	// return tvmHandler.AssetInfoRequestTvm(params)
	// case "svm":
	// 	// return svmHandler.AssetInfoRequestSvm(params)
	// default:
	// 	return nil, fmt.Errorf("Virtual machine %v is unsupported", chainInfo.VM)
	// }
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
