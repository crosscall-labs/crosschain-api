package tonx

// method: ton_detectAddress
type TonDetectAddress struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getAddressBalance
type TonGetAddressBalance struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getAddressInformation
type TonGetAddressInformation struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getAddressState
type TonGetAddressState struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getExtendedAddressInformation
type TonGetExtendedAddressInformation struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getTokenData
type TonGetTokenData struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_getWalletInformation
type TonGetWalletInformation struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
}

// method: ton_packAddress
type TonPackAddress struct {
	Address string `json:"address"` // Required: Identifier of target TON account in raw form.
}

// method: ton_unpackAddress
type TonUnpackAddress struct {
	Address string `json:"address"` // Required: Identifier of target TON account in user-friendly form.
}

// method: ton_getBlockHeader
type TonGetBlockHeader struct {
	Workchain int    `json:"workchain"` // Workchain id
	Shard     string `json:"shard"`     // Required
	Seqno     int    `json:"seqno"`     // Required: Block sequence number
	RootHash  string `json:"root_hash"`
	FileHash  string `json:"file_hash"`
}

// method: ton_getConsensusBlock
// empty method

// method: ton_getMasterchainBlockSignatures
type TonGetMasterchainBlockSignatures struct {
	Seqno int `json:"seqno"` // Required
}

// method: ton_getMasterchainInfo
// empty method

// method: ton_getShardBlockProof
type TonGetShardBlockProof struct {
	Workchain int    `json:"workchain"` // Required
	Shard     string `json:"shard"`     // Required
	Seqno     int    `json:"seqno"`     // Required
	FromSeqno int    `json:"from_seqno"`
}

// method: ton_lookupBlock
// not found

// method: ton_shards
type TonShards struct {
	Seqno int `json:"seqno"` // Required: Block's height
}

// method: ton_getConfigParam
type TonGetConfigParam struct {
	ConfigId int `json:"config_id"` // Required: config id
	Seqno    int `json:"seqno"`     // Block sequence number
}

// method: ton_runGetMethod
type TonRunGetMethod struct {
	Address string   `json:"address"` // Required: Identifier of target TON account in raw form.
	Method  string   `json:"method"`  // Required
	Stack   []string `json:"stack"`   // Required: Stack of execution options.
}

// method: ton_estimateFee
type TonEstimateFee struct {
	Address string `json:"address"` // Required: Identifier of target TON account in any form.
	Body    string `json:"body"`    // Required: Base64 encoded message body.
}

// method: ton_sendBoc
type TonSendBoc struct {
	Boc string `json:"boc"` // Required: Base64 encoded message BOC. Encoded with base64.
}

// method: ton_sendBocReturnHash
type TonSendBocReturnHash struct {
	Boc string `json:"boc"` // Required: Base64 encoded message BOC. Encoded with base64.
}

// method: ton_getBlockTransactions
type TonGetBlockTransactions struct {
	Workchain int    `json:"workchain"` // Required: Workchain id
	Shard     string `json:"shard"`     // Required: shard
	Seqno     int    `json:"seqno"`     // Required: Block sequence number
	RootHash  string `json:"root_hash"`
	FileHash  string `json:"file_hash"`
	AfterLt   string `json:"after_lt"`
	AfterHash string `json:"after_hash"`
}

// method: ton_getTransactions
type TonGetTransactions struct {
	Address  string `json:"address"` // Required: Identifier of target TON account in any form.
	Limit    int    `json:"limit"`   // Maximum number of transactions in response.
	Lt       int    `json:"lt"`      // Logical time of transaction to start with, must be sent with hash.
	Hash     string `json:"hash"`    // Hash of transaction to start with, in base64 or hex encoding , must be sent with lt.
	ToLt     int    `json:"to_lt"`   // Logical time of transaction to finish with (to get tx from lt to to_lt).
	Archival bool   `json:"archival"`
}

// method: ton_tryLocateResultTx
// API docs missing definition
type TonTryLocateResultTx struct {
	Source      string `json:"source"`      // Required
	Destination string `json:"destination"` // Required
	CreatedLt   int    `json:"created_lt"`  // Required
}

// method: ton_tryLocateSourceTx
type TonTryLocateSourceTx struct {
	Source      string `json:"source"`      // Required
	Destination string `json:"destination"` // Required
	CreatedLt   int    `json:"created_lt"`  // Required
}

// method: ton_tryLocateTx
type TonTryLocateTx struct {
	Source      string `json:"source"`      // Required
	Destination string `json:"destination"` // Required
	CreatedLt   int    `json:"created_lt"`  // Required
}

// method: runGetMethod
type RunGetMethod struct {
	Address string `json:"address"` // Required
	Method  string `json:"method"`  // Required
}

// method: getMasterchainInfo
type GetMasterchainInfo struct {
}
