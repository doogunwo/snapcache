package main

import (
    "context"
    "log"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
    ipcpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"
    // 이더리움 클라이언트에 연결
    client, err := ethclient.Dial(ipcpath)
    if err != nil {
        log.Fatalf("Failed to connect to the Ethereum client: %v", err)
    }

    // 최신 블록 헤더 가져오기
    header, err := client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to get the latest block header: %v", err)
    }

    // 상태 루트 해시 (baseRoot) 가져오기
    baseRoot := header.Root

    // 상태 루트 해시 출력
    log.Printf("State Root Hash: %s", baseRoot.Hex())
}

