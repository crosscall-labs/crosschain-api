package handler

import "github.com/laminafinance/crosschain-api/pkg/utils"

type UnsignedRequestParams struct {
	Header  utils.MessageHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}

type UnsignedBytecodeParams struct {
	MessageType  string `query:"msg-type" optional:"true"`
	Signer       string `query:"signer"`
	TargetId     string `query:"destination-id"`
	OriginId     string `query:"origin-id"`
	AssetAmount  string `query:"asset-amount"`
	AssetAddress string `query:"asset-address"`
	Calldata     string `query:"calldata" optional:"true"`
}

type SignedBytecodeParams struct {
	Signer                   string `query:"signer"`
	DestinationId            string `query:"destination-id"`
	OriginId                 string `query:"origin-id"`
	AssetAddress             string `query:"asset-address"`
	AssetAmount              string `query:"asset-amount"`
	UseropSender             string `query:"useropSender"`
	UseropNonce              string `query:"useropNonce"`
	UseropInitCode           string `query:"useropInitCode" optional:"true"`
	UseropCallData           string `query:"useropCallData"`
	UseropAccountGasLimit    string `query:"useropAccountGasLimit"`
	UseropPreVerificationGas string `query:"useropPreVerificationGas"`
	UseropGasFees            string `query:"useropGasFees"`
	UseropPaymasterAndData   string `query:"useropPaymasterAndData"`
	UseropSignature          string `query:"useropSignature"`
}

type SignedEscrowPayoutParams struct {
	Bytecode string `query:"data"`
	TraceId  string `query:"traceid"`
}

type UnsignedBytecodeResponse struct {
	MessageType  string `query:"msg-type" optional:"true"`
	Signer       string `query:"signer"`
	TargetId     string `query:"destination-id"`
	OriginId     string `query:"origin-id"`
	AssetAmount  string `query:"asset-amount"`
	AssetAddress string `query:"asset-address"`
	Calldata     string `query:"calldata" optional:"true"`
}

type MessageResponse interface {
	GetType() string
}

// need to cover tvm<>evm evm<>evm
// we will change this to be able to make anj on the fly suggestion and let the user edit the message values
type UnsignedDataResponse struct {
	Header      utils.MessageHeader `json:"header"`
	FromMessage MessageResponse     `json:"from"`
	ToMessage   MessageResponse     `json:"to"`
}

// type MessageEscrowEvm struct {
// 	EscrowAddress   string `json:"eaddress"`
// 	EscrowInit      string `json:"einit"`
// 	EscrowPayload   string `json:"epayload"`
// 	EscrowAsset     string `json:"easset"`
// 	EscrowAmount    string `json:"eamount"`
// 	EscrowValueType string `json:"evaluetype"`
// 	EscrowValue     string `json:"evalue"`
// }

// func (m MessageEscrowEvm) GetType() string {
// 	return "EVM Escrow"
// }

// type MessageOpEvm struct {
// 	UserOp           PackedUserOperationResponse `json:"op-packed-data"` // parsed data, recommended to validate data
// 	PaymasterAndData PaymasterAndData            `json:"op-paymaster"`
// 	UserOpHash       string                      `json:"op-hash"`
// 	PriceGwei        string                      `json:"op-price"`
// }

// func (m MessageOpEvm) GetType() string {
// 	return "EVM UserOp"
// }

/**
if i sign the full tx the data for the escrow is signed but it means that on every preceeding chain we need to validate all the data
ergo we have type0 and type1 transaction

type0 is signed once but all the bytecode needs to be validated onchain to make the proof work

type1 is signed twice where the overall signed code is trusted on the escrow chain and the escrow payload is proved onchain instead
we can do this since we are determinstic, this design choice was to make svm/tvm integration easier in the shortterm

empty call
http://localhost:8080/api/main?query=unsigned-message&txtype=1&fid=&fsigner=&tid&tid=&tsigner=

http://localhost:8080/api/main?query=unsigned-message&txtype=1&fid=11155111&fsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&tid=1667471769&tsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&payload=00
*/
