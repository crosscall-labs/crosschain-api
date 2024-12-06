package evmHandler

import "github.com/laminafinance/crosschain-api/pkg/utils"

type UnsignedEscrowRequestParams struct {
	Header utils.PartialHeader `query:"header"`
	Amount string              `query:"amount"` // gwei
}

type UnsignedEntryPointRequestParams struct {
	Header  utils.MessageHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}
