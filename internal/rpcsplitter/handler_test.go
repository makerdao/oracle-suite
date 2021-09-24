//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package rpcsplitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var blockWithHashesResp = json.RawMessage(`
	{
		"difficulty": "0x2d50ba175407",
		"extraData": "0xe4b883e5bda9e7a59ee4bb99e9b1bc",
		"gasLimit": "0x47e7c4",
		"gasUsed": "0x5208",
		"hash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"miner": "0x61c808d82a3ac53231750dadc13c777b59310bd9",
		"mixHash": "0xc38853328f753c455edaa4dfc6f62a435e05061beac136c13dbdcd0ff38e5f40",
		"nonce": "0x3b05c6d5524209f1",
		"number": "0x1e8480",
		"parentHash": "0x57ebf07eb9ed1137d41447020a25e51d30a0c272b5896571499c82c33ecb7288",
		"receiptsRoot": "0x84aea4a7aad5c5899bd5cfc7f309cc379009d30179316a2a7baa4a2ea4a438ac",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x28a",
		"stateRoot": "0x96dbad955b166f5119793815c36f11ffa909859bbfeb64b735cca37cbf10bef1",
		"timestamp": "0x57a1118a",
		"totalDifficulty": "0x262c34a6fd1268f6c",
		"transactions": [
		  "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef"
		],
		"transactionsRoot": "0xb31f174d27b99cdae8e746bd138a01ce60d8dd7b224f7c60845914def05ecc58",
		"uncles": [
			"0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6"
		]
	}
`)

var blockWithObjectsResp = json.RawMessage(`
	{
		"difficulty": "0x2d50ba175407",
		"extraData": "0xe4b883e5bda9e7a59ee4bb99e9b1bc",
		"gasLimit": "0x47e7c4",
		"gasUsed": "0x5208",
		"hash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"miner": "0x61c808d82a3ac53231750dadc13c777b59310bd9",
		"mixHash": "0xc38853328f753c455edaa4dfc6f62a435e05061beac136c13dbdcd0ff38e5f40",
		"nonce": "0x3b05c6d5524209f1",
		"number": "0x1e8480",
		"parentHash": "0x57ebf07eb9ed1137d41447020a25e51d30a0c272b5896571499c82c33ecb7288",
		"receiptsRoot": "0x84aea4a7aad5c5899bd5cfc7f309cc379009d30179316a2a7baa4a2ea4a438ac",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x28a",
		"stateRoot": "0x96dbad955b166f5119793815c36f11ffa909859bbfeb64b735cca37cbf10bef1",
		"timestamp": "0x57a1118a",
		"totalDifficulty": "0x262c34a6fd1268f6c",
		"transactions": [
			{
				"blockHash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
				"blockNumber": "0x1e8480",
				"from": "0x32be343b94f860124dc4fee278fdcbd38c102d88",
				"gas": "0x51615",
				"gasPrice": "0x6fc23ac00",
				"hash": "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef",
				"input": "0x",
				"nonce": "0x1efc5",
				"to": "0x104994f45d9d697ca104e5704a7b77d7fec3537c",
				"transactionIndex": "0x0",
				"value": "0x821878651a4d70000",
				"v": "0x1b",
				"r": "0x51222d91a379452395d0abaff981af4cfcc242f25cfaf947dea8245a477731f9",
				"s": "0x3a997c910b4701cca5d933fb26064ee5af7fe3236ff0ef2b58aa50b25aff8ca5"
			}
		],
		"transactionsRoot": "0xb31f174d27b99cdae8e746bd138a01ce60d8dd7b224f7c60845914def05ecc58",
		"uncles": [
			"0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6"
		]
	}
