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

type PartialHeaderResponse struct {
	TxType      string `json:"txtype"`
	ChainName   string `json:"name"`
	ChainType   string `json:"type"`
	ChainId     string `json:"id"`
	ChainSigner string `json:"signer"`
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
	IsTestnet       string `query:"testnet" optional:"true"` // default true
	ExtraData       string `query:"extra-data" optional:"true"`
}

type MessageHeaderResponse struct {
	TxType          string `json:"txtype"`
	FromChainName   string `json:"fname"`
	FromChainType   string `json:"ftype"`
	FromChainId     string `json:"fid"`
	FromChainSigner string `json:"fsigner"`
	ToChainName     string `json:"tname"`
	ToChainType     string `json:"ttype"`
	ToChainId       string `json:"tid"`
	ToChainSigner   string `json:"tsigner"`
}

type ChainInfo struct {
	ID             string
	VM             string
	Name           string
	EscrowType     []int
	EntrypointType []int
	Error          error
}

type AssetInfoRequestParams struct {
	ChainId        string `query:"chainid"`
	VM             string `query:"vm" optional:"true"`
	UserAddress    string `query:"address" optional:"true"`
	EscrowAddress  string `query:"escrow-address" optional:"true"`
	AccountAddress string `query:"account-address" optional:"true"`
	AssetAddress   string `query:"asset-address" optional:"true"`
}

type AssetInfoRequestResponse struct {
	ChainId string `json:"chain-id"`
	VM      string `json:"vm"`
	Name    string `json:"name"`
	Asset   struct {
		Address     string `json:"address"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimal     string `json:"decimal"`
		TotalSupply string `json:"total-supply"`
		Supply      string `json:"supply"`
		Description string `json:"description"`
	} `json:"asset"`
	Escrow struct {
		Init         bool   `json:"init"`
		Balance      string `json:"balance"`
		LockBalance  string `json:"lock-balance"`
		LockDeadline string `json:"lock-deadline"`
	} `json:"escrow"`
	Account struct {
		Init    bool   `json:"init"`
		Balance string `json:"balance"`
		// LockBalance string `json:"lock-balance"`
	} `json:"account"`
}
