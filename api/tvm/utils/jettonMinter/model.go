package jettonMinter

type GetWalletAddressResponse struct {
	WalletAddress string `json:"jetton_wallet_address"`
}

// using canonical naming convention
type GetJettonDataResponse struct {
	TotalSupply      string `json:"total_supply"`
	Mintable         string `json:"mintable"`
	AdminAddress     string `json:"admin_address"`
	JettonContent    string `json:"jetton_content"`
	JettonWalletCode string `json:"jetton_wallet_code"`
}
