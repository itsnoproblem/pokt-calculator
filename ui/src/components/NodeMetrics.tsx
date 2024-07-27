import {
    Box,
    HStack,
    Stat,
    StatHelpText,
    StatLabel,
    StatNumber,
    useBreakpointValue,
    useColorModeValue,
    useDisclosure,
} from "@chakra-ui/react";
import React, {useEffect, useRef, useState} from "react";
import {CryptoNode} from "../types/crypto-node";
import {MonthlyReward, monthNames} from "../types/monthly-reward";
import {
    EVENT_TOGGLE_LIFETIME_AVG,
    EVENT_TOGGLE_LIFETIME_AVG_PER_SESS,
    EVENT_TOGGLE_TEN_THIRTY_NINETY,
    trackGoal
} from "../events";

interface AppStatusProps {
    rewards: MonthlyReward[],
    onNodeLoaded: (node: CryptoNode) => void,
    onRewardsLoaded: (months: MonthlyReward[]) => void,
    isRefreshing: boolean,
    setIsRefreshing: (is: boolean) => void,
}

interface TimeUnits {
    units: string
    value: number
}

export const NodeMetrics = (props: AppStatusProps) => {

    const statHoverColor = useColorModeValue('cyan.800', 'blue.100');
    const statColor = useColorModeValue('cyan.800', 'gray.50');
    const statWidth = useBreakpointValue(["90%", "185px"]);
    const statHeight = useBreakpointValue(["200px", "auto"]);

    const statFontSize = useBreakpointValue(["6xl", "xl"])

    const statLabelSize = useBreakpointValue(["xl", "md"])
    const statNumberSize = useBreakpointValue(["6xl", "2xl"])
    const statHelpTextSize = useBreakpointValue(["xl", "sm"])

    const { isOpen: showAllTime, onToggle: toggleShowAllTime } = useDisclosure({ defaultIsOpen: false });
    const { isOpen: showAllTimePerSess, onToggle: toggleShowAllTimePerSess } = useDisclosure({ defaultIsOpen: false })
    const [tenThirtyNinetyState, setTenThirtyNinetyState] = useState(0);

    const emptyTimeUnits: TimeUnits = { units: '---', value: -1 }
    const [timeBetweenRewardsLatest, setTimeBetweenRewardsLatest] = useState(emptyTimeUnits);
    const [timeBetweenRewardsAllTime, setTimeBetweenRewardsAllTime] = useState(emptyTimeUnits);
    const [timeSinceReward, setTimeSinceReward] = useState(emptyTimeUnits);
    const [avgPoktPerSessionLatestMonth, setAvgPoktPerSessionLatestMonth] = useState(0);
    const [avgPoktPerSession, setAvgPoktPerSession] = useState(0);
    const [latestMonthName, setLatestMonthName] = useState('');


    const scrollRef = useRef(null);


    const avgPoktForLastDays = (numDays: number): number => {
        const today = new Date();
        let pastDate = new Date();
        pastDate.setTime(today.getTime() - (numDays * 86400 * 1000));
        if(numDays === 0) {
            pastDate = new Date();
            pastDate.setHours(0, 0, 1);
        }

        let totalPOKT = 0;
        props.rewards.map((r) => {
            r.transactions.map((t) => {
                const txDate = new Date(t.time);
                if((txDate.getTime() >= pastDate.getTime()) && t.is_confirmed) {
                    totalPOKT += t.reward.amount;
                }

                return this;
            });
            return this;
        });

        let pokt = totalPOKT;
        if(numDays > 0) {
            pokt = (totalPOKT/numDays);
        }

        return Math.round(pokt);
    }

    const timeUnits = (seconds: number): TimeUnits => {
        let interval = seconds / 86400;
        if (interval > 1) {
            return {
                units: " days",
                value: Math.floor(interval)
            };
        }
        interval = seconds / 3600;
        if (interval > 1) {
            return {
                units: " hours",
                value: Math.floor(interval)
            };
        }
        interval = seconds / 60;
        if (interval > 1) {
            return {
                units: "minutes",
                value: Math.floor(interval)
            };
        }

        return {
            units: "seconds",
            value: Math.floor(seconds)
        };
    }

    const sortedRewards = props.rewards;
    const sortedByChain = (sortedRewards[0] !== undefined) ?
        sortedRewards[0].relays_by_chain.sort((a, b) => {
            return (a.num_relays > b.num_relays) ? -1 : 1;
        }) : [];

    useEffect(() => {

        const timeSince = (test: Date): TimeUnits => {
            const now = Date.now();
            const testSec = test.getTime();
            const seconds = (now - testSec) / 1000;
            return timeUnits(seconds);
        }

        const latestTxTime = () => {
            for(let j = 0; j < props.rewards.length; j++) {
                const txs = props.rewards[j].transactions;

                for(let i=txs.length-1; i >= 0; i--) {
                    if(txs[i].is_confirmed) {
                        // console.log("latestTxTime", txs[i].time)
                        return new Date(txs[i].time);
                    };
                }
            }
        }

        let lifetimeNumSessions = 0;
        let lifetimePOKT = 0;
        props.rewards.map((r) => {
            r.transactions.map((t) => {
                lifetimePOKT += t.reward.amount;
                lifetimeNumSessions++;
                return this;
            });
            return this;
        });
        setAvgPoktPerSession(lifetimePOKT/lifetimeNumSessions);

        const latestMonth = sortedRewards[0]?.month;
        const lastRewardDate = latestTxTime();
        setTimeBetweenRewardsLatest(timeUnits(sortedRewards[0]?.avg_sec_between_rewards ?? -1));
        setLatestMonthName(latestMonth ? monthNames[latestMonth] : '---');
        if(latestMonth) {
            setAvgPoktPerSessionLatestMonth((sortedRewards[0].pokt_amount / sortedRewards[0].transactions.length));
        }

        if(lastRewardDate) {
            setTimeSinceReward(timeSince(lastRewardDate));
        }

        let secBetweenRewardsAllTime = 0;
        let numRewardsAllTime = 0;
        sortedRewards.map((r, i) => {
            secBetweenRewardsAllTime += r.total_sec_between_rewards;
            numRewardsAllTime += r.transactions.length;
            return numRewardsAllTime;
        });
        setTimeBetweenRewardsAllTime(timeUnits(secBetweenRewardsAllTime / numRewardsAllTime));

    }, [props, sortedRewards]);

    return(
        <>
            <Box mt={8} mb={8} p={0} pl={4} pr={4}
                 ml={'auto'} mr={'auto'}
                 w={["100vw", 'unset']} h={"100%"}
                 overflow={['scroll', 'unset']}
                 ref={scrollRef}
                 style={ { scrollSnapType: "x mandatory" } }
            >

                <HStack textColor={statColor}>
                    <Box p={5}
                         borderColor={statHoverColor}
                         borderWidth={1}
                         borderRadius={20}
                         minWidth={statWidth} height={statHeight}
                         style={ { scrollSnapAlign: "center"} }
                    >
                        <Stat size={statFontSize} align={"center"}>
                            <StatLabel fontSize={statLabelSize}>Last 24 hrs</StatLabel>
                            <StatNumber fontSize={statNumberSize}>{avgPoktForLastDays(1) ?? 0}</StatNumber>
                            <StatHelpText fontSize={statHelpTextSize}>POKT earned</StatHelpText>
                        </Stat>
                    </Box>
                    <Box  p={5} minWidth={statWidth}
                          height={statHeight}
                          style={ { scrollSnapAlign: "center"} }
                    >
                        <Stat size={statFontSize} align={"center"}>
                            <StatLabel fontSize={statLabelSize}>Last reward</StatLabel>
                            <StatNumber fontSize={statNumberSize}>{timeSinceReward.value > 0 ? timeSinceReward.value : '--'}</StatNumber>
                            <StatHelpText fontSize={statHelpTextSize}>{timeSinceReward.units} ago</StatHelpText>
                        </Stat>
                    </Box>
                    <Box style={ { scrollSnapAlign: "center"} } p={5} minWidth={statWidth} height={statHeight} borderWidth={1} borderRadius={20}
                         borderColor={statHoverColor}
                         _hover={{ borderColor: statHoverColor, color: statHoverColor }}
                         cursor={'pointer'}
                         onClick={() => {
                             trackGoal(EVENT_TOGGLE_LIFETIME_AVG);
                             toggleShowAllTime();
                         }}
                    >
                        <Stat size={statFontSize} _hover={ {color: statHoverColor} } align={"center"}>
                            <StatLabel fontSize={statLabelSize}>
                                {showAllTime ?
                                    (<>Lifetime Avg</>) :
                                    (<>{timeBetweenRewardsLatest.value < 0 ? '---' : latestMonthName} Avg</>)

                                }
                            </StatLabel>
                            <StatNumber fontSize={statNumberSize}>
                                {showAllTime ?
                                    (
                                        <>
                                            {timeBetweenRewardsAllTime.value && timeBetweenRewardsAllTime.value >= 0 ?
                                                `${timeBetweenRewardsAllTime.value.toString()} ${timeBetweenRewardsAllTime.units}` : '---'
                                            }
                                        </>
                                    ) : (
                                        <>
                                            {timeBetweenRewardsLatest.value && timeBetweenRewardsAllTime.value >= 0 ?
                                                timeBetweenRewardsLatest.value.toString() : '--'
                                            }
                                            {timeBetweenRewardsLatest.units}
                                        </>
                                    )
                                }

                            </StatNumber>
                            <StatHelpText fontSize={statHelpTextSize}>between rewards</StatHelpText>
                        </Stat>
                    </Box>
                    <Box p={5} minWidth={statWidth} height={statHeight}
                         _hover={ {color: statHoverColor} }
                         cursor={'pointer'}
                         style={ { scrollSnapAlign: "center"} }
                         onClick={() => {
                             const newState = tenThirtyNinetyState + 1;
                             setTenThirtyNinetyState(newState > 2 ? 0 : newState);
                             trackGoal(EVENT_TOGGLE_TEN_THIRTY_NINETY);
                         }}
                    >
                        {tenThirtyNinetyState === 0 && (
                            <Stat size={statFontSize} align={"center"}>
                                <StatLabel fontSize={statLabelSize}>10 Day Average</StatLabel>
                                <StatNumber fontSize={statNumberSize}>{avgPoktForLastDays(10)}</StatNumber>
                                <StatHelpText fontSize={statHelpTextSize}>POKT per day</StatHelpText>
                            </Stat>
                        )}
                        {tenThirtyNinetyState === 1 && (
                            <Stat size={statFontSize} align={"center"}>
                                <StatLabel fontSize={statLabelSize}>30 Day Average</StatLabel>
                                <StatNumber fontSize={statNumberSize}>{avgPoktForLastDays(30)}</StatNumber>
                                <StatHelpText fontSize={statHelpTextSize}>POKT per day</StatHelpText>
                            </Stat>
                        )}
                        {tenThirtyNinetyState === 2 && (
                            <Stat size={statFontSize} align={"center"}>
                                <StatLabel fontSize={statLabelSize}>90 Day Average</StatLabel>
                                <StatNumber fontSize={statNumberSize}>{avgPoktForLastDays(90)}</StatNumber>
                                <StatHelpText fontSize={statHelpTextSize}>POKT per day</StatHelpText>
                            </Stat>
                        )}

                    </Box>
                    <Box p={5} minWidth={statWidth} height={statHeight}
                         cursor={'pointer'}
                         style={ { scrollSnapAlign: "center"} }
                         onClick={() => {
                             trackGoal(EVENT_TOGGLE_LIFETIME_AVG_PER_SESS);
                             toggleShowAllTimePerSess();
                         }}
                         borderWidth={1}
                         borderRadius={20}
                         borderColor={statHoverColor}
                    >
                        <Stat size={statFontSize} _hover={ {color: statHoverColor} } align={"center"}>
                            <StatLabel fontSize={statLabelSize}>
                                {showAllTimePerSess ?
                                    (<>Lifetime Avg</>) :
                                    (<>{timeBetweenRewardsLatest.value < 0 ? '---' : latestMonthName} Avg</>)

                                }
                            </StatLabel>
                            <StatNumber fontSize={statNumberSize}>
                                {showAllTimePerSess ?
                                    (
                                        <>{avgPoktPerSession >= 0 ? avgPoktPerSession.toFixed(0) : '---'}</>
                                    ) : (
                                        <>{avgPoktPerSessionLatestMonth >= 0 ? avgPoktPerSessionLatestMonth.toFixed(0) : '---'}</>
                                    )
                                }

                            </StatNumber>
                            <StatHelpText fontSize={statHelpTextSize}>POKT per session</StatHelpText>
                        </Stat>
                    </Box>
                    <Box p={5}
                         minWidth={statWidth} height={statHeight}
                         style={ { scrollSnapAlign: "center"} }
                    >
                         <Stat align={"center"}>
                            <StatLabel fontSize={statLabelSize}>Top Chain This Month</StatLabel>
                            <StatNumber fontSize={statNumberSize}>{sortedByChain[0]?.name}</StatNumber>
                            <StatHelpText fontSize={statHelpTextSize}>{sortedByChain[0]?.num_relays?.toLocaleString()} relays</StatHelpText>
                        </Stat>
                    </Box>
                    {/*</HorizontalScroll>*/}
                </HStack>
                {/*)}*/}
            </Box>

        </>

)}
