package jettonWallet

type GetWalletDataResponse struct {
	Balance          string `json:"balance"`
	Owner            string `json:"owner"`
	Jetton           string `json:"jetton"`
	JettonWalletCode string `json:"jetton_wallet_address"`
}
