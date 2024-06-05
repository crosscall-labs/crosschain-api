## TODO
- [x] multicall enables
- [x] interpret data
- [x] check if scw exists
- [x] check if escrow exists
- [x] create init scw code
- [x] create init escrow code
- [x] create transfer from code
- [x] create transfer from execute code
- [x] create test tx for the frontend to sign
  - [x] user can signs via the frontend
  - [x] data is sent from DEX API to Crosschain API
  - [x] receipt and response returned to DEX API
- [x] create userop
- [x] create userop hash
- [x] accept signed userop
- [x] validate escrow deadline & value
- [x] validate signed userop
- [x] execute signed userop
- [x] record telemetry on all requests
- [ ] correct EscrowValue to include gas and paymaster

## Crosschain API v0.0.3

The premise of the crosschain execution is an arbitrary native atomic execution by a solver.

This crosschain API facilitates:
- the generation of the user operation for the signer
- the escrow bytecode for the signer to execute
- the validation of escrow locktime and locked value
- the validation of submitted user operations
- the relay substitue for the Hyperlane relay network to streamline execution

The API has currently two primary functions: Request calls and Submit calls.

Request calls to the API:

The generation of the user operation takes the data created by the protocol conencting to the API. This data includes the signer EOA, calldata, transaction value, transaction asset address, transaction bid, and destination chain ID. The calldata is the transaction to be executed on the user destination chain abstraction account. The user operation will take the form of a `PackedUserOperation` defined in [ERC4337](https://eips.ethereum.org/EIPS/eip-4337). The `PaymasterAndData` field will be specialized to house the deserialized bytecode to be executed on the origin chains ecrow, including the information about the asset address and transaction full value.

The generation of the escrow bytecode is a all-or-none multicall execution of the escrow account generation (if applicable), depositing of escrow funds, and escrow account timelock extenstion. The deposit call has a header and nonce only to be valid for the specified chain. The Escrow is not a prerequisite of the user operation execution, but upon validation if the escrow has insufficent funds the relay will not execute the user operation.

Submit calls to the API:

The API accepts the 