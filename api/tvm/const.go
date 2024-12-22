package tvmHandler

import (
	"encoding/hex"

	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var escrowCodeHex = "b5ee9c724102060100011e000114ff00f4a413f4bcf2c80b01020162020502f8d06c2220c700915be001d0d3030171b0915be0fa403001d31fed44d0fa4001f861fa4001f862d39f01f863d3ff01f864d401d0f866d430d0fa0030f86521c01ee30221c01f8e2031f84212c705f2e192708018c8cb0502fa403012cf1621fa02cb6ac98306fb00e030312082103b307c4bba9130e0208210622117e90304007631f84112c705f2e191d430d0f866f846d749810208baf2e1f5c8f845fa02c9c8f846cf16c9f844f843c8f841cf16f842cf16cb9fcbffccccc9ed540024ba9130e020c01e9130e082100b7ba46fbadc0071a04e37da89a1f48003f0c3f48003f0c5a73e03f0c7a7fe03f0c9a803a1f0cda861a1f40061f0cbf08da60ff0cdf08da7fff0cdf08da7fff0cd5649929b"
var proxyWalletCodeHex = "b5ee9c7241010a01008b000114ff00f4a413f4bcf2c80b0102016202070202ce03060201200405006b1b088831c02456f8007434c0cc1c6c244c383c0074c7f4cfcc74c7cc3c008060841fa1d93beea63e1080683e18bc00b80c2103fcbc20001d3b513434c7c07e1874c7c07e18b46000194f842f841c8cb1fcb1fc9ed54802016e0809000db5473e003f0830000db63ffe003f08500171db07"

var escrowCodeBytes, _ = utils.HexToBytes(escrowCodeHex)
var proxyWalletCodeBytes, _ = hex.DecodeString(proxyWalletCodeHex)

var entryPointAddress, _ = address.ParseAddr("kQAGJK50PW_a1ZbQWK0yldegu56FlX0nXKQIa7xzoWCzQiV2")

var TestnetInfo = &tlb.BlockInfo{
	Workchain: -1,
	Shard:     -9223372036854775808,
	SeqNo:     25632053,
	RootHash: []byte{
		0xb7, 0x00, 0x31, 0xa8, 0x66, 0x6c, 0x86, 0x73,
		0x26, 0x86, 0xff, 0x9d, 0x50, 0xfd, 0xcc, 0xf9,
		0x9f, 0xf5, 0x97, 0x0f, 0xf9, 0x03, 0xbc, 0xeb,
		0x29, 0x4d, 0x89, 0x51, 0xaf, 0xef, 0xa9, 0x64,
	},
	FileHash: []byte{
		0xc7, 0x2a, 0x50, 0x97, 0xf5, 0x57, 0xfe, 0xf2,
		0xe7, 0xbb, 0x70, 0x1b, 0x63, 0x49, 0xbe, 0xfe,
		0x91, 0x85, 0x85, 0xd8, 0xb8, 0x79, 0xa9, 0x60,
		0xb8, 0xd7, 0xbf, 0xd9, 0x9b, 0xdf, 0x05, 0xa3,
	},
}

var MainnetInfo = &tlb.BlockInfo{
	Workchain: -1,
	Shard:     -9223372036854775808,
	SeqNo:     42608517,
	RootHash: []byte{
		0x37, 0x8b, 0xb6, 0x1c, 0x56, 0x66, 0xff, 0x0f,
		0x01, 0x36, 0xf4, 0xa2, 0x7f, 0x02, 0x7b, 0x7c,
		0x88, 0x47, 0xc6, 0xab, 0xa7, 0x3b, 0xee, 0xa9,
		0xad, 0x35, 0x74, 0x99, 0x38, 0xe9, 0x12, 0xc3,
	},
	FileHash: []byte{
		0x88, 0x06, 0x7c, 0x20, 0x22, 0x76, 0xac, 0x7b,
		0xd3, 0x81, 0x56, 0x27, 0x3a, 0xa0, 0xb3, 0x52,
		0x85, 0x4f, 0x64, 0x3b, 0x0f, 0x43, 0x20, 0xb1,
		0xd8, 0xd2, 0x2c, 0x54, 0x13, 0x2d, 0xd5, 0x18,
	},
}

// var faucetAddressMap = map[string]string{
// 	"0x3106A":  "https://testnet-rpc.bitlayer.org",
// 	"200810":   "https://testnet-rpc.bitlayer.org",
// 	"0x4268":   "https://ethereum-holesky-rpc.publicnode.com",
// 	"17000":    "https://ethereum-holesky-rpc.publicnode.com",
// 	"0xAA36A7": "https://ethereum-sepolia.publicnode.com",
// 	"11155111": "https://ethereum-sepolia.publicnode.com",
// 	"0xF35A":   "https://rpc.devnet.citrea.xyz",
// 	"62298":    "https://rpc.devnet.citrea.xyz",
// 	"998":      "0xE646A260699beB8cAcda436b2F96B1EdCBe88291",
// }

// func getMulticallAddress(chainId string) (common.Address, error) {
// 	if multicallAddress, found := multicallAddressMap[chainId]; found {
// 		return common.HexToAddress(multicallAddress), nil
// 	}
// 	return common.Address{}, fmt.Errorf("multicall address could not be found for %v", chainId)
// }
