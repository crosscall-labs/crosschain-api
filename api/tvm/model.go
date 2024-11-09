package tvmHandler

/*
tvm message to the entrypoint
the concept of this message is we describe a message with empty body to transfer funds from the proxy wallet to the solver2.address
const queryId = Math.floor(Math.random() * 10000);
const passing = await entrypoint.sendExecute(solver1.getSender(),
		toNano('5'),
		{
				destination: proxyWallet.address,
				queryId: queryId,
				messageToProxyWallet: {
						queryId: queryId,
						signature: {
								v: 0,
								r: '0',
								s: '0',
						},
						data: {
								regime: 0,
								destination: solver2.address,
								value: toNano('5'),
								body: beginCell().endCell(),
						},
				},
		}
		);

expect(passing.transactions).toHaveTransaction({
		from: solver1.address,
		to: entrypoint.address,
		success: true,
});

expect(passing.transactions).toHaveTransaction({
		from: entrypoint.address,
		to: proxyWallet.address,
});



tvm message to proxy wallet that will be signed and subsquently sent to the entrypoint by the solver
    describe('Execution', () => {
        console.log("Execution tests");
        let message: ProxyWalletMessage  = {
            queryId: Math.floor(Math.random() * 1000000),
            signature: {
                v: 0,
                r: '0',
                s: '0'
            },
            data: {
                regime: 0,
                destination: randomAddress(),
                value: toNano("0.1"),
                body: beginCell().endCell()
                }
            }
        console.log("Message ", message);


const signature = signMessageDataEth(privateKey, message.data);

export function signMessageDataEth(privateKey: string | bigint, messageData: ExecutionData): Signature {
    let slice = beginCell()
                .storeUint(messageData.regime, 1) // 1 bit
                .storeAddress(messageData.destination) // 267 bits
                .storeCoins(messageData.value) // 124 bits
                .storeRef(messageData.body)  // 0 bits
            .asSlice();
    const messageDataBuffer = packSliceToBuffer(slice)
    const signature = signDataEth(privateKey, messageDataBuffer);
    return signature;
}

 export function signDataEth(privateKey: string | bigint, data: Buffer): Signature {
	const preparedHash = dataToEthSignHash(data);
	const signature = secp256k1.sign(preparedHash, privateKey);
	// for some reason the recovery byte is 0, but weirdly the other parts (r,s) are consistent
	return { r: signature.r.toString(16),  s: signature.s.toString(16), v: signature.recovery };
}

export function dataToEthSignHash(data: Uint8Array): Uint8Array {
    const dataHash = keccak256(data);
    return modifyHashToEthSignHash(dataHash);
}
bytes compressed from empty data message is 40 bytes instead on 32
1234567890123456789012345678901234567890123456789012345678901234 6789012345678901234
400e2a75f15ac66566adb43ff18d8c45ef08a91cf7da2659fb091e51c28562fb10c47735940001010001
bc56bef0f34cb0a6c4e05fa0534175aceac5586116fd45e944c34d060b4da98a




we need to have a script to create stonfi liquidity pool and execute in our test environment
this can be done later since transaction (bridge is more important, already proves we have arbitrary execution but still require a reliable way of validating non-atomic txs)





for the escrow
      ;; signature will later be governed using cryptography
      ;; for now we only validate that a payout message is signed on the backend
      ;; goal: 40 bytes escrow address . uint nonce . slice payee . uint amount . signature
      ;; for now: 40 bytes escrow address . uint nonce . slice payee . uint amount . signature


type PaymasterAndDataResponse struct {
	Paymaster                     string `json:"pad-paymaster"`
	PaymasterVerificationGasLimit string `json:"pad-verification-gas-limit"`
	PaymasterPostOpGasLimit       string `json:"pad-post-op-gas-limit"`
	Signer                        string `json:"pad-signer"`
	DestinationDomain             string `json:"pad-destination-domain"`
	MessageType                   string `json:"pad-message-type"`
	AssetAddress                  string `json:"pad-asset-address"`
	AssetAmount                   string `json:"pad-asset-amount"`
}

        let escrowConfig: EscrowConfig = {
            userAddress: user.address,
            adminAddress: deployer.address,
            payee: BigInt(addressFromPublicKey(publicKey)), // error payee is not predefined on the escrow
            id: BigInt('0x' + randomBytes(32).toString('hex')),
            value: toNano(2)
        }



static createFromConfig(config: EscrowConfig, code: Cell, workchain = 0) {
		const data = escrowConfigToCell(config);
		const init = { code, data };
		return new Escrow(contractAddress(workchain, init), init);
}

export function escrowConfigToCell(config: EscrowConfig): Cell {
    return beginCell()
        .storeAddress(config.userAddress)
        .storeAddress(config.adminAddress)
        .storeUint(config.payee, 160)
        .storeUint(config.id, 256)
        .storeRef(beginCell())
        .storeRef(beginCell().storeCoins(config.value))
        .endCell();
}
*/

// we assume the payee == admin (set by the init config)
// will be changed later, for now we will receive funds and delegate rewards after our backend reveives them (workaround)

