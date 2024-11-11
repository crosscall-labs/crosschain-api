package tvmHandler

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/laminafinance/crosschain-api/pkg/utils"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/sha3"

	//cell "github.com/xssnick/tonutils-go/tvm/cell"
	//tlb "github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// still need to the params struct for the inital call
// the protocol calling the client (rest etc) will be providing the actual call to be made
// the call will be in the format ProxyMessageRaw (essentially letting the call decide the value, to, and body)

// this means we need to have a test body

// we need to be able to convert the transaction input into a BoC
// so the user says a to, value, body
// we convert that data to a hash (this specifies the target but not that wallet or nonce, could be double spent)

// we need to reflect the op, data, hash, and value to be locked

// the escrow will be empty

// normally this is generated by the wallet
// our client will verify the gas
// no proxy wallet init required (supposedly), I think we need to evaluate this is true but later
// for now we have the init generation in the op

/*
*

need to create this function in golang

	slice calculate_user_proxy_wallet_address(int evm_address, slice entrypoint_address, cell proxy_wallet_code) inline {
	  return calculate_proxy_wallet_address(calculate_proxy_wallet_state_init(evm_address, entrypoint_address, proxy_wallet_code));
	}

then this

	slice calculate_proxy_wallet_address(cell state_init) inline {
	  return begin_cell().store_uint(4, 3)
	                     .store_int(workchain, 8)
	                     .store_uint(cell_hash(state_init), 256)
	                     .end_cell()
	                     .begin_parse();
	}

then we need to build a message to this with state_init code

	cell calculate_proxy_wallet_state_init(int evm_address, slice entrypoint_address, cell proxy_wallet_code) inline {
	  return begin_cell()
	          .store_uint(0, 2)
	          .store_dict(proxy_wallet_code)
	          .store_dict(pack_proxy_wallet_data(0, evm_address, entrypoint_address, proxy_wallet_code))
	          .store_uint(0, 1)
	         .end_cell();
	}
*/
type Signature struct {
	V uint64
	R string // hex string
	S string // hex string
}

func signatureToCell(signature Signature) (*cell.Cell, error) {
	r := new(big.Int)
	s := new(big.Int)

	if _, ok := r.SetString(signature.R, 16); !ok {
		return nil, fmt.Errorf("invalid R hex string: %s", signature.R)
	}
	if _, ok := s.SetString(signature.S, 16); !ok {
		return nil, fmt.Errorf("invalid S hex string: %s", signature.S)
	}

	c := cell.BeginCell().
		MustStoreUInt(signature.V, 8).
		MustStoreBigInt(r, 256).
		MustStoreBigInt(s, 256)

	return c.EndCell(), nil
}

type ExecutionData struct {
	Regime      bool
	Destination string
	Value       *big.Int
	Body        *cell.Cell
}

func executionDataToCell(data ExecutionData) *cell.Cell {
	destAddr := address.MustParseRawAddr(data.Destination)
	c := cell.BeginCell().
		MustStoreBoolBit(data.Regime).
		MustStoreAddr(destAddr).
		MustStoreBigCoins(data.Value)

	if data.Body != nil {
		c.MustStoreRef(data.Body)
	}

	return c.EndCell()
}

func hashCellWithEthereumPrefix(cellData []byte) ([]byte, error) {
	initialHash := sha3.NewLegacyKeccak256()
	initialHash.Write(cellData)
	cellHash := initialHash.Sum(nil)

	prefix := []byte("\x19Ethereum Signed Message:\n32")
	prefixedMessage := append(prefix, cellHash...)

	finalHash := sha3.NewLegacyKeccak256()
	finalHash.Write(prefixedMessage)
	return finalHash.Sum(nil), nil
}

type ProxyWalletMessage struct {
	QueryId   uint64
	Signature Signature
	Data      ExecutionData
}

func proxyWalletMessageToCell(message ProxyWalletMessage) *cell.Cell {
	signatureCell, _ := signatureToCell(message.Signature)
	executionDataCell := executionDataToCell(message.Data)

	c := cell.BeginCell().
		MustStoreUInt(11, 32).
		MustStoreUInt(message.QueryId, 64).
		MustStoreRef(signatureCell).
		MustStoreRef(executionDataCell)

	return c.EndCell()
}

func packProxyWalletData(nonce uint64, entrypointAddress *address.Address, ownerEvmAddress uint64, ownerTvmAddress *address.Address) *cell.Cell {
	c := cell.BeginCell().
		MustStoreUInt(nonce, 64).
		MustStoreAddr(entrypointAddress).
		MustStoreUInt(ownerEvmAddress, 160).
		MustStoreAddr(ownerTvmAddress)
	return c.EndCell()
}

func calculateProxyWalletStateInit(evmAddress uint64, tvmAddress *address.Address, entrypointAddress *address.Address, proxyWalletCode *cell.Dictionary) *cell.Cell {
	proxyWalletData := packProxyWalletData(0, entrypointAddress, evmAddress, tvmAddress)
	return cell.BeginCell().
		MustStoreUInt(0, 2).
		MustStoreDict(proxyWalletCode).
		MustStoreDict(proxyWalletData.BeginParse().MustLoadDict(256)).
		MustStoreUInt(0, 1).
		EndCell()
}

