package infoHandler

import "fmt"

var chainTypeMap = map[string]struct {
	ChainId string
	VM      string
}{
	"0x3106A":    {ChainId: "200810", VM: "evm"},
	"200810":     {ChainId: "200810", VM: "evm"},
	"0x4268":     {ChainId: "17000", VM: "evm"},
	"17000":      {ChainId: "17000", VM: "evm"},
	"0xAA36A7":   {ChainId: "11155111", VM: "evm"},
	"11155111":   {ChainId: "11155111", VM: "evm"},
	"0xF35A":     {ChainId: "62298", VM: "evm"},
	"62298":      {ChainId: "62298", VM: "evm"},
	"0x63639999": {ChainId: "1667471769", VM: "tvm"},
	"1667471769": {ChainId: "1667471769", VM: "tvm"},
	"998":        {ChainId: "998", VM: "evm"},
}

func getChainType(chainId string) (string, string, error) {
	if data, found := chainTypeMap[chainId]; found {
		return data.ChainId, data.VM, nil
	}

	return "", "", fmt.Errorf("unsupporting chain id: %s", chainId)
}
