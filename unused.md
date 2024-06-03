		// // handling raw bytes
		// tuples := [][2]interface{}{
		// 	{common.HexToAddress(chainInfo.AddressEscrow), bytesval0},
		// 	{common.HexToAddress(chainInfo.AddressEscrow), bytesval1},
		// 	{common.HexToAddress(chainInfo.AddressEscrow), bytesval2},
		// 	{common.HexToAddress(chainInfo.AddressEscrow), bytesval3},
		// 	{common.HexToAddress(chainInfo.AddressEscrow), bytesval4},
		// }

		// paddedTuples := make([][]byte, len(tuples))
		// paddedTuplesLen := len(tuples)

		// // create tuple raw bytes
		// for i, tuple := range tuples {
		// 	addrBytes := padLeft(tuple[0].(common.Address).Bytes())
		// 	dataBytes := tuple[1].([]byte)
		// 	paddedLen := ((len(dataBytes) + 31) / 32) * 32 // future error?
		// 	paddedBytes := make([]byte, paddedLen)
		// 	copy(paddedBytes, dataBytes)

		// 	// Concatenate the padded address and padded bytes
		// 	tupleBytes := append(addrBytes, common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000040")...)
		// 	tupleBytes = append(tupleBytes, padLeftHex(len(dataBytes))...)
		// 	tupleBytes = append(tupleBytes, paddedBytes...)
		// 	paddedTuples[i] = tupleBytes
		// }

		// var buffer bytes.Buffer

		// parse, _ := common.ParseHexOrString("multicallView((address,bytes)[])")
		// hash := sha3.NewLegacyKeccak256()
		// hash.Write(parse)
		// selector := hash.Sum(nil)[:4]
		// buffer.Write(selector)

		// buffer.Write(common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000020"))

		// buffer.Write(padLeftHex(paddedTuplesLen))

		// buffer.Write(padLeftHex(paddedTuplesLen * 32))
		// var sum int
		// for i := 1; i < len(paddedTuples); i++ {
		// 	sum += len(paddedTuples[i-1]) // Adjust index to access the correct tuple
		// 	buffer.Write(padLeftHex(sum + paddedTuplesLen*32))
		// }

		// for _, paddedTuple := range paddedTuples {
		// 	buffer.Write(paddedTuple)
		// }

		// bufferBytes := buffer.Bytes()


		func padLeft(b []byte) []byte {
	return append(make([]byte, 32-len(b)), b...)
}

func padRight(b []byte) []byte {
	padded := make([]byte, ((len(b)+31)/32)*32) // Round up to the nearest 32 bytes
	copy(padded, b)
	return padded
}

func encodeUint32(value uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, value)
	return padLeft(b) // Pad the 4-byte uint32 to fit into 32 bytes
}

func padLeftHex(length int) []byte {
	hexStr := fmt.Sprintf("%064x", length)
	padded, _ := hex.DecodeString(hexStr)
	return padded
}

func padLeft(b []byte) []byte {
    padded := make([]byte, 32)
    copy(padded[32-len(b):], b)
    return padded
}