package evmHandler

import "github.com/laminafinance/crosschain-api/pkg/utils"

type UnsignedEscrowRequestParams struct {
	Header utils.MessageHeader `query:"header"`
}

type UnsignedEntryPointRequestParams struct {
	Header  utils.MessageHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}
