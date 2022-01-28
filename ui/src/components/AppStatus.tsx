import {
    Badge,
    Box, ButtonSpinner, Flex, Grid, GridItem,
    HStack, IconButton, Popover, PopoverBody, PopoverContent, PopoverTrigger, SimpleGrid, Spinner,
    Stat,
    StatHelpText,
    StatLabel,
    StatNumber,
    useBreakpointValue,
    useDisclosure
} from "@chakra-ui/react";
import React, {useContext} from "react";
import {MonthlyReward} from "../types/monthly-reward";
import {NodeContext} from "../node-context";
import {MdRefresh} from "react-icons/all";
import {SpinnerIcon} from "@chakra-ui/icons";
import {ConnectedChainsBadge} from "./badges/ConnectedChainsBadge";

interface AppStatusProps {
    // onNodeLoaded: (b: CryptoNode) => void,
    rewards: MonthlyReward[],
}

export const AppStatus = (props: AppStatusProps) => {
    const isMobile = useBreakpointValue([true, false])
    const node = useContext(NodeContext);
    const {isOpen: isSpinning, onOpen: startSpinner, onClose: stopSpinner, onToggle: toggleSpinner} = useDisclosure();

    const avgPoktForLastDays = (numDays: number): number => {
        const today = new Date();
        let pastDate = new Date();
        pastDate.setTime(today.getTime() - (numDays * 86400 * 1000));
        if(numDays === 0) {
            pastDate = new Date();
            pastDate.setHours(0, 0, 1);
        }

        let total = 0;
        props.rewards.map((r) => {
            r.transactions.map((t) => {
                const txDate = new Date(t.time);
                if((txDate.getTime() >= pastDate.getTime()) && t.is_confirmed) {
                    total += t.num_proofs;
                }
                return this;
            });
            return this;
        });

        let relays = total;
        if(numDays > 0) {
            relays = (total/numDays);
        }

        return Math.round(relays * 0.0089);
    }

    const sortedRewards = props.rewards;
    const sortedByChain = (sortedRewards[0] !== undefined) ?
        sortedRewards[0].relays_by_chain.sort((a, b) => {
            return (a.num_relays > b.num_relays) ? -1 : 1;
        }) : [];

    return(
        <>
            <HStack mt={4} ml={'auto'} mr={'auto'} p={0}>
                {!isMobile && (
                    <>
                        <Box  p={5} minWidth={"185px"} borderWidth={1} borderRadius={20} borderColor={"gray.50"}>
                            <Stat align={"center"}>
                                <StatLabel>Top Chain This Month</StatLabel>
                                <StatNumber>{sortedByChain[0]?.name}</StatNumber>
                                <StatHelpText>{sortedByChain[0]?.num_relays?.toLocaleString()} relays</StatHelpText>
                            </Stat>
                        </Box>
                        <Box  p={5} minWidth={"185px"}>
                            <Stat align={"center"}>
                                <StatLabel>10 Day Average</StatLabel>
                                <StatNumber>{avgPoktForLastDays(10)}</StatNumber>
                                <StatHelpText>POKT per day</StatHelpText>
                            </Stat>
                        </Box>
                        <Box borderWidth={1} borderRadius={20} p={5} minWidth={"185px"} borderColor={"gray.50"}>
                            <Stat align={"center"}>
                                <StatLabel>30 Day Average</StatLabel>
                                <StatNumber>{avgPoktForLastDays(30)}</StatNumber>
                                <StatHelpText>POKT per day</StatHelpText>
                            </Stat>
                        </Box>
                        <Box  p={5} minWidth={"185px"}>
                            <Stat align={"center"}>
                                <StatLabel>90 Day Average</StatLabel>
                                <StatNumber>{avgPoktForLastDays(90)}</StatNumber>
                                <StatHelpText>POKT per day</StatHelpText>
                            </Stat>
                        </Box>
                        <Box  p={5} minWidth={"185px"} borderWidth={1} borderRadius={20} borderColor={"gray.50"}>
                            <Stat align={"center"}>
                                <StatLabel>Last 24 hrs</StatLabel>
                                <StatNumber>{avgPoktForLastDays(1) ?? 0}</StatNumber>
                                <StatHelpText>POKT earned</StatHelpText>
                            </Stat>
                        </Box>
                    </>
                )}
            </HStack>
            {(!isMobile && node !== undefined) && (
                <Flex>
                    <Box ml="auto" mr="auto" mb={4} mt={12}>
                        <em>Updated: {node.lastChecked?.toLocaleString()}</em>

                        <IconButton
                            ml={4} mr={4}
                            aria-label={"Refresh"}
                            variant={"ghost"}
                            _focus={{boxShadow: 0}}
                            icon={isSpinning ? (<Spinner size={"xs"}/>) : (<MdRefresh/>)}
                            onClick={toggleSpinner}
                        />

                        <ConnectedChainsBadge/>
                    </Box>
                </Flex>
            )}
        </>

)
}
