package evmHandler

import "fmt"

var chainRpcMap = map[string]string{
	"0x3106A":  "https://testnet-rpc.bitlayer.org",
	"200810":   "https://testnet-rpc.bitlayer.org",
	"0x4268":   "https://ethereum-holesky-rpc.publicnode.com",
	"17000":    "https://ethereum-holesky-rpc.publicnode.com",
	"0xAA36A7": "https://ethereum-sepolia.publicnode.com",
	"11155111": "https://ethereum-sepolia.publicnode.com",
	"0xF35A":   "https://rpc.devnet.citrea.xyz",
	"62298":    "https://rpc.devnet.citrea.xyz",
	"998":      "https://api.hyperliquid-testnet.xyz/evm",
}

func getChainRpc(chainId string) (string, error) {
	if jsonrpc, found := chainRpcMap[chainId]; found {
		return jsonrpc, nil
	}

	return "", fmt.Errorf("Unsupporting Chain ID: %s", chainId)
}
