	// messaeType enum: legacy, ecdsa, eddsa, secp256k1, secp256r1, secp256k1-1byte
	// case "unsigned-bytecode2":
	// 	if messageType == "" ||
	// 		signer == "" ||
	// 		destinationId == "" ||
	// 		originId == "" ||
	// 		assetAddress == "" ||
	// 		assetAmount == "" ||
	// 		calldata == "" {
	// 		errMalformedRequest(w)
	// 		return
	// 	}
	// 	messageTypeInt, err := strconv.Atoi(messageType)
	// 	if err != nil {
	// 		fmt.Println("Invalid integer string:", err)
	// 		errMalformedRequest(w)
	// 		return
	// 	}

	// 	var unsignedDataResponse UnsignedDataResponse2

	// 	// need to refactor this api as root handler and modularize evm/svm handling
	// 	// func checkChainType(chainId string, messageType int) (string, error)
	// 	// return type "evm", "tvm", "svm"
	// 	// chainType, entrypointTypes, _, err := checkChainType(originId)
	// 	// if err != nil {
	// 	// 	json.NewEncoder(w).Encode(err)
	// 	// 	return
	// 	// }
	// 	// err = hasInt(entrypointTypes, messageTypeInt)
	// 	// if err != nil {
	// 	// 	json.NewEncoder(w).Encode(err)
	// 	// 	return
	// 	// }
	// 	// switch chainType {
	// 	// case "evm":
	// 	// 	return createEscrowBytecodeEVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
	// 	// case "tvm":
	// 	// 	return createEscrowBytecodeTVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
	// 	// case "svm":
	// 	// 	return createEscrowBytecodeSVM(messageTypeInt, signer, originId, assetAddress, assetAmount)
	// 	// default:
	// 	// 	return "", "", &Error{
	// 	// 		Code:    500,
	// 	// 		Message: "Internal error: chain type could not be determined",
	// 	// 	}
	// 	// }
	// 	chainType, entrypointTypes, _, err := checkChainType(originId)
	// 	if err != nil {
	// 		json.NewEncoder(w).Encode(err)
	// 		return
	// 	}
	// 	switch chainType {
	// 	case "evm":
	// 		err := hasInt(entrypointTypes, messageTypeInt)
	// 		if err != nil {
	// 			json.NewEncoder(w).Encode(err)
	// 			return
	// 		}
	// 		unsignedDataResponse.Escrow, unsignedDataResponse.EscrowInit, err = createEscrowBytecodeEVM(messageTypeInt, signer, originId, assetAddress, assetAmount) // but need to build the uerop to estimate bid cost
	// 		if err != nil {
	// 			json.NewEncoder(w).Encode(err)
	// 			return
	// 		}
	// 		// type UnsignedDataResponse2 struct {
	// 		// 	Signer           string                      `json:"signer"`
	// 		// 	ScwInit          bool                        `json:"swc-init"`
	// 		// 	Escrow           string                      `json:"escrow"`
	// 		// 	EscrowInit       bool                        `json:"escrow-init"`
	// 		// 	EscrowPayload    string                      `json:"escrow-payload"`
	// 		// 	EscrowAsset      string                      `json:"escrow-asset"`
	// 		// 	EscrowValue      string                      `json:"escrow-value"`  // need to implement
	// 		// 	UserOp           PackedUserOperationResponse `json:"packed-userop"` // parsed data, recommended to validate data
	// 		// 	PaymasterAndData PaymasterAndData            `json:"paymaster-and-data"`
	// 		// 	UserOpHash       string                      `json:"userop-hash"`
	// 		// }
	// 		// for tvm this is the escrow factory not the tba escrow address
	// 		//func createEscrowBytecode(messageTypeInt int, signer string, originId string, assetAddress string, assetAmount string) (string, string, error) {
	// 		// unsignedDataResponse.Escrow, unsignedDataResponse.EscrowInit, err = createEscrowBytecode(messageTypeInt, signer, originId, assetAddress, assetAmount) // but need to build the uerop to estimate bid cost
	// 		// if err != nil {
	// 		// 	json.NewEncoder(w).Encode(err)
	// 		// 	return
	// 		// }

	// 		//createUserOpBytecode(messageType, signer, originId, destinationId, assetAddress, assetAmount, calldata, escrowAddress)
	// 		// escrow address
	// 		// escrow factory
	// 		// data to create escrow
	// 		// data to deposit and lock
	// 		// data to extendlocktime
	// 		// combined call to gas savings

	// 		// the

	// 	}
	// 	chainType, _, escrowTypes, err := checkChainType(destinationId)
	// 	if err != nil {
	// 		json.NewEncoder(w).Encode(err)
	// 		return
	// 	}
	// 	switch chainType {
	// 	case "evm":
	// 		err := hasInt(entrypointTypes, messageTypeInt)
	// 		if err != nil {
	// 			json.NewEncoder(w).Encode(err)
	// 			return
	// 		}
	// 		// create the userop data to sign
	// 	fmt.Printf("escrowTypes: %s", escrowTypes)

	// 	// if chainType0 == 0 || chainType0 == 1 {
	// 	// 	// this will call evm api
	// 	// 	// evaluate escrow
	// 	// }

	// 	client, chainInfo, err := checkChainStatus(originId)
	// 	if err != nil {
	// 		json.NewEncoder(w).Encode(err)
	// 		return
	// 	}
	// 	if client == nil {
	// 		errUnsupportedChain(w)
	// 		return
	// 	}
	// 	fmt.Printf("chainInfo: %s", chainInfo)

	// 	client2, chainInfo2, err := checkChainStatus(destinationId)
	// 	if err != nil {
	// 		json.NewEncoder(w).Encode(err)
	// 		return
	// 	}
	// 	if client2 == nil {
	// 		errUnsupportedChain(w)
	// 		return
	// 	}
	// 	fmt.Printf("chainInfo2: %s", chainInfo2)
	// 	w.Header().Set("Content-Type", "application/json")
	// 	if err := json.NewEncoder(w).Encode(unsignedDataResponse); err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}