`)

var transaction1Resp = json.RawMessage(`
	{
		"hash": "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
		"blockHash": "0x1d59ff54b1eb26b013ce3cb5fc9dab3705b415a67127a003c3e61eb445bb8df2",
		"blockNumber": "0x5daf3b",
		"from": "0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
		"gas": "0xc350",
		"gasPrice": "0x4a817c800",
		"input": "0x68656c6c6f21",
		"nonce": "0x15",
		"r": "0x1b5e176d927f8e9ab405058b2d2457392da3e20f328b16ddabcebc33eaac5fea",
		"s": "0x4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c",
		"to": "0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
		"transactionIndex": "0x41",
		"v": "0x25",
		"value": "0xf3dbb76162000"
	}
`)

var transaction2Resp = json.RawMessage(`
	{
		"blockHash": "0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6",
		"blockNumber": "0x1e8480",
		"from": "0x32be343b94f860124dc4fee278fdcbd38c102d88",
		"gas": "0x51615",
		"gasPrice": "0x6fc23ac00",
		"hash": "0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef",
		"input": "0x",
		"nonce": "0x1efc5",
		"to": "0x104994f45d9d697ca104e5704a7b77d7fec3537c",
		"transactionIndex": "0x0",
		"value": "0x821878651a4d70000",
		"v": "0x1b",
		"r": "0x51222d91a379452395d0abaff981af4cfcc242f25cfaf947dea8245a477731f9",
		"s": "0x3a997c910b4701cca5d933fb26064ee5af7fe3236ff0ef2b58aa50b25aff8ca5"
	}
`)

var transactionReceipt1Resp = json.RawMessage(`
	{
		"transactionHash": "0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b",
		"blockHash": "0x8243343df08b9751f5ca0c5f8c9c0460d8a9b6351066fae0acbd4d3e776de8bb",
		"blockNumber": "0x429d3b",
		"contractAddress": null,
		"cumulativeGasUsed": "0x64b559",
		"from": "0x00b46c2526e227482e2ebb8f4c69e4674d262e75",
		"gasUsed": "0xcaac",
		"logs": [
			{
				"blockHash": "0x8243343df08b9751f5ca0c5f8c9c0460d8a9b6351066fae0acbd4d3e776de8bb",
				"address": "0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907",
				"logIndex": "0x56",
				"data": "0x000000000000000000000000000000000000000000000000000000012a05f200",
				"removed": false,
				"topics": [
					"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
					"0x00000000000000000000000000b46c2526e227482e2ebb8f4c69e4674d262e75",
					"0x00000000000000000000000054a2d42a40f51259dedd1978f6c118a0f0eff078"
				],
				"blockNumber": "0x429d3b",
				"transactionIndex": "0xac",
				"transactionHash": "0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b"
			}
		],
		"logsBloom": "0x00000000040000000000000000000000000000000000000000000000000000080000000010000000000000000000000000000000000040000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000010100000000000000000000000000004000000000000200000000000000000000000000000000000000000000",
		"root": "0x3ccba97c7fcc7e1636ce2d44be1a806a8999df26eab80a928205714a878d5114",
		"status": null,
		"to": "0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907",
		"transactionIndex": "0xac"
	}
