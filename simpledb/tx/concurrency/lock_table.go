package concurrency

import (
	"fmt"
	"sync"
	"time"

	"simpledb/file"
)

const maxLockTime = 10 * time.Second

var ErrTimeout = fmt.Errorf("timeout error")

type LockTable struct {
	locks map[file.BlockID]int
	mux   *sync.Mutex
}

func newLockTable() *LockTable {
	return &LockTable{
		locks: make(map[file.BlockID]int),
		mux:   &sync.Mutex{},
	}
}

// TODO: time.Sleepを消してかっこいい方法で実装する
func (l *LockTable) SLock(blockID file.BlockID) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	timeoutChan := time.After(maxLockTime)
	for {
		select {
		case <-timeoutChan:
			return ErrTimeout
		default:
			if !l.hasXLock(blockID) {
				goto notlocked
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
notlocked:

	l.locks[blockID]++
	return nil
}

func (l *LockTable) XLock(blockID file.BlockID) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	timeoutChan := time.After(maxLockTime)
	for {
		select {
		case <-timeoutChan:
			return ErrTimeout
		default:
			if !l.hasOtherSLocks(blockID) {
				goto notlocked
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
notlocked:

	l.locks[blockID] = -1
	return nil
}

func (l *LockTable) Unlock(blockID file.BlockID) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.locks[blockID] > 1 {
		l.locks[blockID]--
	} else {
		delete(l.locks, blockID)
		// TODO notifyAll()
	}
}

func (l *LockTable) hasXLock(blockID file.BlockID) bool {
	return l.locks[blockID] < 0
}

func (l *LockTable) hasOtherSLocks(blockID file.BlockID) bool {
	return l.locks[blockID] > 1
}