type PaymasterAndDataBase struct {
	Signer            string `query:"pad-signer"`
	Escrow            string `query:"pad-escrow"`                        // we'll confirm the signer is owner by generating the address using the "config" params
	DestinationDomain string `query:"pad-destination-domain"`            // domain unique to protocol
	MessageType       string `query:"pad-message-type" optional:"true"`  // assume type 1
	AssetAddress      string `query:"pad-asset-address" optional:"true"` // assume 0 (naive ton, later will accept usdc/usdt, maybe more)
	AssetAmount       string `query:"pad-asset-amount"`                  // in nano
	Signature         string `query:"pad-signatrue"`                     // 0 for type 0 and "from" response (to be filled in my user)
}

// workchain := int32(int8(workchain))

type EscrowConfig struct {
	UserAddress  string `query:"e-user"`
	AdminAddress string `query:"e-admin"`
	Payee        string `query:"e-payee"`
	Id           string `query:"e-id" optional:"true"` // needs telemetry to recover
	Value        string `query:"e-value" optional:"true"`
}

type EscrowConfigInput struct {
	UserAddress  string `query:"e-user"`
	AdminAddress string `query:"e-admin"`
	Payee        string `query:"e-payee"`
	Id           string `query:"e-id" optional:"true"` // needs telemetry to recover
	Value        string `query:"e-value" optional:"true"`
}

type PaymasterAndDataType1 struct {
	PaymasterAndData     PaymasterAndDataBase `query:"pad"`
	PaymasterAndDataHash string               `query:"pad-hash"`
}

type Address struct {
	Workchain int32
	Hash      []byte
}

// export type EscrowConfig = {
// 	userAddress: Address;
// 	adminAddress: Address;
// 	payee: bigint;
// 	id: bigint;
// 	value: bigint;
// };

// export function escrowConfigToCell(config: EscrowConfig): Cell {
// 	return beginCell()
// 			.storeAddress(config.userAddress)
// 			.storeAddress(config.adminAddress)
// 			.storeUint(config.payee, 160)
// 			.storeUint(config.id, 256)
// 			.storeRef(beginCell())
// 			.storeRef(beginCell().storeCoins(config.value))
// 			.endCell();
// }

// func EscrowConfigToCell() {

// 	var hexBOC = "b5ee9c724102060100011e000114ff00f4a413f4bcf2c80b01020162020502f8d06c2220c700915be001d0d3030171b0915be0fa403001d31fed44d0fa4001f861fa4001f862d39f01f863d3ff01f864d401d0f866d430d0fa0030f86521c01ee30221c01f8e2031f84212c705f2e192708018c8cb0502fa403012cf1621fa02cb6ac98306fb00e030312082103b307c4bba9130e0208210622117e90304007631f84112c705f2e191d430d0f866f846d749810208baf2e1f5c8f845fa02c9c8f846cf16c9f844f843c8f841cf16f842cf16cb9fcbffccccc9ed540024ba9130e020c01e9130e082100b7ba46fbadc0071a04e37da89a1f48003f0c3f48003f0c5a73e03f0c7a7fe03f0c9a803a1f0cda861a1f40061f0cbf08da60ff0cdf08da7fff0cdf08da7fff0cd5649929b"
// 	codeCellBytes, _ := hex.DecodeString(hexBOC)

// how to go from contract code BoC to cell
// 	codeCell, err := cell.FromBOC(codeCellBytes)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// GenerateContractAddress generates the contract address from the workchain and init cells.
// func GenerateContractAddress(workchain int32, codeCell *cell.Cell, dataCell *cell.Cell) (*Address, error) {
// 	// Step 1: Create a StateInit cell
// 	stateInitCell := NewBuilder()

// 	// Serialize `code` and `data` cells into StateInit cell
// 	if err := stateInitCell.StoreCell(codeCell); err != nil {
// 		return nil, fmt.Errorf("failed to store code cell: %v", err)
// 	}
// 	if err := stateInitCell.StoreCell(dataCell); err != nil {
// 		return nil, fmt.Errorf("failed to store data cell: %v", err)
// 	}

// 	// Step 2: Serialize StateInit and hash it
// 	stateInitSerialized, err := stateInitCell.EndCell().ToBOC()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to serialize state init: %v", err)
// 	}
// 	hash := sha256.Sum256(stateInitSerialized)

// 	// Step 3: Combine workchain and hash to create the address
// 	address := &Address{
// 		Workchain: workchain,
// 		Hash:      hash[:],
// 	}

// 	return address, nil
// }

// for the crosschain api we cannot calculate the escrow address if we don't know the salt input
/*
fot the proxy wallet the backend needs to generate:
	nonce
	entrypoint
	owner_evm_address (this will be genericified to full pubkey length later, this is for easier secp256k1 composability)
	owner_ton_address (this will later be changed)

for the escrow the backend needs to generate:
	userAddress (this will just be assumed to be the ton address for now)
	adminAddress (this will just be the backend for now)
	payee (also set to the backend)
	id (since the escrow will be unique to the user address, this will always be 0)
	value (only non-zero when user is staking which initalized which is m00t since the contract will then have a value anyways)

the backend can simply store the value of the finalized escrow and
	expiration date to a table and just check the table for easy of user (needs to be fixed later on)
*/