`)

func Test_RPC_BlockNumber(t *testing.T) {
	t.Run("median-in-range", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_blockNumber").
			mockClientCall(0, `0x4`, "eth_blockNumber").
			mockClientCall(1, `0x5`, "eth_blockNumber").
			mockClientCall(2, `0x6`, "eth_blockNumber").
			expectedResult(`0x4`).
			run()
	})
	t.Run("median-outside-range", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_blockNumber").
			mockClientCall(0, `0x1`, "eth_blockNumber").
			mockClientCall(1, `0x5`, "eth_blockNumber").
			mockClientCall(2, `0x6`, "eth_blockNumber").
			expectedResult(`0x5`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_blockNumber").
			mockClientCall(0, `0x3`, "eth_blockNumber").
			mockClientCall(1, `0x4`, "eth_blockNumber").
			mockClientCall(2, errors.New("error#1"), "eth_blockNumber").
			expectedResult(`0x3`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_blockNumber").
			mockClientCall(0, `0x3`, "eth_blockNumber").
			mockClientCall(1, errors.New("error#1"), "eth_blockNumber").
			mockClientCall(2, errors.New("error#2"), "eth_blockNumber").
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
}

func Test_RPC_GetBlockByHash(t *testing.T) {
	hash := newHash("0xc0f4906fea23cf6f3cce98cb44e8e1449e455b28d684dfa9ff65426495584de6")
	t.Run("with-hashes", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByHash", hash, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(2, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			expectedResult(blockWithHashesResp).
			run()
	})
	t.Run("with-objects", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByHash", hash, true).
			mockClientCall(0, blockWithObjectsResp, "eth_getBlockByHash", hash, true).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByHash", hash, true).
			mockClientCall(2, blockWithObjectsResp, "eth_getBlockByHash", hash, true).
			expectedResult(blockWithObjectsResp).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByHash", hash, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(2, errors.New("error#1"), "eth_getBlockByHash", hash, false).
			expectedResult(blockWithHashesResp).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByHash", hash, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(1, errors.New("error#1"), "eth_getBlockByHash", hash, false).
			mockClientCall(2, errors.New("error#2"), "eth_getBlockByHash", hash, false).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getBlockByHash", hash, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByHash", hash, false).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByHash", hash, false).
			expectedError("").
			run()
	})
}

func Test_RPC_GetBlockByNumber(t *testing.T) {
	number := newNumber("0x1e8480")
	t.Run("with-hashes", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByNumber", number, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(2, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			expectedResult(blockWithHashesResp).
			run()
	})
	t.Run("with-objects", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByNumber", number, true).
			mockClientCall(0, blockWithObjectsResp, "eth_getBlockByNumber", number, true).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByNumber", number, true).
			mockClientCall(2, blockWithObjectsResp, "eth_getBlockByNumber", number, true).
			expectedResult(blockWithObjectsResp).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByNumber", number, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(1, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(2, errors.New("error#1"), "eth_getBlockByNumber", number, false).
			expectedResult(blockWithHashesResp).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockByNumber", number, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(1, errors.New("error#1"), "eth_getBlockByNumber", number, false).
			mockClientCall(2, errors.New("error#2"), "eth_getBlockByNumber", number, false).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getBlockByNumber", number, false).
			mockClientCall(0, blockWithHashesResp, "eth_getBlockByNumber", number, false).
			mockClientCall(1, blockWithObjectsResp, "eth_getBlockByNumber", number, false).
			expectedError("").
			run()
	})
}

func Test_RPC_GetTransactionByHash(t *testing.T) {
	hash := newHash("0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionByHash", hash).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(1, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(2, transaction1Resp, "eth_getTransactionByHash", hash).
			expectedResult(transaction1Resp).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionByHash", hash).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(1, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionByHash", hash).
			expectedResult(transaction1Resp).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionByHash", hash).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionByHash", hash).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionByHash", hash).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getTransactionByHash", hash).
			mockClientCall(0, transaction1Resp, "eth_getTransactionByHash", hash).
			mockClientCall(1, transaction2Resp, "eth_getTransactionByHash", hash).
			expectedError("").
			run()
	})
}

func Test_RPC_GetTransactionCount(t *testing.T) {
	addr := newAddress("0xc94770007dda54cF92009BFF0dE90c06F603a09f")
	bn := newNumber("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionCount", addr, bn).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", addr, bn).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", addr, bn).
			mockClientCall(2, `0x5`, "eth_getTransactionCount", addr, bn).
			expectedResult(`0x5`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionCount", addr, bn).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", addr, bn).
			mockClientCall(1, `0x5`, "eth_getTransactionCount", addr, bn).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionCount", addr, bn).
			expectedResult(`0x5`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionCount", addr, bn).
			mockClientCall(0, `0x5`, "eth_getTransactionCount", addr, bn).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionCount", addr, bn).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionCount", addr, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getTransactionCount", addr, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, `0x5`, "eth_getTransactionCount", addr, bn).
			expectedResult(`0x5`).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getTransactionCount", addr, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, `0x5`, "eth_getTransactionCount", addr, bn).
			expectedResult(`0x5`).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getTransactionCount", addr, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_GetTransactionReceipt(t *testing.T) {
	hash := newHash("0xab059a62e22e230fe0f56d8555340a29b2e9532360368f810595453f6fdd213b")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionReceipt", hash).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			mockClientCall(1, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			mockClientCall(2, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			expectedResult(transactionReceipt1Resp).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionReceipt", hash).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			mockClientCall(1, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			mockClientCall(2, errors.New("error#1"), "eth_getTransactionReceipt", hash).
			expectedResult(transactionReceipt1Resp).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionReceipt", hash).
			mockClientCall(0, transactionReceipt1Resp, "eth_getTransactionReceipt", hash).
			mockClientCall(1, errors.New("error#1"), "eth_getTransactionReceipt", hash).
			mockClientCall(2, errors.New("error#2"), "eth_getTransactionReceipt", hash).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
}

func Test_RPC_GetBlockTransactionCountByHash(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockTransactionCountByHash").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_GetBlockTransactionCountByNumber(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBlockTransactionCountByNumber").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_GetTransactionByBlockHashAndIndex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionByBlockHashAndIndex").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_GetTransactionByBlockNumberAndIndex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getTransactionByBlockNumberAndIndex").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_SendRawTransaction(t *testing.T) {
	tx := newBytes("0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675")
	hash1 := newHash("0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d15273310")
	hash2 := newHash("0xc55e2b90168af6972193c1f86fa4d7d7b31a29c156665d15b9cd48618b5177ef")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_sendRawTransaction", tx).
			mockClientCall(0, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(1, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(2, hash1, "eth_sendRawTransaction", tx).
			expectedResult(hash1).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_sendRawTransaction", tx).
			mockClientCall(0, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(1, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(2, errors.New("error#1"), "eth_sendRawTransaction", tx).
			expectedResult(hash1).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_sendRawTransaction", tx).
			mockClientCall(0, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(1, errors.New("error#1"), "eth_sendRawTransaction", tx).
			mockClientCall(2, errors.New("error#2"), "eth_sendRawTransaction", tx).
			expectedResult(hash1).
			run()
	})
	t.Run("all-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_sendRawTransaction", tx).
			mockClientCall(0, errors.New("error#1"), "eth_sendRawTransaction", tx).
			mockClientCall(1, errors.New("error#2"), "eth_sendRawTransaction", tx).
			mockClientCall(2, errors.New("error#3"), "eth_sendRawTransaction", tx).
			expectedError("error#1").
			expectedError("error#2").
			expectedError("error#3").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_sendRawTransaction", tx).
			mockClientCall(0, hash1, "eth_sendRawTransaction", tx).
			mockClientCall(1, hash2, "eth_sendRawTransaction", tx).
			expectedError("").
			run()
	})
}

func Test_RPC_GetBalance(t *testing.T) {
	addr := newAddress("0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907")
	balance := newNumber("0x100000000000")
	bn := newNumber("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBalance", addr, bn).
			mockClientCall(0, balance, "eth_getBalance", addr, bn).
			mockClientCall(1, balance, "eth_getBalance", addr, bn).
			mockClientCall(2, balance, "eth_getBalance", addr, bn).
			expectedResult(balance).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBalance", addr, bn).
			mockClientCall(0, balance, "eth_getBalance", addr, bn).
			mockClientCall(1, balance, "eth_getBalance", addr, bn).
			mockClientCall(2, errors.New("error#1"), "eth_getBalance", addr, bn).
			expectedResult(balance).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getBalance", addr, bn).
			mockClientCall(0, balance, "eth_getBalance", addr, bn).
			mockClientCall(1, errors.New("error#1"), "eth_getBalance", addr, bn).
			mockClientCall(2, errors.New("error#2"), "eth_getBalance", addr, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getBalance", addr, bn).
			mockClientCall(0, newNumber("0x100000000000"), "eth_getBalance", addr, bn).
			mockClientCall(1, newNumber("0x100000000001"), "eth_getBalance", addr, bn).
			expectedError("").
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getBalance", addr, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, balance, "eth_getBalance", addr, bn).
			expectedResult(balance).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getBalance", addr, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, balance, "eth_getBalance", addr, bn).
			expectedResult(balance).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getBalance", addr, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_GetCode(t *testing.T) {
	addr := newAddress("0xb59f67a8bff5d8cd03f6ac17265c550ed8f33907")
	code1 := newBytes("0x606060405236156100965763ffffffff")
	code2 := newBytes("0x606060405236156100965763ffffff00")
	bn := newNumber("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getCode", addr, bn).
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			mockClientCall(1, code1, "eth_getCode", addr, bn).
			mockClientCall(2, code1, "eth_getCode", addr, bn).
			expectedResult(code1).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getCode", addr, bn).
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			mockClientCall(1, code1, "eth_getCode", addr, bn).
			mockClientCall(2, errors.New("error#1"), "eth_getCode", addr, bn).
			expectedResult(code1).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getCode", addr, bn).
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			mockClientCall(1, errors.New("error#1"), "eth_getCode", addr, bn).
			mockClientCall(2, errors.New("error#2"), "eth_getCode", addr, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getCode", addr, bn).
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			mockClientCall(1, code2, "eth_getCode", addr, bn).
			expectedError("").
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getCode", addr, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			expectedResult(code1).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getCode", addr, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, code1, "eth_getCode", addr, bn).
			expectedResult(code1).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getBalance", addr, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_GetStorageAt(t *testing.T) {
	addr := newAddress("0xc94770007dda54cF92009BFF0dE90c06F603a09f")
	pos := newNumber("0x0")
	bn := newNumber("0x10")
	val1 := newHash("0x0000000000000000000000000000000000000000000000000000000000000100")
	val2 := newHash("0x0000000000000000000000000000000000000000000000000000000000000200")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(1, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(2, val1, "eth_getStorageAt", addr, pos, bn).
			expectedResult(val1).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(1, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(2, errors.New("error#1"), "eth_getStorageAt", addr, pos, bn).
			expectedResult(val1).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(1, errors.New("error#1"), "eth_getStorageAt", addr, pos, bn).
			mockClientCall(2, errors.New("error#2"), "eth_getStorageAt", addr, pos, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			mockClientCall(1, val2, "eth_getStorageAt", addr, pos, bn).
			expectedError("").
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getStorageAt", addr, pos, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			expectedResult(val1).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getStorageAt", addr, pos, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, val1, "eth_getStorageAt", addr, pos, bn).
			expectedResult(val1).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_getBalance", addr, pos, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_Accounts(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_accounts").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_GetProof(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getProof").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_Call(t *testing.T) {
	call := newJSON(`
		{
			"from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
			"to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
			"gas": "0x76c0",
			"gasPrice": "0x9184e72a000",
			"value": "0x9184e72a",
			"data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
		}
	`)
	bn := newNumber("0x10")
	res1 := newBytes("0x1")
	res2 := newBytes("0x2")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_call", call, bn).
			mockClientCall(0, res1, "eth_call", call, bn).
			mockClientCall(1, res1, "eth_call", call, bn).
			mockClientCall(2, res1, "eth_call", call, bn).
			expectedResult(res1).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_call", call, bn).
			mockClientCall(0, res1, "eth_call", call, bn).
			mockClientCall(1, res1, "eth_call", call, bn).
			mockClientCall(2, errors.New("error#1"), "eth_call", call, bn).
			expectedResult(res1).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_call", call, bn).
			mockClientCall(0, res1, "eth_call", call, bn).
			mockClientCall(1, errors.New("error#1"), "eth_call", call, bn).
			mockClientCall(2, errors.New("error#2"), "eth_call", call, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_call", call, bn).
			mockClientCall(0, res1, "eth_call", call, bn).
			mockClientCall(1, res2, "eth_call", call, bn).
			expectedResult(res1).
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_call", call, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, res1, "eth_call", call, bn).
			expectedResult(res1).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_call", call, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, res1, "eth_call", call, bn).
			expectedResult(res1).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_call", call, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_GetLogs(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_getLogs").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_ProtocolVersion(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_protocolVersion").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_GasPrice(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_gasPrice").
			mockClientCall(0, `0x1`, "eth_gasPrice").
			mockClientCall(1, `0x5`, "eth_gasPrice").
			mockClientCall(2, `0x6`, "eth_gasPrice").
			expectedResult(`0x5`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_gasPrice").
			mockClientCall(0, `0x3`, "eth_gasPrice").
			mockClientCall(1, `0x4`, "eth_gasPrice").
			mockClientCall(2, errors.New("error#1"), "eth_gasPrice").
			expectedResult(`0x3`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_gasPrice").
			mockClientCall(0, `0x3`, "eth_gasPrice").
			mockClientCall(1, errors.New("error#1"), "eth_gasPrice").
			mockClientCall(2, errors.New("error#2"), "eth_gasPrice").
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
}

func Test_RPC_EstimateGas(t *testing.T) {
	call := newJSON(`
		{
			"from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
			"to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
			"gas": "0x76c0",
			"gasPrice": "0x9184e72a000",
			"value": "0x9184e72a",
			"data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
		}
	`)
	bn := newNumber("0x10")
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_estimateGas", call, bn).
			mockClientCall(0, `0x1`, "eth_estimateGas", call, bn).
			mockClientCall(1, `0x5`, "eth_estimateGas", call, bn).
			mockClientCall(2, `0x6`, "eth_estimateGas", call, bn).
			expectedResult(`0x5`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_estimateGas", call, bn).
			mockClientCall(0, `0x3`, "eth_estimateGas", call, bn).
			mockClientCall(1, `0x4`, "eth_estimateGas", call, bn).
			mockClientCall(2, errors.New("error#1"), "eth_estimateGas", call, bn).
			expectedResult(`0x3`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_estimateGas", call, bn).
			mockClientCall(0, `0x3`, "eth_estimateGas", call, bn).
			mockClientCall(1, errors.New("error#1"), "eth_estimateGas", call, bn).
			mockClientCall(2, errors.New("error#2"), "eth_estimateGas", call, bn).
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("latest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_estimateGas", call, newBlockNumber("latest")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, `0x4`, "eth_estimateGas", call, bn).
			expectedResult(`0x4`).
			run()
	})
	t.Run("pending-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_estimateGas", call, newBlockNumber("pending")).
			mockClientCall(0, bn, "eth_blockNumber").
			mockClientCall(0, `0x4`, "eth_estimateGas", call, bn).
			expectedResult(`0x4`).
			run()
	})
	t.Run("earliest-block", func(t *testing.T) {
		prepareRPCTest(t, 1, "eth_estimateGas", call, newBlockNumber("earliest")).
			expectedError("").
			run()
	})
}

func Test_RPC_FeeHistory(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_feeHistory").
			expectedError("does not exist").
			run()
	})
}

func Test_RPC_MaxPriorityFeePerGas(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_maxPriorityFeePerGas").
			mockClientCall(0, `0x1`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x5`, "eth_maxPriorityFeePerGas").
			mockClientCall(2, `0x6`, "eth_maxPriorityFeePerGas").
			expectedResult(`0x5`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_maxPriorityFeePerGas").
			mockClientCall(0, `0x3`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, `0x4`, "eth_maxPriorityFeePerGas").
			mockClientCall(2, errors.New("error#1"), "eth_maxPriorityFeePerGas").
			expectedResult(`0x3`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_maxPriorityFeePerGas").
			mockClientCall(0, `0x3`, "eth_maxPriorityFeePerGas").
			mockClientCall(1, errors.New("error#1"), "eth_maxPriorityFeePerGas").
			mockClientCall(2, errors.New("error#2"), "eth_maxPriorityFeePerGas").
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
}

func Test_RPC_ChainId(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_chainId").
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x1`, "eth_chainId").
			mockClientCall(2, `0x1`, "eth_chainId").
			expectedResult(`0x1`).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_chainId").
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x1`, "eth_chainId").
			mockClientCall(2, errors.New("error#1"), "eth_chainId").
			expectedResult(`0x1`).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "eth_chainId").
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, errors.New("error#1"), "eth_chainId").
			mockClientCall(2, errors.New("error#2"), "eth_chainId").
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "eth_chainId").
			mockClientCall(0, `0x1`, "eth_chainId").
			mockClientCall(1, `0x2`, "eth_chainId").
			expectedError("").
			run()
	})
}

