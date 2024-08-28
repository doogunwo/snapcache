package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"

  "bufio"
  

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	bloomfilter "github.com/holiman/bloomfilter/v2"
	"github.com/holiman/uint256"
)



func cacheload(cachePath string) (*fastcache.Cache){
  
  //load cache 
  loadedCache, err := fastcache.LoadFromFile(cachePath) 
  if err != nil {
    log.Fatalf("Failed to load cache : %v", err)
  }

  return loadedCache
}

func NewBlock(ipcpath string,number *big.Int) (map[common.Hash][]byte, error){

  client, err := rpc.Dial(ipcpath)
  if err != nil {
    return nil, err
  }
  ec := ethclient.NewClient(client)
  
  block, err := ec.BlockByNumber(context.Background(), number)
  if err != nil {
    return nil, err
  }

  accountMap := make(map[common.Hash][]byte)

  for _, tx := range block.Transactions() {
    txHash := tx.Hash()
    tx, _, _:= ec.TransactionByHash(context.Background(), txHash)
    
    var toAddress common.Address 
    if tx.To() != nil {
      toAddress = *tx.To()
    }

    balanceBigInt, err := ec.BalanceAt(context.Background(), toAddress, nil)
    if err != nil {
      return nil, err
    }

    balance := new(uint256.Int) 
    balance.SetFromBig(balanceBigInt) 
    account := types.SlimAccount{
      Balance: balance,
    }
  
    blob, err := rlp.EncodeToBytes(account)
    if err != nil {
      return nil, err
    }

    accountMap[common.BytesToHash(toAddress.Bytes())] = blob   
  }

  return accountMap, nil
}

type diffLayer struct {
	origin *diskLayer // Base disk layer to directly use on bloom misses
	parent snapshot   // Parent snapshot modified by this one, never nil
	memory uint64     // Approximate guess as to how much memory we use

	root  common.Hash // Root hash to which this snapshot diff belongs to
	stale atomic.Bool // Signals that the layer became stale (state progressed)


	destructSet map[common.Hash]struct{}               // Keyed markers for deleted (and potentially) recreated accounts
	accountList []common.Hash                          // List of account for iteration. If it exists, it's sorted, otherwise it's nil
	accountData map[common.Hash][]byte                 // Keyed accounts for direct retrieval (nil means deleted)
	storageList map[common.Hash][]common.Hash          // List of storage slots for iterated retrievals, one per account. Any existing lists are sorted if non-nil
	storageData map[common.Hash]map[common.Hash][]byte // Keyed storage slots for direct retrieval. one per account (nil means deleted)

	diffed *bloomfilter.Filter // Bloom filter tracking all the diffed items up to the disk layer

	lock sync.RWMutex
}


var (

  cacheMisses = uint64(0)
  cachehits = uint64(0)
  
	aggregatorMemoryLimit = uint64(4 * 1024 * 1024)


	aggregatorItemLimit = aggregatorMemoryLimit / 42


	bloomTargetError = 0.02


	bloomSize = math.Ceil(float64(aggregatorItemLimit) * math.Log(bloomTargetError) / math.Log(1/math.Pow(2, math.Log(2))))


	bloomFuncs = math.Round((bloomSize / float64(aggregatorItemLimit)) * math.Log(2))

	bloomDestructHasherOffset = 0
	bloomAccountHasherOffset  = 0
	bloomStorageHasherOffset  = 0
)

func init() {
	// Init the bloom offsets in the range [0:24] (requires 8 bytes)
	bloomDestructHasherOffset = rand.Intn(25)
	bloomAccountHasherOffset = rand.Intn(25)
	bloomStorageHasherOffset = rand.Intn(25)

	// The destruct and account blooms must be different, as the storage slots
	// will check for destruction too for every bloom miss. It should not collide
	// with modified accounts.
	for bloomAccountHasherOffset == bloomDestructHasherOffset {
		bloomAccountHasherOffset = rand.Intn(25)
	}
}


// destructBloomHash is used to convert a destruct event into a 64 bit mini hash.
func destructBloomHash(h common.Hash) uint64 {
	return binary.BigEndian.Uint64(h[bloomDestructHasherOffset : bloomDestructHasherOffset+8])
}

// accountBloomHash is used to convert an account hash into a 64 bit mini hash.
func accountBloomHash(h common.Hash) uint64 {
	return binary.BigEndian.Uint64(h[bloomAccountHasherOffset : bloomAccountHasherOffset+8])
}

// storageBloomHash is used to convert an account hash and a storage hash into a 64 bit mini hash.
func storageBloomHash(h0, h1 common.Hash) uint64 {
	return binary.BigEndian.Uint64(h0[bloomStorageHasherOffset:bloomStorageHasherOffset+8]) ^
		binary.BigEndian.Uint64(h1[bloomStorageHasherOffset:bloomStorageHasherOffset+8])
}

func (dl *diffLayer) Account(hash common.Hash) ([]byte, error) {

	data, err := dl.AccountRLP(hash)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
  return data, nil 
}

