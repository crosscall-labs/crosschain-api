package handler

type UnsignedBytecodeParams struct {
	MessageType  string `query:"msg-type" optional:"true"`
	Signer       string `query:"signer"`
	TargetId     string `query:"destination-id"`
	OriginId     string `query:"origin-id"`
	AssetAmount  string `query:"asset-amount"`
	AssetAddress string `query:"asset-address"`
	Calldata     string `query:"calldata" optional:"true"`
}

type SignedBytecodeParams struct {
	Signer                   string `query:"signer"`
	DestinationId            string `query:"destination-id"`
	OriginId                 string `query:"origin-id"`
	AssetAddress             string `query:"asset-address"`
	AssetAmount              string `query:"asset-amount"`
	UseropSender             string `query:"useropSender"`
	UseropNonce              string `query:"useropNonce"`
	UseropInitCode           string `query:"useropInitCode" optional:"true"`
	UseropCallData           string `query:"useropCallData"`
	UseropAccountGasLimit    string `query:"useropAccountGasLimit"`
	UseropPreVerificationGas string `query:"useropPreVerificationGas"`
	UseropGasFees            string `query:"useropGasFees"`
	UseropPaymasterAndData   string `query:"useropPaymasterAndData"`
	UseropSignature          string `query:"useropSignature"`
}

type SignedEscrowPayoutParams struct {
	Bytecode string `query:"data"`
	TraceId  string `query:"traceid"`
}
