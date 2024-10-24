package evmHandler

import "github.com/laminafinance/crosschain-api/pkg/utils"

type UnsignedEscrowRequestParams struct {
	Header utils.PartialHeader `query:"header"`
}

type UnsignedEntryPointRequestParams struct {
	Header  utils.PartialHeader `query:"header"`
	Payload string              `query:"payload" optional:"true"`
}
