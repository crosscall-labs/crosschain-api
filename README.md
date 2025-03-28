## TODO
- [ ] refactor crosschain api, evm, svm, tvm to be seperate apis
- [ ] execute tvm
- [x] enable token only execution
- [x] enable new signature methods
- [x] enable message typing
- [ ] write design flow for TVM
- [ ] eventually create golang port of tvm ts

## Crosschain API v0.0.4a

The premise of the crosschain execution is an arbitrary native atomic execution by a solver.

This crosschain API facilitates:
- the generation of the user operation for the signer
- the escrow bytecode for the signer to execute
- the validation of escrow locktime and locked value
- the validation of submitted user operations
- the relay substitue for the Hyperlane relay network to streamline execution

The API has currently two primary functions: Request calls and Submit calls.

Request calls to the API:

The generation of the user operation takes the data created by the protocol conencting to the API. This data includes the signer EOA, calldata, calldata target, transaction value, transaction asset address, transaction bid, and destination chain ID. The calldata is the transaction to be executed on the user destination chain abstraction account. The user operation will take the form of a `PackedUserOperation` defined in [ERC4337](https://eips.ethereum.org/EIPS/eip-4337). The `PaymasterAndData` field will be specialized to house the deserialized bytecode to be executed on the origin chains ecrow, including the information about the asset address and transaction full value.

The generation of the escrow bytecode is a all-or-none multicall execution of the escrow account generation (if applicable), depositing of escrow funds, and escrow account timelock extenstion. The deposit call has a header and nonce only to be valid for the specified chain. The Escrow is not a prerequisite of the user operation execution, but upon validation if the escrow has insufficent funds the relay will not execute the user operation.

Submit calls to the API:

The API accepts the full user operation with the filled signature field, calldata target, transaaction value, transaction asset address, transaction bid, and destination chain ID.

The API will validate the the escrow locktime and lock value for the target chain using a multicall view request to the target chain. This call will validate the escrow has a sufficent lock time exceeding the processing time of one epoch on the destination chain plus one epoch of the origin chain. For sake of the MVP the lock time requirement is set to greater than one hour. The asset value will be evaluated from the signer's escrow account (which is recovered determinstically).

The validation of the user operation suppliments the implementation of our modifed version of (Silius Bundler)[https://github.com/silius-rs/silius] for the MVP. The user operation is validated for valid execution, valid nonce, and valid ECDSA signer.

The API lastly will execute the user operation on the destination chain if the previous validation was successfull. The relayer EOA will execute on the usser operation on chain entrypoint contract. The processing onn the operation then goes through the phases: preOp, handler, and postOp. During the preOp the user operation will be validated on chain and the message to the origin chain will be executed by the paymaster to Hyperlane. The handler will execute the user operationc calldata on the target (the SCW). The postOp will finish paying for the Hyperlane message.

For the sake of the MVP, upon receipt of the validly executed user operation transaction, the relay will execute the payout message via the Hyperlane contract on the origin chain. This execution will on-chain validate the msg.sender and message data.