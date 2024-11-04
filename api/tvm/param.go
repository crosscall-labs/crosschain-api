package tvmHandler

type ProxyInitParams struct {
	Nonce           string `query:"nonce" optional:"true"`
	EntryPoint      string `query:"entrypoint" optional:"true"`
	PayeeAddress    string `query:"payee-address" optional:"true"`
	OwnerEvmAddress string `query:"evm-address"`
	OwnerTvmAddress string `query:"tvm-address" optional:"true"`
}

type EscrowLockParams struct {
	SignerAddress string `query:"signer-address"`
	AdminAddress  string `query:"admin-address" optional:"true"`
	PayeeAddress  string `query:"payee-address" optional:"true"`
	Id            string `query:"id" optional:"true"`
	Value         string `query:"value" optional:"true"`
}

/*
// because of how the message headers operate it means we need to store in our db the users:
	 ton address, evm address, and escrow address for any chains (this will speed up development but up-cost)
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
*/
