package requestHandler

import "github.com/laminafinance/crosschain-api/pkg/utils"

type MessageResponse interface {
	GetType() string
}

type UnsignedDataResponse struct {
	Header      utils.MessageHeader `json:"header"`
	FromMessage MessageResponse     `json:"from"`
	ToMessage   MessageResponse     `json:"to"`
}
