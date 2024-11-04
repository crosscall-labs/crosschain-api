package utils

type VersionResponse struct {
	Version string `json:"version"`
}

type Error struct {
	Code    uint64 `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
	Origin  string `json:"origin"`
}

type PartialHeader struct {
	TxType      string `query:"txtype"`               // for now just type1 tx and type0 (legacy)
	ChainName   string `query:"name" optional:"true"` // add later for QoL
	ChainType   string `query:"type" optional:"true"` // add later for QoL
	ChainId     string `query:"id"`
	ChainSigner string `query:"signer"`
}

type MessageHeader struct {
	TxType          string `query:"txtype"`                // for now just type1 tx and type0 (legacy)
	FromChainName   string `query:"fname" optional:"true"` // add later for QoL
	FromChainType   string `query:"ftype" optional:"true"` // add later for QoL
	FromChainId     string `query:"fid"`
	FromChainSigner string `query:"fsigner"`
	ToChainName     string `query:"tname" optional:"true"` // add later for QoL
	ToChainType     string `query:"ttype" optional:"true"` // add later for QoL
	ToChainId       string `query:"tid"`
	ToChainSigner   string `query:"tsigner"`
}

type ChainInfo struct {
	ID             string
	VM             string
	Name           string
	EscrowType     []int
	EntrypointType []int
	Error          error
}
