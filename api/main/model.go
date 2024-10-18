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