func calculate_proxy_wallet_address(state_init *cell.Cell, workchain int) *cell.Cell {
	hash := state_init.Hash()
	return cell.BeginCell().
		MustStoreUInt(4, 3).
		MustStoreInt(int64(workchain), 8).
		MustStoreUInt(binary.LittleEndian.Uint64(hash), 256).
		EndCell()
}

func entrypointMessageWithProxyInit(
	evmAddress uint64,
	tvmAddress *address.Address,
	entrypointAddress *address.Address,
	proxyWalletCode *cell.Dictionary,
	message ProxyWalletMessage,
	workchain int,
) (*cell.Cell, error) {
	stateInit := calculateProxyWalletStateInit(evmAddress, tvmAddress, entrypointAddress, proxyWalletCode)
	proxyAddress, err := CellToAddress(calculate_proxy_wallet_address(stateInit, workchain))
	if err != nil {
		return nil, err
	}
	proxyWalletMsgCell := proxyWalletMessageToCell(message)

	proxyBody := cell.BeginCell().
		MustStoreRef(stateInit).
		MustStoreRef(proxyWalletMsgCell).
		EndCell()

	entrypointBody := cell.BeginCell().
		MustStoreAddr(proxyAddress).
		MustStoreRef(proxyBody).
		EndCell()

	entrypointMsg := cell.BeginCell().
		MustStoreUInt(1, 32).             // op
		MustStoreUInt(rand.Uint64(), 64). // queryId: random uint64 (we currently aren't checking for collisions)
		MustStoreRef(entrypointBody).
		EndCell()

	return entrypointMsg, nil
}

func entrypointMessageWithoutProxyInit(
	evmAddress uint64,
	tvmAddress *address.Address,
	entrypointAddress *address.Address,
	proxyWalletCode *cell.Dictionary,
	message ProxyWalletMessage,
	workchain int,
) (*cell.Cell, error) {
	stateInit := calculateProxyWalletStateInit(evmAddress, tvmAddress, entrypointAddress, proxyWalletCode)
	proxyAddress, err := CellToAddress(calculate_proxy_wallet_address(stateInit, workchain))
	if err != nil {
		return nil, err
	}
	proxyWalletMsgCell := proxyWalletMessageToCell(message)

	entrypointBody := cell.BeginCell().
		MustStoreAddr(proxyAddress).
		MustStoreRef(proxyWalletMsgCell).
		EndCell()

	entrypointMsg := cell.BeginCell().
		MustStoreUInt(1, 32).             // op
		MustStoreUInt(rand.Uint64(), 64). // queryId: random uint64 (we currently aren't checking for collisions)
		MustStoreRef(entrypointBody).
		EndCell()

	return entrypointMsg, nil
}

func createEntrypointMessage(
	evmAddress uint64,
	tvmAddress *address.Address,
	entrypointAddress *address.Address,
	proxyWalletCode *cell.Dictionary,
	message ProxyWalletMessage,
	workchain int,
	withProxyInit bool,
) (*cell.Cell, error) {
	var entrypointMessage *cell.Cell
	var err error
	if withProxyInit {
		entrypointMessage, err = entrypointMessageWithProxyInit(evmAddress, tvmAddress, entrypointAddress, proxyWalletCode, message, workchain)
	} else {
		entrypointMessage, err = entrypointMessageWithoutProxyInit(evmAddress, tvmAddress, entrypointAddress, proxyWalletCode, message, workchain)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create entrypoint message: %w", err)
	}

	finalCell := cell.BeginCell().
		MustStoreUInt(1, 32).            // operation (op)
		MustStoreUInt(1234, 64).         // queryId, update with actual queryId
		MustStoreRef(entrypointMessage). // store the entrypoint message cell reference
		EndCell()

	return finalCell, nil
}

func derivePrivateKeyFromMnemonic(mnemonic string) (ed25519.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	hash := sha256.Sum256(seed)
	privKey := ed25519.NewKeyFromSeed(hash[:32])
	return privKey, nil
}

func createWalletFromMnemonic(mnemonic string, api wallet.TonAPI, version wallet.VersionConfig) (*wallet.Wallet, error) {
	privKey, err := derivePrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %v", err)
	}

	wallet, err := wallet.FromPrivateKey(api, privKey, version)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	return wallet, nil
}

func connectToClient(config string) (context.Context, ton.APIClientWrapped, error) {
	client := liteclient.NewConnectionPool()

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("get config err: ", err.Error())
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connection err: ", err.Error())
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	ctx := client.StickyContext(context.Background())
	return ctx, api, nil
}

func ConnectToTestnetClient() (context.Context, ton.APIClientWrapped, error) {
	return connectToClient("https://ton.org/testnet-global.config.json")
}

func ConnectToMainnetClient() (context.Context, ton.APIClientWrapped, error) {
	return connectToClient("https://ton.org/global.config.json")
}

