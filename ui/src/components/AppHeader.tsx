import {
    Box,
    HStack,
    IconButton,
    Kbd,
    Spacer,
    Spinner,
    Text,
    Tooltip,
    useBreakpointValue,
    useInterval
} from "@chakra-ui/react";
import {NodeContext} from "../context";
import {ColorModeSwitcher} from "./ColorModeSwitcher";
import {BiCoin, CgFileDocument, MdBrightness1, MdRefresh} from "react-icons/all";
import * as React from "react";
import {useCallback, useContext, useEffect, useState} from "react";
import {CryptoNode} from '../types/crypto-node';
import {useParams} from "react-router-dom";
import {NodeChooser} from "./NodeChooser";

import {useLocalStorage} from "react-use";
import {getClaims, getHeight, getNode} from "../MonitoringService";
import {HamburgerMenu} from "./HamburgerMenu";
import {PendingRelaysBadge} from "./badges/PendingRelaysBadge";
import {MonthlyReward} from "../types/monthly-reward";
import {getActivePath, pathIdErrors, pathIdRewards} from "../App";
import {QuestionOutlineIcon} from "@chakra-ui/icons";

type AppHeaderProps = {
    address: string,
    onNodeLoaded: (n: CryptoNode) => void,
    rewards: MonthlyReward[],
    onRewardsLoaded: (m: MonthlyReward[]) => void,
    isRefreshing: boolean,
    setIsRefreshing: (is: boolean) => void,
}

export const AppHeader = (props: AppHeaderProps) => {
    const defaultSavedAddresses: Array<string> = [];
    const [savedAddresses] = useLocalStorage("savedAddresses", defaultSavedAddresses);
    const [currentAddress, setCurrentAddress] = useLocalStorage("currentAddress", "");
    const isMobile = useBreakpointValue([true, false]);
    const [currentHeight, setCurrentHeight] = useState(0);
    const [pendingRelays, setPendingRelays] = useState(0);
    const [poktAmt, setpoktAmt] = useState(0.00);
    const node = useContext(NodeContext);
    const activePath = getActivePath();

    let {address} = useParams();
    if(address === '' || address === undefined) {
        if(currentAddress !== '') {
            address = currentAddress;
        } else if(savedAddresses !== undefined && savedAddresses.length > 0) {
            setCurrentAddress(savedAddresses[0]);
        }
    }

    const [hasLoaded, setHasLoaded] = useState(false);
    let statusColor = "#797979";
    if (node.isJailed || !node.exists) {
        statusColor = "#FF0000";
    } else if((node.latestBlockHeight !== 0) && node.latestBlockHeight < currentHeight) {
        statusColor = "#ffb700";
    } else {
        statusColor = "#2bd950";
    }
    
    const status = (node.isJailed || !node.exists) ?
        (node.exists ? "Jailed" : "Invalid address") : "Not Jailed";

    const updateBalance = useCallback(() => {
        if(!address) {
            return;
        }

        getNode(address).then((node) => {
            setCurrentAddress(node.address);
            props.onNodeLoaded(node);
        })
            .catch((err) => console.error(err))
            .finally(() => {
                setHasLoaded(true);
            });
    }, [props, address, setCurrentAddress]);

    const updateNodeData = useCallback(async () => {
        if(!node.address) {
            return;
        }

        props.setIsRefreshing(true);
        try {
            const n = await getNode(node.address);
            props.onNodeLoaded(n);
            getHeight().then((h) => setCurrentHeight(h));
            updateBalance();

            const c = await getClaims(node.address);
            props.onRewardsLoaded(c);
        }
        catch (err) {
            console.error("updateNodeData", err);
        }
        props.setIsRefreshing(false);

    }, [node, props, updateBalance]);
    useInterval(updateNodeData, 900000);

    useEffect(() => {
        if(!hasLoaded) {
            updateBalance();
            getHeight().then((h) => setCurrentHeight(h));
        }

        if(props.rewards[0]) {
            let pending = 0, poktAmount = 0;
            props.rewards[0].transactions.map((t) => {
                if(!t.is_confirmed && t.expire_height > currentHeight) {
                    pending += t.num_relays;
                    poktAmount += t.reward.amount;
                }
                return t;
            })
            setPendingRelays(pending);
            setpoktAmt(poktAmount);
        }
    }, [pendingRelays, props, currentHeight, hasLoaded, updateBalance]);


    return (
        <HStack justifyContent={"space-between"} mt={[1,2]} w={"100%"}>
            <HamburgerMenu/>
            {(!isMobile && node.address !== '') && (
                <HStack>

                    <IconButton
                        icon={(<BiCoin/>)}
                        variant={activePath === pathIdRewards ? 'outline' : 'ghost'}
                        aria-label={'rewards'}
                        title={"Rewards"}
                        onClick={() => window.location.href=`/node/${node.address}/rewards`}
                    >
                        Rewards
                    </IconButton>
                    <IconButton
                        aria-label={'logs'}
                        icon={(<CgFileDocument/>)}
                        variant={activePath === pathIdErrors ? 'outline' : 'ghost'}
                        title={"Logs"}
                        onClick={() => window.location.href=`/node/${node.address}/errors`}
                    >
                        Logs
                    </IconButton>

                    {/*<ConnectedChainsBadge/>*/}

                    <IconButton
                        ml={4} mr={4}
                        aria-label={"Refresh"}
                        variant={"ghost"}
                        _focus={{boxShadow: 0}}
                        icon={props.isRefreshing ? (<Spinner size={"xs"}/>) : (<MdRefresh/>)}
                        onClick={updateNodeData}
                    />
                    <Box  color={"gray.400"} ml={8} fontSize={"sm"}>
                        <em>Updated: {node.lastChecked?.toLocaleString()}</em>
                    </Box>
                </HStack>

            )}
            <Box ml={2}><PendingRelaysBadge num={pendingRelays} amt={poktAmt}/></Box>
            <Spacer/>
            <Box pl={2} fontSize={"sm"}>
                <Text d="inline" fontSize="xs" fontWeight={600} textTransform={"uppercase"}>height:</Text>
                {node.latestBlockHeight === 0 ? (
                    <Tooltip label={"pokt.tools can't reach "+node.service_url+"/v1/query/height"}>
                        <QuestionOutlineIcon ml={2} mr={2}/>
                    </Tooltip>
                ) : (
                    <Kbd title={"node Height"}>{node.latestBlockHeight}</Kbd>
                )}
                <>/ <Kbd title={"network height"}>{currentHeight}</Kbd></>
            </Box>
            <Box>
                <MdBrightness1
                    title={"Node status: " + status}
                    aria-label={"Node status: " + status}
                    fill={statusColor}
                />
            </Box>

            <Box pt={2} pr={2}>
                <NodeChooser address={address ?? ''}/>
            </Box>

            {!isMobile && (
                // <Box alignSelf={"flex-start"}>
                <Box pt={2} pr={2}>
                    <ColorModeSwitcher
                        _focus={{boxShadow: "none"}}
                        alignSelf="flex-end"
                    />
                </Box>
            )}
        </HStack>
    )
}