func Test_RPC_Version(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		prepareRPCTest(t, 3, "net_version").
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, 1, "net_version").
			mockClientCall(2, 1, "net_version").
			expectedResult(1).
			run()
	})
	t.Run("one-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "net_version").
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, 1, "net_version").
			mockClientCall(2, errors.New("error#1"), "net_version").
			expectedResult(1).
			run()
	})
	t.Run("two-failed", func(t *testing.T) {
		prepareRPCTest(t, 3, "net_version").
			mockClientCall(0, 1, "net_version").
			mockClientCall(1, errors.New("error#1"), "net_version").
			mockClientCall(2, errors.New("error#2"), "net_version").
			expectedError("error#1").
			expectedError("error#2").
			run()
	})
	t.Run("different-responses", func(t *testing.T) {
		prepareRPCTest(t, 2, "net_version").
			mockClientCall(0, `0x1`, "net_version").
			mockClientCall(1, `0x2`, "net_version").
			expectedError("").
			run()
	})
}

func Test_useMostCommon(t *testing.T) {
	tests := []struct {
		in      []interface{}
		minReq  int
		want    interface{}
		wantErr bool
	}{
		{
			in: []interface{}{
				newNumber("0xA"),
			},
			minReq:  1,
			want:    newNumber("0xA"),
			wantErr: false,
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0xA"),
				newNumber("0x1E"),
			},
			minReq:  1,
			want:    newNumber("0xA"),
			wantErr: false,
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0xA"),
				newNumber("0x1E"),
			},
			minReq:  2,
			want:    newNumber("0xA"),
			wantErr: false,
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0xA"),
				newNumber("0xA"),
			},
			minReq:  3,
			want:    newNumber("0xA"),
			wantErr: false,
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				errors.New("error#1"),
				errors.New("error#2"),
			},
			minReq:  1,
			want:    newNumber("0xA"),
			wantErr: false,
		},
		// Fails because there is not enough of the same responses:
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0xA"),
				newNumber("0xF"),
			},
			minReq:  3,
			want:    nil,
			wantErr: true,
		},
		// Fails because there are multiple responses and all of them occurs only once:
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0xF"),
				newNumber("0x1E"),
			},
			minReq:  1,
			want:    nil,
			wantErr: true,
		},
		// Fails because we got an error:
		{
			in: []interface{}{
				errors.New("error#1"),
			},
			minReq:  1,
			want:    nil,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMostCommon(tt.in, tt.minReq)
			if tt.wantErr {
				assert.Error(t, err)
				for _, i := range tt.in {
					if iErr, ok := i.(error); ok {
						assert.Contains(t, err.Error(), iErr.Error())
					}
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, compare(got, tt.want))
			}
		})
	}
}

