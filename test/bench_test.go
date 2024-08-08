package test

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"testing"
  "sync"
	"sync/atomic"
  "math/rand"
  "math"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethclient"
  "github.com/ethereum/go-ethereum/core/types"
  bloomfilter "github.com/holiman/bloomfilter/v2"
  "github.com/ethereum/go-ethereum/rlp"
  "github.com/ethereum/go-ethereum/ethdb"
)


const (
  ipcpath = "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"  
)
func cacheload(cachePath string) (*fastcache.Cache){
  
  //load cache 
  loadedCache, err := fastcache.LoadFromFile(cachePath) 
  if err != nil {
    log.Fatalf("Failed to load cache : %v", err)
  }

  return loadedCache
}
//DiskAccess 함수는 eth.getBalance를 호출
//disklayer에는 메모리DB를 적재
//디스크DB는 eth.getBlance로 역할 대체함
func DiskAccess(hash common.Hash){
   
}

func NewBlock(ipcpath string,number *big.Int) (map[common.Hash][]byte, error){

  client, _ := ethclient.Dial(ipcpath)
 
  block, err := client.BlockByNumber(context.Background(), number)
  if err != nil {
    return nil, err
  }

  accountMap := make(map[common.Hash][]byte)

  for _, tx := range block.Transactions() {
    to := tx.To()
    if to == nil {
      continue
    }

    balance, err := client.BalanceAt(context.Background(), *to, block.Number())
    if err != nil {
      return nil, fmt.Errorf("failed to get balance for account %s at block %d: %v", to.Hex(), number, err)
    }

    accountMap[common.BytesToHash(to.Bytes())] = balance.Bytes()
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
	// aggregatorMemoryLimit is the maximum size of the bottom-most diff layer
	// that aggregates the writes from above until it's flushed into the disk
	// layer.
	//
	// Note, bumping this up might drastically increase the size of the bloom
	// filters that's stored in every diff layer. Don't do that without fully
	// understanding all the implications.
	aggregatorMemoryLimit = uint64(4 * 1024 * 1024)

	// aggregatorItemLimit is an approximate number of items that will end up
	// in the aggregator layer before it's flushed out to disk. A plain account
	// weighs around 14B (+hash), a storage slot 32B (+hash), a deleted slot
	// 0B (+hash). Slots are mostly set/unset in lockstep, so that average at
	// 16B (+hash). All in all, the average entry seems to be 15+32=47B. Use a
	// smaller number to be on the safe side.
	aggregatorItemLimit = aggregatorMemoryLimit / 42

	// bloomTargetError is the target false positive rate when the aggregator
	// layer is at its fullest. The actual value will probably move around up
	// and down from this number, it's mostly a ballpark figure.
	//
	// Note, dropping this down might drastically increase the size of the bloom
	// filters that's stored in every diff layer. Don't do that without fully
	// understanding all the implications.
	bloomTargetError = 0.02

	// bloomSize is the ideal bloom filter size given the maximum number of items
	// it's expected to hold and the target false positive error rate.
	bloomSize = math.Ceil(float64(aggregatorItemLimit) * math.Log(bloomTargetError) / math.Log(1/math.Pow(2, math.Log(2))))

	// bloomFuncs is the ideal number of bits a single entry should set in the
	// bloom filter to keep its size to a minimum (given it's size and maximum
	// entry count).
	bloomFuncs = math.Round((bloomSize / float64(aggregatorItemLimit)) * math.Log(2))

	// the bloom offsets are runtime constants which determines which part of the
	// account/storage hash the hasher functions looks at, to determine the
	// bloom key for an account/slot. This is randomized at init(), so that the
	// global population of nodes do not all display the exact same behaviour with
	// regards to bloom content
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

func (dl *diffLayer) Account(hash common.Hash) (*types.SlimAccount, error) {
	data, err := dl.AccountRLP(hash)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	account := new(types.SlimAccount)
	if err := rlp.DecodeBytes(data, account); err != nil {
		panic(err)
	}
	return account, nil
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
	cache  *fastcache.Cache    // Cache to avoid hitting the disk for direct access

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
	Account(hash common.Hash) (*types.SlimAccount, error)

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

func (dl *diskLayer) Account(hash common.Hash) (*types.SlimAccount, error) {
	data, err := dl.AccountRLP(hash)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 { // can be both nil and []byte{}
		return nil, nil
	}
	account := new(types.SlimAccount)
	if err := rlp.DecodeBytes(data, account); err != nil {
		panic(err)
	}
	return account, nil
}

func (dl *diskLayer) AccountRLP(hash common.Hash) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	// Try to retrieve the account from the memory cache
	if blob, found := dl.cache.HasGet(nil, hash[:]); found {
			return blob, nil
	}
	// Cache doesn't contain account, pull from disk and cache for later
	blob := rawdb.ReadAccountSnapshot(dl.diskdb, hash)
	dl.cache.Set(hash[:], blob)

	return blob, nil
}

func TestSnapshot2(t *testing.T){
  
  makeRoot := func(height uint64) common.Hash {
		var buffer [8]byte
		binary.BigEndian.PutUint64(buffer[:], height)
		return common.BytesToHash(buffer[:])
	}
  
  base := &diskLayer{
		diskdb: rawdb.NewMemoryDatabase(),
		root:   makeRoot(1),
		cache:  fastcache.New(1024 * 500),
	}
  
  snaps := &Tree{
    layers: map[common.Hash] snapshot {
      base.root: base,
    },
  }
  //big.NewInt(int64(i)
  diff_start  := int64(20474737)
  diff_end    := int64(20474737-128)
  
  
 

  last := common.HexToHash("0x01")

  for i:= diff_end; i<=diff_start; i++ {
    head := common.BytesToHash([]byte{byte(i+2)})
    accounts, _ := NewBlock(ipcpath, big.NewInt(int64(i)))
    snaps.Update(head, last, nil, accounts, nil)
    last = head
  }
  
}
