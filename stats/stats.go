package main

import (
  "fmt"
  "log"
  "time"
  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"
  "github.com/ethereum/go-ethereum/common"  
)

    
// Database 구조체 정의 (생략한 부분 포함)
type Database struct {
    fn string      // filename for reporting
    db *leveldb.DB // LevelDB instance
}

// Stat 함수 정의 (생략한 부분 포함)
func (db *Database) Stat() (string, error) {
    var stats leveldb.DBStats
    if err := db.db.Stats(&stats); err != nil {
        return "", err
    }
    var (
        message       string
        totalRead     int64
        totalWrite    int64
        totalSize     int64
        totalTables   int
        totalDuration time.Duration
    )
    if len(stats.LevelSizes) > 0 {
        message += " Level |   Tables   |    Size(MB)   |    Time(sec)  |    Read(MB)   |   Write(MB)\n" +
            "-------+------------+---------------+---------------+---------------+---------------\n"
        for level, size := range stats.LevelSizes {
            read := stats.LevelRead[level]
            write := stats.LevelWrite[level]
            duration := stats.LevelDurations[level]
            tables := stats.LevelTablesCounts[level]

            if tables == 0 && duration == 0 {
                continue
            }
            totalTables += tables
            totalSize += size
            totalRead += read
            totalWrite += write
            totalDuration += duration
            message += fmt.Sprintf(" %3d   | %10d | %13.5f | %13.5f | %13.5f | %13.5f\n",
                level, tables, float64(size)/1048576.0, duration.Seconds(),
                float64(read)/1048576.0, float64(write)/1048576.0)
        }
        message += "-------+------------+---------------+---------------+---------------+---------------\n"
        message += fmt.Sprintf(" Total | %10d | %13.5f | %13.5f | %13.5f | %13.5f\n",
            totalTables, float64(totalSize)/1048576.0, totalDuration.Seconds(),
            float64(totalRead)/1048576.0, float64(totalWrite)/1048576.0)
        message += "-------+------------+---------------+---------------+---------------+---------------\n\n"
    }
    message += fmt.Sprintf("Read(MB):%.5f Write(MB):%.5f\n", float64(stats.IORead)/1048576.0, float64(stats.IOWrite)/1048576.0)
    message += fmt.Sprintf("BlockCache(MB):%.5f FileCache:%d\n", float64(stats.BlockCacheSize)/1048576.0, stats.OpenedTablesCount)
    message += fmt.Sprintf("MemoryCompaction:%d Level0Compaction:%d NonLevel0Compaction:%d SeekCompaction:%d\n", stats.MemComp, stats.Level0Comp, stats.NonLevel0Comp, stats.SeekComp)
    message += fmt.Sprintf("WriteDelayCount:%d WriteDelayDuration:%s Paused:%t\n", stats.WriteDelayCount, common.PrettyDuration(stats.WriteDelayDuration), stats.WritePaused)
    message += fmt.Sprintf("Snapshots:%d Iterators:%d\n", stats.AliveSnapshots, stats.AliveIterators)
    return message, nil
}

func main() {
    // LevelDB 데이터베이스 열기
    dbPath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth/chaindata" // 적절한 경로로 변경
    options := &opt.Options{}
    db, err := leveldb.OpenFile(dbPath, options)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Database 구조체 초기화
    myDB := &Database{
        fn: dbPath,
        db: db,
    }
    ticker := time.Tick(1*time.Second)
    for range ticker {
      stats,err := myDB.Stat()
      if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(stats)
  }

}