func Test_useMedian(t *testing.T) {
	tests := []struct {
		in      []interface{}
		minReq  int
		want    *numberType
		wantErr bool
	}{
		{
			in: []interface{}{
				newNumber("0xA"),
			},
			minReq: 1,
			want:   newNumber("0xA"),
		},

		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0x1E"),
			},
			minReq: 1,
			want:   newNumber("0x14"),
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0x1E"),
			},
			minReq: 2,
			want:   newNumber("0x14"),
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0x1E"),
				errors.New("error#1"),
			},
			minReq: 2,
			want:   newNumber("0x14"),
		},
		{
			in: []interface{}{
				newNumber("0x1"),
				newNumber("0xA"),
				newNumber("0x64"),
			},
			minReq: 3,
			want:   newNumber("0xA"),
		},
		// Fails because we got en error:
		{
			in: []interface{}{
				errors.New("error#1"),
			},
			minReq:  1,
			wantErr: true,
		},
		// Fails because there are not enough responses:
		{
			in: []interface{}{
				newNumber("0xA"),
				errors.New("error#1"),
				errors.New("error#2"),
			},
			minReq:  3,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMedian(tt.in, tt.minReq)
			if tt.wantErr {
				assert.Error(t, err)
				for _, i := range tt.in {
					if iErr, ok := i.(error); ok {
						assert.Contains(t, err.Error(), iErr.Error())
					}
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Compare(got))
			}
		})
	}
}

