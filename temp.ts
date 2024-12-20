import { addressFromPublicKey, generateKeyPair } from '../wrappers/utils/cryptography';



let genKey = generateKeyPair();
let privateKey2 = genKey.privateKey;
let publicKey2 = genKey.publicKey;
Buffer.from(publicKey2).toString('base64')
//let {privateKey2, publicKey2} =  {, genKey.publicKey }

export async function run(provider: NetworkProvider) {
    const ui = provider.ui();
    const privateKey = await ui.input('privateKey');
    const publicKey = await ui.input('publicKey');

    const queryId = Math.floor(Math.random() * 10000);
    const ownerEvmAddress = addressFromPublicKey(publicKey);

    await sendExecuteNativeTransfer(
        provider.api() as TonClient4,
        provider.sender(),
        toNano(0.35),
        ENTRYPOINT_ADDRESS.toString(),
        {
            privateKey,
            value: toNano(50),
            regime: 0,
            queryId,
            ownerEvmAddress,
            destination: Address.parse("0QCyq_hrs4smOyq_uQ5P-BBAiM-SJtkd6V6FE3MnNJFhWC0J")
        }
    )
}


sendNativeTransfer

userAddress: Address.parse('UQAXOW_uAJ1Q-KI4kv4_WShvhENcZ7D2-fjqo9WQkdVi-Q68'), // need to input
adminAddress: Address.parse('EQDFPU5bJ7t5LPLlm7z7gAj35ovBs5BnPUtJ1wPoSSDxu-4N'),

senLockWithSignature

