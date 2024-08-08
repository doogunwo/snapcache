package test

import (
	"testing"
  "path/filepath"
	"github.com/ethereum/go-ethereum/core/rawdb"
  "github.com/ethereum/go-ethereum/common"
) 



func TestRawDB(t *testing.T){
  
  dataDir := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth/chaindata"
  ancientDir := filepath.Join(dataDir, "ancient/chain")
  t.Log(ancientDir)

  
  options := rawdb.OpenOptions{
        Type:              "pebble",       // 사용할 데이터베이스 유형 (leveldb 또는 pebble)
        Directory:         dataDir,  // 데이터베이스 파일이 저장될 디렉토리
        AncientsDirectory: ancientDir,      // 고정 데이터(ancient 데이터)가 저장될 디렉토리
        Namespace:         "namespace",     // 네임스페이스 (데이터베이스 관련 메트릭)
        Cache:             1024 * 1024 * 1024,             // 캐시 크기 (MB 단위)
        Handles:           128,             // 동시에 열 수 있는 파일 핸들의 수
        ReadOnly:          false,           // 읽기 전용 모드 여부
        Ephemeral:         false,           // 휘발성 모드 여부 (파일 시스템 동기화 방지)
    }

  db, err := rawdb.Open(options)
  if err != nil {
    t.Logf("err : %v", err)
  }

  key := common.HexToHash("0x9dc7d4d74f9E738638a6eA837080c4965e83A375")

  data, err := db.Get(key[:])
  if err != nil {
    t.Logf("db Get err : %v", err)
  }
  t.Log("Value :v",data)

}
