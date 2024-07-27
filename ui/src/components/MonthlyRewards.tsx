import {
    Accordion,
    AccordionButton,
    AccordionIcon,
    AccordionItem,
    AccordionPanel,
    Box,
    Grid,
    GridItem,
    HStack,
    Skeleton,
    Stack,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    useBreakpointValue,
    useColorModeValue
} from "@chakra-ui/react";
import React, {useCallback, useContext, useEffect, useState} from "react";
import {NodeContext} from "../context";
import {MonthlyReward, monthNames} from "../types/monthly-reward";
import {RewardTransaction} from "./RewardTransaction";
import {PieChart} from "./PieChart";
import {getClaims, getHeight} from "../MonitoringService";
import {EVENT_MONTH_CLOSE, EVENT_MONTH_METRICS, EVENT_MONTH_OPEN, EVENT_MONTH_TRANSACTIONS, trackGoal} from "../events";
import {DailyChartStacked} from "./DailyChartStacked";
import {DayOfWeekChart} from "./DayOfWeekChart";

type MonthlyRewardsProps = {
    rewards: MonthlyReward[],
    onRewardsLoaded: (r: MonthlyReward[]) => void,
    isRefreshing: boolean,
    setIsRefreshing: (is: boolean) => void,
}

export const MonthlyRewards = (props: MonthlyRewardsProps) => {
    const [hasLoaded, setHasLoaded] = useState(false);
    const [currentHeight, setCurrentHeight] = useState(0)

    const node = useContext(NodeContext)
    const isMobile = useBreakpointValue([true, false]);
    const headerColor = useColorModeValue('gray.100', 'gray.700')

    const getRewards = useCallback(() => {
        props.setIsRefreshing(true);

        getHeight()
            .then((h) => setCurrentHeight(h))
            .catch((err) => { console.error("ERROR", err); });

        getClaims(node.address).then((months) => {
            props.onRewardsLoaded(months);
        }).catch((err) => {
            console.error("ERROR", err);
        }).finally(() => {
            setHasLoaded(true);
            props.setIsRefreshing(false);
        });
    },[node, props]);

    useEffect(() => {
        if(!hasLoaded) {
            getRewards();
        }
    },  [hasLoaded, getRewards]);

    const bgOdd = useColorModeValue("gray.200", "gray.800");
    const bgEven = useColorModeValue("gray.50", "gray.700");

    const relaysByChain = useCallback((month: MonthlyReward) => {
        let data = [];

        for(let i = 0; i < month.relays_by_chain.length; i++) {
            data[i] = {
                id: month.relays_by_chain[i].name,
                label: month.relays_by_chain[i].name,
                value: month.relays_by_chain[i].num_relays,
            };
        }
        return data.sort((i, j) => {
            return (i.value < j.value) ? 1 : -1;
        });
    }, []);

    interface chartData {
        id: string
        label: string
        value: number
    }

    const getDaysOfWeekData = (month: MonthlyReward) => {
        const daysOfWeekData: chartData[] = [];

        for(let i = 0; i < 7; i++) {
            daysOfWeekData[i] = {
                id: month.days_of_week[i].name.slice(0, 3),
                label: month.days_of_week[i].name,
                value: 0,
            }
        }

        month.transactions.map((tx) => {
            const dow = new Date(tx.time)
            daysOfWeekData[dow.getDay()].value += tx.num_relays;
            return daysOfWeekData;
        })

        return daysOfWeekData;
    }


    return props.isRefreshing ? (
        <Stack w={["100vw", "1280px"]} ml={"auto"} mr={"auto"} mt={2}>

            {Object.keys(monthNames).map((k, i) => {
                i++;
                return (<Skeleton key={i} height={'48px'}/>)
            })}
        </Stack>
    ) : (
        <Accordion
            onChange={(e: number[]) => {
                if(e.length === 0) {
                    trackGoal(EVENT_MONTH_CLOSE);
                } else {
                    trackGoal(EVENT_MONTH_OPEN);
                }
            }}
            allowMultiple={true}
            maxW="100%"
            w={["100%", "1280px"]}
            ml={"auto"} mr={"auto"}
            mt={[2,8]} p={0}
        >

            <AccordionItem
                key={"header"}
                borderTopWidth={0}
                backgroundColor={headerColor}
                borderTopLeftRadius={17}
                borderTopRightRadius={17}
                lineHeight={4}
            >
                <h2>
                    <AccordionButton disabled={true}>
                        <Box flex='1'>
                            <HStack pt={1} pb={1}>
                                <Box pl={[1,8]} w={["200px", "40%"]} textAlign='left'><b>Month</b></Box>
                                { !isMobile && (
                                    <>
                                        <Box
                                            fontSize={"sm"}
                                            w={"15%"}
                                            flexGrow={1}
                                            textAlign={"right"}
                                        >
                                            <b>Relays</b>
                                        </Box>
                                        <Box
                                            fontSize={"sm"}
                                            w={"15%"}
                                            flexGrow={1}
                                            textAlign={"right"}
                                        >
                                            <b>Sessions</b>
                                        </Box>
                                        <Box
                                            fontSize={"sm"}
                                            w={"15%"}
                                            flexGrow={1}
                                            textAlign={"right"}
                                        >
                                            <b>pokt / sess</b>
                                        </Box>
                                    </>
                                ) }
                                <Box fontSize={"sm"} minW={["150px", "15%"]} pr={[2,8]} flexGrow={1} textAlign='right'>
                                    <b>pokt</b>
                                </Box>
                                <Box w={"20px"}/>
                            </HStack>
                        </Box>
                    </AccordionButton>
                </h2>
            </AccordionItem>

            {props.rewards.map((month: MonthlyReward, i) => {
                const relays = relaysByChain(month)
                return (
                    <AccordionItem key={i}>
                        <h2>
                            <AccordionButton>
                                <Box flex='1'>
                                    <HStack pt={1} pb={1}>
                                        <Box pl={[1,8]} w={["200px", "40%"]} textAlign='left'>{monthNames[month.month]} {month.year}</Box>
                                        {/*<Spacer/>*/}
                                        { !isMobile && (
                                            <>
                                                <Box
                                                    fontSize={"sm"}
                                                    w={"15%"} flexGrow={1} textAlign={"right"}
                                                >
                                                    {month.num_relays.toLocaleString()}
                                                </Box>
                                                <Box fontSize={"sm"} w={"15%"} flexGrow={1} textAlign={"right"}>
                                                    {month.transactions.length.toLocaleString()}
                                                </Box>
                                                <Box fontSize={"sm"} w={"15%"} flexGrow={1} textAlign={"right"}>
                                                    {(month.pokt_amount/month.transactions.length).toFixed(2)}
                                                </Box>
                                            </>
                                        ) }
                                        <Box fontSize={"sm"} minW={["150px", "15%"]} pr={[2,8]} flexGrow={1} textAlign='right'>
                                            {Number(month.pokt_amount).toFixed(4)}
                                        </Box>
                                    </HStack>
                                </Box>
                                <AccordionIcon />
                            </AccordionButton>
                        </h2>
                        <AccordionPanel pb={4}>
                            <Tabs onChange={(i) => {
                                console.log(`tab ${i}`)
                                if(i === 0) {
                                    trackGoal(EVENT_MONTH_TRANSACTIONS);
                                } else if(i === 1) {
                                    trackGoal(EVENT_MONTH_METRICS);
                                }
                            }
                            }>
                                <TabList>
                                    <Tab onChange={(e) => { console.log(e); }}>Transactions</Tab>
                                    <Tab onChange={(e) => { console.log(e); }}>Metrics</Tab>
                                </TabList>
                                <TabPanels>
                                    <TabPanel p={0}>
                                        <Grid templateColumns={['repeat(5, auto)', 'repeat(10, auto)']} fontFamily={"monospace"} fontSize={"xs"} p={[1, 5]}>
                                            <GridItem padding={2} fontWeight={900} align="left" pl={4}>Height</GridItem>
                                            <GridItem padding={2} fontWeight={900}>Time</GridItem>
                                            {!isMobile && (<GridItem padding={2} fontWeight={900} pr={4} align={"right"}>Relays</GridItem>)}
                                            <GridItem padding={2} fontWeight={900} pr={4} pl={4} align={"left"}>Chain</GridItem>
                                            {!isMobile && (<GridItem padding={2} fontWeight={900} pr={4} align={"right"}>Stake Weight</GridItem>)}
                                            {!isMobile && (<GridItem padding={2} fontWeight={900} pr={4} align={"right"}>Rate / relay</GridItem>)}
                                            <GridItem padding={2} fontWeight={900} pr={4} align={"right"} >Amount</GridItem>
                                            {!isMobile && (<GridItem padding={2} fontWeight={900}>App Pubkey</GridItem>)}
                                            {!isMobile && (<GridItem padding={2} fontWeight={900}>Hash</GridItem>)}
                                            <GridItem padding={2} fontWeight={900}>Confirmed</GridItem>
                                            {month.transactions.slice(0).reverse().map((tx, j) => {
                                                const rowColor = (j % 2 === 0) ? bgEven : bgOdd;
                                                return (
                                                    <RewardTransaction
                                                        currentHeight={currentHeight}
                                                        key={tx.hash}
                                                        tx={tx}
                                                        color={rowColor}
                                                    />
                                                )
                                            })}
                                        </Grid>
                                    </TabPanel>
                                    <TabPanel minHeight={"400px"}>
                                        <Box w={"100%"} margin={"auto"} mb={20}>
                                            <DailyChartStacked txs={month.transactions}/>
                                        </Box>
                                        <Stack direction={["column", "row"]}>
                                            <Box w={["100%", "100%"]} height={"400px"} color={"gray.50"}>
                                                <PieChart data={relays}/>
                                            </Box>
                                            <Box w={"100%"} height={"400px"}>
                                                <DayOfWeekChart data={getDaysOfWeekData(month)}/>
                                            </Box>
                                        </Stack>

                                    </TabPanel>
                                </TabPanels>
                            </Tabs>
                        </AccordionPanel>
                    </AccordionItem>
                )
            })}
        </Accordion>
    )
}
