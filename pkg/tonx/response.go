package tonx

// response: ton_detectAddress
type TonDetectAddressResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		RawForm    string `json:"raw_form"` // Raw address in any form.
		Bounceable struct {
			B64    string `json:"b64"`
			B64URL string `json:"b64url"`
		} `json:"bounceable"`
		NonBounceable struct {
			B64    string `json:"b64"`
			B64URL string `json:"b64url"`
		} `json:"non_bounceable"`
		GivenType string `json:"given_type"`
		TestOnly  bool   `json:"test_only"`
	} `json:"result"`
}

// response: ton_getAddressBalance
type TonGetAddressBalanceResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"` // Balance of the account in nanotokens.
}

// response: ton_getConfigParamResponse
type TonGetConfigParamResponse struct {
	Jsonrpc string `json:"jsonrpc"` // Defaults to 2.0
	Id      int    `json:"id"`      // Defaults to 1
	Result  struct {
		Type   string `json:"@type"`
		Config struct {
			Type  string `json:"@type"` // config type
			Bytes string `json:"bytes"` // config bytes
		} `json:"config"`
		Extra string `json:"@extra"` // Extra Information
	} `json:"result"`
}

// response: ton_sendBocReturnHash
type TonSendBocReturnHashResponse struct {
	Jsonrpc string `json:"jsonrpc"` // Required: Defaults to 2.0
	Id      int    `json:"id"`      // Required: Defaults to 1
	Result  struct {
		Type  string `json:"@type"`  // Type of the result
		Hash  string `json:"hash"`   // Required: Transaction hash
		Extra string `json:"@extra"` // Extra Information
	} `json:"result"` // Required
}
