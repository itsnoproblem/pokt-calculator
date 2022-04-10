import {Box, Button, FormControl, FormLabel, Input, Spinner, Stack, useDisclosure, useToast} from "@chakra-ui/react";
import React, {useContext, useEffect, useRef, useState} from "react";
import {NodeContext} from "../context";
import {OptionBase, Select} from "chakra-react-select";
import {allChains} from "../MonitoringService";
import {simulateRelays} from "../NodeChecker";
import {Chain} from "../types/chain";
import {RelayTestResponse} from "../types/relay-test-response";
import {RelayResult} from "./RelayResult";

interface ChainOption extends OptionBase {
    label: string;
    value: string;
}

export const RelaySimulator = () => {
    const toast = useToast();
    const node = useContext(NodeContext);
    const {isOpen: testIsRunning, onOpen: startTest, onClose: stopTest} = useDisclosure();
    const [selectedChains, setSelectedChains] = useState((): ChainOption[] => []);
    const [nodeURL, setNodeURL] = useState(node.service_url);
    const emptyTestResponse = {} as Record<string, RelayTestResponse>
    const [relayTestResponse, setRelayTestResponse] = useState(emptyTestResponse);
    const chainPickerRef = useRef(null);

    const runTest = async () => {
        startTest();

        const fail = (err: Error) => {
            toast({
                title: `Failed to run tests`,
                description: `${err}`,
                status: 'error',
                duration: 9000,
                isClosable: true,
            })
            stopTest();
        }

        try {
            const chains: string[] = [];
            selectedChains.map((ch, index) => {
                chains[index] = ch.value
            });

            console.log("runTests", chains);

            return simulateRelays(nodeURL, node.address, chains).then((result) => {
                console.log("Done", result);
                if(result.errorMessage) {
                    fail(result.errorMessage)
                }
                else {
                    setRelayTestResponse(result);
                }

                stopTest();
            }).catch((err) => {
                fail(err);
            });

        } catch(err) {
            fail(err as Error)
        }
    }


    const chainPickerData = (chains?: Chain[]) => {
        if(!chains?.length) {
            chains = allChains
        }
        const data: ChainOption[] = [];

        chains.map((ch) => {
            data.push({
                value: ch.id,
                label: ch.name + " (" + ch.id + ")"
            })
        });

        return data;
    }

    const response = (relayTestResponse as Record<string, RelayTestResponse>);
    useEffect(() => {
        if(selectedChains.length === 0) {
            setSelectedChains(chainPickerData(node.chains))
        }
    })

    return (
        <Box textAlign={"left"}>
            <Stack mt={4}>
                <FormControl mt={2}>
                    <FormLabel>Node URL</FormLabel>
                    <Box>
                    <Input type={"text"}
                           fontFamily={"monospace"}
                           defaultValue={node.service_url}
                           onBlur={(v) => setNodeURL(v.target.value)}
                    />
                    </Box>
                </FormControl>

                <FormControl>
                    <Box mt={4}>
                        <FormLabel>Chains</FormLabel>
                        <Select<ChainOption, true>
                            isMulti
                            name="chains"
                            id="relayTestChains"
                            options={chainPickerData()}
                            placeholder="Select the chains to test..."
                            closeMenuOnSelect={false}
                            defaultValue={chainPickerData(node.chains)}
                            onChange={(t) => setSelectedChains(t as ChainOption[])}
                            ref={chainPickerRef}
                        />
                    </Box>
                </FormControl>
                <FormControl>
                    <Box textAlign={"center"}>
                        <FormLabel>&nbsp;</FormLabel>
                        {testIsRunning ? (
                            <Spinner/>
                        ) : (
                            <Button
                                disabled={testIsRunning}
                                colorScheme={"messenger"}
                                onClick={runTest}
                            >Run Relay Tests</Button>
                        )}
                    </Box>
                </FormControl>
            </Stack>

            <RelayResult relayResponse={response}/>

        </Box>
    )
}
