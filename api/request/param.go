package requestHandler

import (
	"github.com/laminafinance/crosschain-api/pkg/utils"
)

type UnsignedCrosschainRequestParams struct {
	Header  utils.MessageHeader `query:"header"`
	Target  string              `query:"target" optional:"true"`
	Value   string              `query:"value" optional:"true"`
	Payload string              `query:"payload" optional:"true"`
	//Extra   string              `query:"extra" options:"true"` // stores extra data for tvm
}