func Test_useMedianDist(t *testing.T) {
	tests := []struct {
		in      []interface{}
		dist    int64
		minReq  int
		want    *numberType
		wantErr bool
	}{
		{
			in: []interface{}{
				newNumber("0xA"),
			},
			dist:   1,
			minReq: 1,
			want:   newNumber("0xA"),
		},
		{
			in: []interface{}{
				newNumber("0x4"),
				newNumber("0x5"),
				newNumber("0x6"),
				newNumber("0x7"),
				newNumber("0x8"),
				newNumber("0x9"),
				newNumber("0xA"),
			},
			dist:   -2,
			minReq: 7,
			want:   newNumber("0x5"),
		},
		{
			in: []interface{}{
				newNumber("0x4"),
				newNumber("0x5"),
				newNumber("0x6"),
				newNumber("0x7"),
				newNumber("0x8"),
				newNumber("0x9"),
				newNumber("0xA"),
			},
			dist:   2,
			minReq: 7,
			want:   newNumber("0x9"),
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0x9"),
				newNumber("0x8"),
				newNumber("0x7"),
				newNumber("0x6"),
				newNumber("0x5"),
				newNumber("0x4"),
			},
			dist:   -2,
			minReq: 7,
			want:   newNumber("0x5"),
		},
		{
			in: []interface{}{
				newNumber("0xA"),
				newNumber("0x9"),
				newNumber("0x8"),
				newNumber("0x7"),
				newNumber("0x6"),
				newNumber("0x5"),
				newNumber("0x4"),
			},
			dist:   2,
			minReq: 7,
			want:   newNumber("0x9"),
		},
		// Fails because there are not enough responses:
		{
			in: []interface{}{
				newNumber("0xA"),
				errors.New("error#1"),
				errors.New("error#2"),
			},
			dist:    1,
			minReq:  3,
			wantErr: true,
		},
		// Fails because we got only an error:
		{
			in: []interface{}{
				errors.New("error#1"),
			},
			dist:    1,
			minReq:  1,
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMedianDist(tt.in, tt.minReq, tt.dist)
			if tt.wantErr {
				assert.Error(t, err)
				for _, i := range tt.in {
					if iErr, ok := i.(error); ok {
						assert.Contains(t, err.Error(), iErr.Error())
					}
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Compare(got))
			}
		})
	}
}
