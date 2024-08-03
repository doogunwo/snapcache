package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

const (
	targetBlockNumber = 1000000
	gethIPCPath       = "/mnt/nvme1n1/ethereum/m2p/geth.ipc"
)

func main() {
	// Geth 프로세스의 PID를 찾습니다.
	startTime := time.Now()
	gethPID, err := findGethPID()
	if err != nil {
		log.Fatalf("Failed to find Geth process: %v", err)
	}
	if gethPID == 0 {
		log.Fatalf("Geth process not found")
	}
	fmt.Printf("Geth PID: %d\n", gethPID)

	// Geth 클라이언트와 연결합니다.
	client, err := rpc.Dial(gethIPCPath)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 시스템 시그널을 처리합니다.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Received shutdown signal, exiting...")
		cancel()
		killProcess(gethPID)
	}()

	// 블록을 모니터링합니다.
	err = monitorBlocks(ctx, client, gethPID)
	if err != nil {
		log.Fatalf("Error monitoring blocks: %v", err)
	}

	fmt.Println("Sync to target block completed.")

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Sync to target block completed in %02d:%02d:%02d.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
}

func findGethPID() (int, error) {
	cmd := exec.Command("pgrep", "-f", "geth")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	pidStr := strings.TrimSpace(out.String())
	if pidStr == "" {
		return 0, nil
	}

	// 여러 PID가 반환된 경우, 첫 번째 PID만 사용합니다.
	pids := strings.Split(pidStr, "\n")
	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func killProcess(pid int) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Failed to find process with PID %d: %v", pid, err)
		return
	}
	err = proc.Kill()
	if err != nil {
		log.Printf("Failed to kill process with PID %d: %v", pid, err)
	} else {
		fmt.Printf("Process with PID %d has been killed.\n", pid)
	}
}

func monitorBlocks(ctx context.Context, client *rpc.Client, gethPID int) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var syncResult map[string]interface{}
			err := client.CallContext(ctx, &syncResult, "eth_syncing")
			if err != nil {
				log.Printf("Failed to get sync status: %v", err)
				continue
			}

			if syncResult == nil {
				fmt.Println("Sync is complete")
				return nil
			}

			currentBlock, ok := new(big.Int).SetString(syncResult["currentBlock"].(string), 0)
			if !ok {
				log.Printf("Failed to parse currentBlock: %v", syncResult["currentBlock"])
				continue
			}

			fmt.Printf("Current block number: %s\n", currentBlock.String())
			if currentBlock.Cmp(big.NewInt(targetBlockNumber)) >= 0 {
				fmt.Println("Target block reached, stopping Geth.")
				killProcess(gethPID)
				return nil
			}
		}
	}
}
