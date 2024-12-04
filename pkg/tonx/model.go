package tonx

// need to define this all methods
type TonXRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type GetMasterchainInfoResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		First map[string]interface{} `json:"first"`
		Last  map[string]interface{} `json:"last"`
	} `json:"result"`
	ID int `json:"id"`
}
