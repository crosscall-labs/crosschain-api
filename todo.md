### TODO

- [x] host
- [ ] change to gRPC
- [ ] test messages
- [ ] ci/cd test (github) (kinda already done with vercel but need action tests)
- [x] modulate project
- [ ] setup compiled contract code else where to fetch, rather than update manually
- [ ] deserialization does not match es16 decoding
	- [x] deserialization go rewrite
	- [ ] determine why @ton/core can not deserialize it's own toBoc buffer
- [ ] generate tonutils contract compilation
- [x] deploy and increament counter contract via the backend using es16 contract compiled code
- [ ] create calls to tvm
	- [x] deploy counter
	- [x] call counter view
	- [x] call counter
	- [x] deploy + call and verify
	- [ ] tonutils-go has 1.5 sec latency, try tonx
		- [x] get method
		- [ ] send method (TBD since we can't yet convert msg to raw BoC)
		- [x] static masterchain info, saves 0.25 sec 
		- [x] masterchain tonx, required for seqno (different masterchain info, verified)
	- [ ] call upon listen
		- [x] generation event
		- [ ] add new contract to db
		- [ ] trigger listener update (edge case, what if listener is slow than block propegation)
- [ ] ws for frontend transactions
- [ ] TVM InitClient needs to be modified to input shard and workchain
- [ ] tonx fee estimation a fee estimation in general not working for tvm
- [x] tvm<>evm entrypoint messages
- [ ] tvm<>evm escrow messages
	- [x] evm>tvm tx flow
- [ ] run local evm network
- [ ] all evm selectors should be precalculated
- [ ] migrate info apis to crosschain-api
	- [x] create faucet
	- [x] migrate faucet
	- [ ] create spot
	- [ ] migrate spot
	- [ ] create user
	- [ ] migrate user
	- [x] create asset
	- [x] migrate asset
	- [x] create chain
	- [ ] migrate chain

### random chores (low priority)

- [ ] conditional tags for chain name, id, and type
- [ ] create diagrams for how escrows currently work
- [ ] create diagrams for how escrow should work
  - [ ] for now we will receive funds and delegate rewards after our backend reveives them (workaround)
- [ ] need to add documentation to tonx-go api
- [ ] inquire TonX team as to the missing docs for ton_tryLocateResultTx
- [ ] finish creating TonX api response structs
- [ ] create non-must ton functions for better error handling
- [ ] convert boc serialization/deserialization offset -> reader
- [ ] run local tvm network 
- [ ] add a generic mailbox address to all chains, allows anon triggering (no owner)


type UnsignedEntryPointRequestParams struct {
	Header      utils.MessageHeader `query:"header"`
	ProxyParams ProxyParams         `query:"proxy"`
}

type ProxyParams struct {
	ProxyHeader     ProxyHeaderParams   `query:"p-header"`
	ExecutionData   ExecutionDataParams `query:"p-exe"`
	WithProxyInit   string              `query:"p-init"` // Required: Initalize the proxy wallet
	ProxyWalletCode string              `query:"p-code" optional:"true"`
	WorkChain       string              `query:"p-workchain" optional:"true"` // assume 0 for testnet atm
}

type ProxyHeaderParams struct {
	Nonce           string `query:"p-nonce" optional:"true"`
	EntryPoint      string `query:"p-entrypoint" optional:"true"` // possible that a better one is accepted in the future
	PayeeAddress    string `query:"p-payee" optional:"true"`      // solver is us for now
	OwnerEvmAddress string `query:"p-evm"`                        // easy to derive
	OwnerTvmAddress string `query:"p-tvm" optional:"true"`        // our social login SHOULD generate this
}

type ExecutionDataParams struct {
	Regime      string `query:"exe-regime" optional:"true"`
	Destination string `query:"exe-target" optional:"true"`
	Value       string `query:"exe-value" optional:"true"`
	Body        string `query:"exe-body" optional:"true"`
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

http://localhost:8080/api/tvm?query=unsigned-entrypoint-request&txtype=1&fid=12345&fsigner=1234567890&tid=67890&tsigner=1234567890&p-init=false&p-workchain=-1&p-evm=f39Fd6e51aad88F6F4ce6aB8827279cffFb92266&p-tvm=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf

http://localhost:8080/api/main?query=unsigned-message&txtype=1&fid=11155111&fsigner=0x19E7E376E7C213B7E7e7e46cc70A5dD086DAff2A&tid=1667471769&tsigner=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf&payload=00&target=UQAzC1P9oEQcVzKIOgyVeidkJlWbHGXvbNlIute5W5XHwNgf