func (dl *diffLayer) AccountRLP(hash common.Hash) ([]byte, error) {
	dl.lock.RLock()

	hit := dl.diffed.ContainsHash(accountBloomHash(hash))
	if !hit {
		hit = dl.diffed.ContainsHash(destructBloomHash(hash))
	}
	var origin *diskLayer
	if !hit {
		origin = dl.origin 
	}
	dl.lock.RUnlock()

	if origin != nil {
			return origin.AccountRLP(hash)
	}
	// The bloom filter hit, start poking in the internal maps
	return dl.accountRLP(hash, 0)
}

// accountRLP is an internal version of AccountRLP that skips the bloom filter
// checks and uses the internal maps to try and retrieve the data. It's meant
// to be used if a higher layer's bloom filter hit already.
func (dl *diffLayer) accountRLP(hash common.Hash, depth int) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	// If the account is known locally, return it
	if data, ok := dl.accountData[hash]; ok {
			return data, nil
	}
	// If the account is known locally, but deleted, return it
	if _, ok := dl.destructSet[hash]; ok {	
		return nil, nil
	}
	// Account unknown to this diff, resolve from parent
	if diff, ok := dl.parent.(*diffLayer); ok {
		return diff.accountRLP(hash, depth+1)
	}
	return dl.parent.AccountRLP(hash)
}

func newDiffLayer(parent snapshot, root common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte) *diffLayer {
	// Create the new layer with some pre-allocated data segments
	dl := &diffLayer{
		parent:      parent,
		root:        root,
		destructSet: destructs,
		accountData: accounts,
		storageData: storage,
		storageList: make(map[common.Hash][]common.Hash),
	}
	switch parent := parent.(type) {
	case *diskLayer:
		dl.rebloom(parent)
	case *diffLayer:
		dl.rebloom(parent.origin)
	default:
		panic("unknown parent type")
	}
	// Sanity check that accounts or storage slots are never nil
	for accountHash, blob := range accounts {
		if blob == nil {
			panic(fmt.Sprintf("account %#x nil", accountHash))
		}
		// Determine memory size and track the dirty writes
		dl.memory += uint64(common.HashLength + len(blob))
		}
	for accountHash, slots := range storage {
		if slots == nil {
			panic(fmt.Sprintf("storage %#x nil", accountHash))
		}
		// Determine memory size and track the dirty writes
		for _, data := range slots {
			dl.memory += uint64(common.HashLength + len(data))
				}
	}
	dl.memory += uint64(len(destructs) * common.HashLength)
	return dl
}

func (dl *diffLayer) rebloom(origin *diskLayer) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	// Inject the new origin that triggered the rebloom
	dl.origin = origin

	// Retrieve the parent bloom or create a fresh empty one
	if parent, ok := dl.parent.(*diffLayer); ok {
		parent.lock.RLock()
		dl.diffed, _ = parent.diffed.Copy()
		parent.lock.RUnlock()
	} else {
		dl.diffed, _ = bloomfilter.New(uint64(bloomSize), uint64(bloomFuncs))
	}
	// Iterate over all the accounts and storage slots and index them
	for hash := range dl.destructSet {
		dl.diffed.AddHash(destructBloomHash(hash))
	}
	for hash := range dl.accountData {
		dl.diffed.AddHash(accountBloomHash(hash))
	}
	for accountHash, slots := range dl.storageData {
		for storageHash := range slots {
			dl.diffed.AddHash(storageBloomHash(accountHash, storageHash))
		}
	}
	
}

func (dl *diffLayer) Update(blockRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte) *diffLayer {
	return newDiffLayer(dl, blockRoot, destructs, accounts, storage)
}

type diskLayer struct {
	diskdb ethdb.KeyValueStore // Key-value store containing the base snapshot
	immutableCache  *fastcache.Cache    // Cache to avoid hitting the disk for direct access
  mutableCache    *fastcache.Cache

	root  common.Hash // Root hash of the base snapshot

	lock sync.RWMutex

}

func (dl *diskLayer) Update(blockHash common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte) *diffLayer {
	return newDiffLayer(dl, blockHash, destructs, accounts, storage)
}

type Snapshot interface {
	// Root returns the root hash for which this snapshot was made.

	// Account directly retrieves the account associated with a particular hash in
	// the snapshot slim data format.
	Account(hash common.Hash) ([]byte, error)

	// AccountRLP directly retrieves the account RLP associated with a particular
	// hash in the snapshot slim data format.
	AccountRLP(hash common.Hash) ([]byte, error)

	// Storage directly retrieves the storage data associated with a particular hash,
	// within a particular account.
}


type snapshot interface {
	Snapshot
  Update(blockRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte) *diffLayer

}

type Tree struct {
	diskdb ethdb.KeyValueStore      // Persistent database to store the snapshot
	layers map[common.Hash]snapshot // Collection of all known layers
	lock   sync.RWMutex

	// Test hooks
	onFlatten func() // Hook invoked when the bottom most diff layers are flattened
}

