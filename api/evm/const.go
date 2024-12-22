package evmHandler

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

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

	return "", fmt.Errorf("unsupporting chain id: %s", chainId)
}

var multicallAddressMap = map[string]string{
	"0x3106A":  "https://testnet-rpc.bitlayer.org",
	"200810":   "https://testnet-rpc.bitlayer.org",
	"0x4268":   "https://ethereum-holesky-rpc.publicnode.com",
	"17000":    "https://ethereum-holesky-rpc.publicnode.com",
	"0xAA36A7": "https://ethereum-sepolia.publicnode.com",
	"11155111": "https://ethereum-sepolia.publicnode.com",
	"0xF35A":   "https://rpc.devnet.citrea.xyz",
	"62298":    "https://rpc.devnet.citrea.xyz",
	"998":      "0xE646A260699beB8cAcda436b2F96B1EdCBe88291",
}

func getMulticallAddress(chainId string) (common.Address, error) {
	if multicallAddress, found := multicallAddressMap[chainId]; found {
		return common.HexToAddress(multicallAddress), nil
	}
	return common.Address{}, fmt.Errorf("multicall address could not be found for %v", chainId)
}
