import {
    Box, Button,
    HStack,
    IconButton,
    Popover,
    PopoverBody,
    PopoverContent,
    PopoverTrigger, Spacer,
    useBreakpointValue
} from "@chakra-ui/react";
import {HomeButton} from "./HomeButton";
import {NodeContext} from "../node-context";
import {ColorModeSwitcher} from "./ColorModeSwitcher";
import {FaGithub, MdBrightness1} from "react-icons/all";
import * as React from "react";
import {CryptoNode, NodeProps} from '../types/crypto-node';
import {useParams} from "react-router-dom";
import {NodeChooser} from "./NodeChooser";
import {useCallback, useContext, useEffect, useState} from "react";
import axios from "axios";
import {NodeStatus} from "./NodeStatus";
import {useLocalStorage} from "react-use";

declare const window: any;

export const AppHeader = (props: NodeProps) => {
    const pathIdRewards = 'rewards';
    const pathIdErrors = 'errors';

    const isMobile = useBreakpointValue([true, false]);
    const defaultSavedAddresses: Array<string> = [];
    const [savedAddresses, setSavedAddresses] = useLocalStorage("savedAddresses", defaultSavedAddresses);
    const [currentAddress, setCurrentAddress] = useLocalStorage("currentAddress", "");
    const node = useContext(NodeContext);

    let {address} = useParams();
    if(address === '' || address === undefined) {
        if(currentAddress !== '') {
            address = currentAddress;
        } else if(savedAddresses !== undefined && savedAddresses.length > 0) {
            setCurrentAddress(savedAddresses[0]);
        }
    }
    let activePath = '';

    const [rpcEndpoint, setRpcEndpoint] = useState("");
    const [hasLoaded, setHasLoaded] = useState(false);
    const statusColor = (props.node.isJailed || !props.node.exists) ? "#FF0000"   : "#2bd950";
    const status = (props.node.isJailed || !props.node.exists) ?
        (props.node.exists ? "Jailed" : "Invalid address") : "Not Jailed";

    const pathElements = window.location.pathname.split('/');
    switch(pathElements[pathElements.length-1]) {
        case pathIdRewards:
            activePath = pathIdRewards;
            break;
        case pathIdErrors:
            activePath = pathIdErrors;
            break;
    }

    const updateBalance = useCallback(() => {
        if(rpcEndpoint === "" || address === "") {
            console.error(`ABORT addr: ${address} rpc: ${rpcEndpoint}`)
            return;
        }

        axios.get(rpcEndpoint)
            .then(async (result) => {
                // console.log("Node status result", result);
                const node: CryptoNode = {
                    exists: result.data.data.address !== "",
                    address: result.data.data.address,
                    balance: result.data.data.balance,
                    chains: result.data.data.chains,
                    isJailed: result.data.data.is_jailed,
                    pubkey: result.data.data.pubkey,
                    stakedBalance: result.data.data.staked_balance,
                }
                node.lastChecked = new Date();
                props.onNodeLoaded(node);
                setCurrentAddress(node.address);
                setHasLoaded(true);
            })
            .catch((err) => {
                console.error(err);
                // node.exists = false;
                // props.onNodeLoaded(node);
                setHasLoaded(true);
            });
    }, [props, address, rpcEndpoint]);

    useEffect(() => {
        if(!hasLoaded) {
            const rpcUrl = `${window._env_.RPC_URL}/node/${address}`;
            setRpcEndpoint(rpcUrl);
            updateBalance();
        }
    }, [address, hasLoaded, props, updateBalance]);


    return (
        <HStack justifyContent={"space-between"}>
            {/* Nav Links */}
            {window.location.pathname === "/" ? ( <Box/> ) : ( <HomeButton alignSelf="flex-start"/> )}
            {(node.address && !isMobile) && (
                <>
                    <Button
                        variant={activePath === pathIdRewards ? "outline" : "ghost"}
                        title={"Rewards"}
                        onClick={() => window.location.href=`/node/${props.node.address}/rewards`}
                    >
                        Rewards
                    </Button>
                    <Button
                        variant={activePath === pathIdErrors ? "outline" : "ghost"}
                        title={"Logs"}
                        onClick={() => window.location.href=`/node/${props.node.address}/errors`}
                    >
                        Logs
                    </Button>
                </>
            )}
            <Spacer/>
            <Box>
                <MdBrightness1
                    title={"Node status: " + status}
                    aria-label={"Node status: " + status}
                    fill={statusColor}
                />
            </Box>

            <NodeChooser address={address ?? ''}/>

            {!isMobile && (
                <HStack alignSelf={"flex-start"}>
                    {/* Color Mode */}
                    <ColorModeSwitcher
                        _focus={{boxShadow: "none"}}
                        alignSelf="flex-end"
                    />
                     {/* Source code link */}
                    <IconButton
                        aria-label={"Source code"}
                        icon={(<FaGithub/>)}
                        onClick={() => { window.location.href="https://github.com/itsnoproblem/pokt-calculator"} }
                        _focus={{boxShadow: "none"}}
                        alignSelf="flex-end"
                        size="md"
                        fontSize="lg"
                        variant="ghost"
                        color="current"
                        marginLeft="2"
                    />
                </HStack>
            )}
        </HStack>
    )
}