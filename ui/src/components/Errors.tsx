import {NodeContext} from "../node-context";
import {useContext, useEffect, useState} from "react";
import axios from "axios";
import {PocketError} from "../types/error";
import {
    Box,
    Flex,
    FormControl,
    HStack,
    IconButton,
    Select, Skeleton, Stack,
    Table,
    Tbody,
    Td,
    Text,
    Th,
    Thead,
    Tr
} from "@chakra-ui/react";
import {ArrowLeftIcon, ArrowRightIcon} from "@chakra-ui/icons";

export const Errors = () => {
    const node = useContext(NodeContext)
    const defaultErrors: PocketError[] = [];
    const [errors, setErrors] = useState(defaultErrors);
    const [page, setPage] = useState(1);
    const [chain, setChain] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const resultsPerPage = 25;

    useEffect(() => {
        const offset = (page * resultsPerPage) - resultsPerPage;
        if(chain === '') {
            return;
        }

        setIsLoading(true);

        console.log(`Get chain ${chain}`);
        const rpcUrl = `https://metrics-api.portal.pokt.network:3000/error?and=(nodepublickey.eq.${node.pubkey},blockchain.eq.${chain})&limit=${resultsPerPage}&offset=${offset}`;
        console.log(rpcUrl);

        axios.get(rpcUrl)
            .then(async (response) => {
                if(response.status !== 200) {
                    throw new Error(`got status [${response.status}] from metrics-api`)
                }
                setErrors(response.data as PocketError[]);
            })
            .finally(() => {
                setIsLoading(false);
            });
    }, [node, chain, page]);

    const switchChains = (e: any) => {
        const ch = e.target.value;
        console.log("switch chains", ch);
        setPage(1);
        setChain(ch);
    }

    const SkeletonRows = () => {
        let rows = [];
        for(let i=0; i < resultsPerPage; i++) {
            rows.push(i)
        }
        return (
            <Stack>
                {rows.map(() => (<Skeleton height={'53px'}/>))}
            </Stack>
        )
    }

    return (
        <Box>
            <Flex pt={10} pb={10} pl={40} pr={40}>
                <FormControl>
                    <Select onChange={switchChains} isFullWidth={false} placeholder={"Select a chain"}>
                        {node.chains.map((ch) => {
                            const selected = (chain === ch.id) ? "selected" : "";
                            return (<option value={ch.id} {...selected}>{ch.name}</option>)
                        })}
                    </Select>
                </FormControl>
            </Flex>
            <Flex pt={1} pb={1} pl={40} pr={40} w={"100%"} align={"center"}>
                { chain && (
                    <HStack margin={"auto"} mb={4}>
                        <Box>
                            <IconButton disabled={(page <= 1)} aria-label={"back"} icon={(<ArrowLeftIcon/>)} onClick={() => { setPage(page - 1); }}/>
                        </Box>
                        <Box w={100} align={"center"}>{page}</Box>
                        <Box>
                            <IconButton disabled={(errors.length < resultsPerPage)} aria-label={"next"} icon={(<ArrowRightIcon/>)} onClick={() => { setPage(page + 1); }}/>
                        </Box>
                    </HStack>
                )}
            </Flex>
            { isLoading && (<SkeletonRows/>) }
            {(errors.length === 0 && !isLoading) && (
                <Box align={"center"} m={"auto"}>No results</Box>
            )}
            { (errors.length > 0  && !isLoading) && (
                <Table variant={"striped"} w={"100%"}>
                    <Thead>
                        <Th w={"250px"}>Timestamp</Th>
                        <Th w={"100px"}>Chain</Th>
                        <Th w={"200px"}>Method</Th>
                        <Th w={"80px"}>Code</Th>
                        <Th w={"auto"}>Message</Th>
                    </Thead>
                    <Tbody>
                    {errors.map((err, index) => {
                        return (
                            <Tr>
                                <Td w={"250px"}>{(new Date(err.timestamp)).toLocaleString()}</Td>
                                <Td w={"100px"}>{err.blockchain}</Td>
                                <Td w="250px" maxW={"250px"} overflow={"hidden"} isTruncated={true}>{err.method}</Td>
                                <Td w={"80px"}>{err.code ?? '-'}</Td>
                                <Td w={"auto"}>{err.message}</Td>
                            </Tr>
                        )
                    })}
                    </Tbody>
                </Table>
            )}
        </Box>
    )
}