// This file was generated by counterfeiter
package generatorfakes

import (
	"sync"

	"github.com/cloudfoundry/cli/words/generator"
)

type FakeWordGenerator struct {
	BabbleStub        func() string
	babbleMutex       sync.RWMutex
	babbleArgsForCall []struct{}
	babbleReturns     struct {
		result1 string
	}
}

func (fake *FakeWordGenerator) Babble() string {
	fake.babbleMutex.Lock()
	fake.babbleArgsForCall = append(fake.babbleArgsForCall, struct{}{})
	fake.babbleMutex.Unlock()
	if fake.BabbleStub != nil {
		return fake.BabbleStub()
	} else {
		return fake.babbleReturns.result1
	}
}

func (fake *FakeWordGenerator) BabbleCallCount() int {
	fake.babbleMutex.RLock()
	defer fake.babbleMutex.RUnlock()
	return len(fake.babbleArgsForCall)
}

func (fake *FakeWordGenerator) BabbleReturns(result1 string) {
	fake.BabbleStub = nil
	fake.babbleReturns = struct {
		result1 string
	}{result1}
}

var _ generator.WordGenerator = new(FakeWordGenerator)