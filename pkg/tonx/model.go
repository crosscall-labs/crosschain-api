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

type TonRunGetMethodResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Type     string      `json:"@type"`
		GasUsed  int         `json:"gas_used"`
		Stack    [][2]string `json:"stack"`
		ExitCode int         `json:"exit_code"`
		Extra    string      `json:"@extra"`
	} `json:"result"`
	Id int `json:"id"`
}

type TonEstimateFeeResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Type       string `json:"@type"`
		SourceFees struct {
			Type       string `json:"@type"`
			InFwdFee   int64  `json:"in_fwd_fee"`
			StorageFee int64  `json:"storage_fee"`
			GasFee     int64  `json:"gas_fee"`
			FwdFee     int64  `json:"fwd_fee"`
		} `json:"source_fees"`
		DestinationFees []struct {
			// Placeholder if fields are added later
		} `json:"destination_fees"`
		Extra string `json:"@extra"`
	} `json:"result"`
}
