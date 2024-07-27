import {Chain} from './chain'

export type Reward = {
    amount: number,
    stake_weight: number,
    pokt_per_relay: number,
}

export type Transaction = {
    hash: string,
    height: number,
    time: string,
    type: string,
    chain_id: string,
    chain: Chain,
    num_relays: number,
    reward: Reward,
    session_height: number,
    expire_height: number,
    app_pubkey: string,
    is_confirmed?: boolean,
}