func CalculateWallet(
	evmAddress uint64,
	tvmAddress *address.Address,
	entrypointAddress *address.Address,
	workchain int,
) (*address.Address, error) {
	fmt.Print("\ninside of CalculateWallet\n")
	proxyWalletCode, err := ByteArrayToCellDictionary(proxyWalletCodeBytes)
	if err != nil {
		return nil, err
	}

	// deserialization does not match es6
	// blah, _ := cell.FromBOC(proxyWalletCodeBytes)
	//data, _ := cell.FromBOC(proxyWalletCodeHex)
	// data, _ := hex.DecodeString(proxyWalletCodeHex)
	// // data2 := hex.Dump(data)
	// tl.FromBytes(data)
	// blah, _ := cell.FromBOC(data)
	// blah2 := blah.ToBOC()
	// blah3, _ := cell.FromBOC(blah2)
	// blah4 := blah3.ToBOC()
	// //hh := blah[0]
	// fmt.Print("\nresult of proxyWalletCode0\n: %s", proxyWalletCodeHex)
	// fmt.Print("\nresult of proxyWalletCode1\n: %s", hex.EncodeToString(blah2))
	// fmt.Print("\nresult of proxyWalletCode2\n: %s", hex.EncodeToString(blah4))
	// // gg, _ := blah.MarshalJSON()
	// fmt.Print("\nresult of proxyWalletCode3\n: %s", hex.EncodeToString(gg))
	// fmt.Print("\nresult of proxyWalletCode3\n: %s", hex.Dump(gg))

	//fmt.Print("\nresult of proxyWalletCode4\n: %s", blah.Hash())

	stateInit := calculateProxyWalletStateInit(evmAddress, tvmAddress, entrypointAddress, proxyWalletCode)
	proxyAddress, err := CellToAddress(calculate_proxy_wallet_address(stateInit, workchain))
	if err != nil {
		return nil, err
	}

	return proxyAddress, nil
}

// func buildMessage() {
// 	w := &wallet.Wallet{}
// }

// func CheckProxyWalletInitialized(proxyAddress string) (bool, error) {
// 	// Create a new instance of the TON client
// 	client := your_ton_client_library.NewClient()

// 	state, err := client.GetContractState(proxyAddress)
// 	if err != nil {
// 		return false, err
// 	}

// 	if state.CodeSize > 0 {
// 		return true, nil
// 	}

// 	return false, nil
// }

// both need to be fixed
type MessageEscrowTvm struct {
	EscrowAddress   string `json:"eaddress"`
	EscrowInit      string `json:"einit"`
	EscrowPayload   string `json:"epayload"`
	EscrowAsset     string `json:"easset"`
	EscrowAmount    string `json:"eamount"`
	EscrowValueType string `json:"evaluetype"`
	EscrowValue     string `json:"evalue"`
}

type EntrypointMessageParams struct {
	EvmAddress        string `query:"pw-evm-address"`
	EntrypointAddress string `query:"pw-entrypoint"`
	ProxyWalletCode   string `query:"pw-code"`
}

type UnsignedEscrowRequestParams struct {
	Header utils.PartialHeader `query:"header"`
	Escrow EscrowLockParams    `query:"escrow"`
}

// Start of UnsignedEntryPointRequestParams
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

// UnsignedEntryPointRequestResponse:
type MessageOpTvm struct {
	Header       utils.MessageHeader `json:"header"`
	ProxyParams  ProxyParams         `json:"proxy"`
	ProxyAddress string              `json:"proxy-address"`
	ValueNano    string              `json:"value"`
	MessageHash  string              `json:"hash"`
}

// End of UnsignedEntryPointRequestParams

// NEED to create the escrow payload
type EscrowLockParams struct {
	SignerAddress string `query:"signer-address"`
	AdminAddress  string `query:"admin-address" optional:"true"`
	PayeeAddress  string `query:"payee-address" optional:"true"`
	Id            string `query:"id" optional:"true"`
	Value         string `query:"value" optional:"true"`
}

// message directly to entrypoint
type ExternalMessageRaw struct {
	Via      string `json:"via"` // the solver sender (the backend)
	Value    string `json:"value"`
	SendMode string `json:"send-mode"`
	Body     string `json:"body"`
}

type EntryPointMessageRaw struct {
	Destination string                  `json:"proxy-wallet"`
	QueryId     string                  `json:"query-id"` // kinda pointless atm
	Message     MessageToProxyWalletRaw `json:"proxy-message"`
}

type MessageToProxyWalletRaw struct {
	QueryId  string          `json:"msg-id"` // no security but used for onchain hash
	Sigature SignatureRaw    `json:"msg-sig"`
	Data     ProxyMessageRaw `json:"msg-data"`
}

type SignatureRaw struct {
	V string `query:"v"` // tvm is often a garbage value at least from ts
	R string `query:"r"`
	S string `query:"s"`
}

type ProxyMessageRaw struct {
	Regime      string `query:"msg-regime"` // optional, set to 0 by default, apparently 0 or 1
	Destination string `query:"msg-to"`     // required
	Value       string `query:"msg-value"`  // required
	Body        string `query:"msg-body"`   // required
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

func (m MessageOpTvm) GetType() string {
	return "EVM UserOp"
}