func (t *Tree) Snapshot(blockRoot common.Hash) Snapshot {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.layers[blockRoot]
}

func (t *Tree) Update(blockRoot common.Hash, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte) error {
	
	parent := t.Snapshot(parentRoot)
	snap := parent.(snapshot).Update(blockRoot, destructs, accounts, storage)

	// Save the new snapshot for later
	t.lock.Lock()
	defer t.lock.Unlock()

	t.layers[snap.root] = snap
	return nil
}

func (dl *diskLayer) Account(hash common.Hash) ([]byte, error) {
	data, err := dl.AccountRLP(hash)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	
  return data, nil
}

func (dl *diskLayer) AccountRLP(hash common.Hash) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()
  
  blob, found := dl.immutableCache.HasGet(nil, hash[:])
  if found == true {
    atomic.AddUint64(&cachehits, 1)
    return blob, nil
  }

  blob, found = dl.mutableCache.HasGet(nil, hash[:])
  if found == true {
    atomic.AddUint64(&cachehits, 1)
    return blob, nil
  }

  atomic.AddUint64(&cacheMisses, 1)
  balance := big.NewInt(1234)
  blob, _ = rlp.EncodeToBytes(balance)

  dl.mutableCache.Set(hash[:], blob)
  return blob, nil
}


func cacheLoadLayer() *diskLayer {
  
  cachePath := "test/addrUnder5"
  loadedCache, err := fastcache.LoadFromFile(cachePath)
  if err != nil {
    log.Fatalf("Failed to load cache : %v", err)
  }

  return &diskLayer{
        diskdb          :   memorydb.New(),      // Use the provided ethdb.KeyValueStore
        immutableCache :   loadedCache,// Use the loaded cache
        mutableCache   :  fastcache.New(500 * 1024),
    }
}

func copyDestructs(destructs map[common.Hash]struct{}) map[common.Hash]struct{} {
	copy := make(map[common.Hash]struct{})
	for hash := range destructs {
		copy[hash] = struct{}{}
	}
	return copy
}


func copyStorage(storage map[common.Hash]map[common.Hash][]byte) map[common.Hash]map[common.Hash][]byte {
	copy := make(map[common.Hash]map[common.Hash][]byte)
	for accHash, slots := range storage {
		copy[accHash] = make(map[common.Hash][]byte)
		for slotHash, blob := range slots {
			copy[accHash][slotHash] = blob
		}
	}
	return copy
}

func TestSnapshot_basic(t *testing.T){
  ipcpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc" 
  
  var (
		destructs = make(map[common.Hash]struct{})
		accounts  = make(map[common.Hash][]byte)
		storage   = make(map[common.Hash]map[common.Hash][]byte) 
	)
  
   
   // diff range
  diff_init   := int64(20474737)
  diff_start  := int64(20474736)
  diff_end    := int64(20474737-127)
  
  //new parent:
  accounts, err := NewBlock(ipcpath,big.NewInt(diff_init))
  if err != nil {
    t.Logf("newblock err : %v", err)
  }
  parent := newDiffLayer(cacheLoadLayer(), common.Hash{}, copyDestructs(destructs), accounts, copyStorage(storage))

  // 2 ~ 128 range diff layer
  for i:=diff_end; i>=diff_start; i-- {
    accounts, err := NewBlock(ipcpath, big.NewInt(int64(i)))
    if err != nil {
      t.Logf("newblock err : %v", err)
    }
    child := parent.Update(common.Hash{}, copyDestructs(destructs), accounts, copyStorage(storage))
    parent = child
  }

  file, err := os.Open("addr_fullnode2.txt")
  if err != nil {
    t.Logf("txt error : %v", err)
  }
  cnt := 1
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {

    line := scanner.Text()
    address := common.HexToAddress(line)
    accountHash := crypto.Keccak256Hash(address.Bytes())

    var acc []byte
    var err error

    acc , err = parent.Account(accountHash)
    if err != nil  || acc == nil {
      t.Log(cnt)
    } else {
    t.Log(cnt)  
    }
    cnt = cnt +1
  }

  t.Log("test complete")
  t.Log("Total execution:",cnt," cachehits : " ,cachehits," cacheMisses", cacheMisses)
 
}

func makeCachefile(){

  file, _ := os.Open("test/addrUnder5.txt")
 

  ipcpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"
  client, _ := rpc.Dial(ipcpath)
  
  cache := fastcache.New(1 * 1024 * 1024)


  ec := ethclient.NewClient(client)
  
  
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {

    line := scanner.Text()
    address := common.HexToAddress(line)
    bal, _ := ec.BalanceAt(context.Background(), address, nil)
    balBytes := bal.Bytes()
    accountHash := crypto.Keccak256Hash(address.Bytes()) // accountHash 생성
    cache.Set(accountHash.Bytes(), balBytes) 
  }
  
  filepath := "test/addrUnder5"
  cache.SaveToFile(filepath)
 
}


