// Code generated by counterfeiter. DO NOT EDIT.
package dbfakes

import (
	sync "sync"

	atc "github.com/concourse/concourse/atc"
	db "github.com/concourse/concourse/atc/db"
)

type FakeResourceVersion struct {
	CheckOrderStub        func() int
	checkOrderMutex       sync.RWMutex
	checkOrderArgsForCall []struct {
	}
	checkOrderReturns struct {
		result1 int
	}
	checkOrderReturnsOnCall map[int]struct {
		result1 int
	}
	IDStub        func() int
	iDMutex       sync.RWMutex
	iDArgsForCall []struct {
	}
	iDReturns struct {
		result1 int
	}
	iDReturnsOnCall map[int]struct {
		result1 int
	}
	MetadataStub        func() db.ResourceConfigMetadataFields
	metadataMutex       sync.RWMutex
	metadataArgsForCall []struct {
	}
	metadataReturns struct {
		result1 db.ResourceConfigMetadataFields
	}
	metadataReturnsOnCall map[int]struct {
		result1 db.ResourceConfigMetadataFields
	}
	PartialStub        func() bool
	partialMutex       sync.RWMutex
	partialArgsForCall []struct {
	}
	partialReturns struct {
		result1 bool
	}
	partialReturnsOnCall map[int]struct {
		result1 bool
	}
	ReloadStub        func() (bool, error)
	reloadMutex       sync.RWMutex
	reloadArgsForCall []struct {
	}
	reloadReturns struct {
		result1 bool
		result2 error
	}
	reloadReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	ResourceConfigStub        func() db.ResourceConfig
	resourceConfigMutex       sync.RWMutex
	resourceConfigArgsForCall []struct {
	}
	resourceConfigReturns struct {
		result1 db.ResourceConfig
	}
	resourceConfigReturnsOnCall map[int]struct {
		result1 db.ResourceConfig
	}
	SpaceStub        func() atc.Space
	spaceMutex       sync.RWMutex
	spaceArgsForCall []struct {
	}
	spaceReturns struct {
		result1 atc.Space
	}
	spaceReturnsOnCall map[int]struct {
		result1 atc.Space
	}
	VersionStub        func() db.Version
	versionMutex       sync.RWMutex
	versionArgsForCall []struct {
	}
	versionReturns struct {
		result1 db.Version
	}
	versionReturnsOnCall map[int]struct {
		result1 db.Version
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeResourceVersion) CheckOrder() int {
	fake.checkOrderMutex.Lock()
	ret, specificReturn := fake.checkOrderReturnsOnCall[len(fake.checkOrderArgsForCall)]
	fake.checkOrderArgsForCall = append(fake.checkOrderArgsForCall, struct {
	}{})
	fake.recordInvocation("CheckOrder", []interface{}{})
	fake.checkOrderMutex.Unlock()
	if fake.CheckOrderStub != nil {
		return fake.CheckOrderStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.checkOrderReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) CheckOrderCallCount() int {
	fake.checkOrderMutex.RLock()
	defer fake.checkOrderMutex.RUnlock()
	return len(fake.checkOrderArgsForCall)
}

func (fake *FakeResourceVersion) CheckOrderCalls(stub func() int) {
	fake.checkOrderMutex.Lock()
	defer fake.checkOrderMutex.Unlock()
	fake.CheckOrderStub = stub
}

func (fake *FakeResourceVersion) CheckOrderReturns(result1 int) {
	fake.checkOrderMutex.Lock()
	defer fake.checkOrderMutex.Unlock()
	fake.CheckOrderStub = nil
	fake.checkOrderReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeResourceVersion) CheckOrderReturnsOnCall(i int, result1 int) {
	fake.checkOrderMutex.Lock()
	defer fake.checkOrderMutex.Unlock()
	fake.CheckOrderStub = nil
	if fake.checkOrderReturnsOnCall == nil {
		fake.checkOrderReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.checkOrderReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *FakeResourceVersion) ID() int {
	fake.iDMutex.Lock()
	ret, specificReturn := fake.iDReturnsOnCall[len(fake.iDArgsForCall)]
	fake.iDArgsForCall = append(fake.iDArgsForCall, struct {
	}{})
	fake.recordInvocation("ID", []interface{}{})
	fake.iDMutex.Unlock()
	if fake.IDStub != nil {
		return fake.IDStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.iDReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) IDCallCount() int {
	fake.iDMutex.RLock()
	defer fake.iDMutex.RUnlock()
	return len(fake.iDArgsForCall)
}

func (fake *FakeResourceVersion) IDCalls(stub func() int) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = stub
}

func (fake *FakeResourceVersion) IDReturns(result1 int) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = nil
	fake.iDReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeResourceVersion) IDReturnsOnCall(i int, result1 int) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = nil
	if fake.iDReturnsOnCall == nil {
		fake.iDReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.iDReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *FakeResourceVersion) Metadata() db.ResourceConfigMetadataFields {
	fake.metadataMutex.Lock()
	ret, specificReturn := fake.metadataReturnsOnCall[len(fake.metadataArgsForCall)]
	fake.metadataArgsForCall = append(fake.metadataArgsForCall, struct {
	}{})
	fake.recordInvocation("Metadata", []interface{}{})
	fake.metadataMutex.Unlock()
	if fake.MetadataStub != nil {
		return fake.MetadataStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.metadataReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) MetadataCallCount() int {
	fake.metadataMutex.RLock()
	defer fake.metadataMutex.RUnlock()
	return len(fake.metadataArgsForCall)
}

func (fake *FakeResourceVersion) MetadataCalls(stub func() db.ResourceConfigMetadataFields) {
	fake.metadataMutex.Lock()
	defer fake.metadataMutex.Unlock()
	fake.MetadataStub = stub
}

func (fake *FakeResourceVersion) MetadataReturns(result1 db.ResourceConfigMetadataFields) {
	fake.metadataMutex.Lock()
	defer fake.metadataMutex.Unlock()
	fake.MetadataStub = nil
	fake.metadataReturns = struct {
		result1 db.ResourceConfigMetadataFields
	}{result1}
}

func (fake *FakeResourceVersion) MetadataReturnsOnCall(i int, result1 db.ResourceConfigMetadataFields) {
	fake.metadataMutex.Lock()
	defer fake.metadataMutex.Unlock()
	fake.MetadataStub = nil
	if fake.metadataReturnsOnCall == nil {
		fake.metadataReturnsOnCall = make(map[int]struct {
			result1 db.ResourceConfigMetadataFields
		})
	}
	fake.metadataReturnsOnCall[i] = struct {
		result1 db.ResourceConfigMetadataFields
	}{result1}
}

func (fake *FakeResourceVersion) Partial() bool {
	fake.partialMutex.Lock()
	ret, specificReturn := fake.partialReturnsOnCall[len(fake.partialArgsForCall)]
	fake.partialArgsForCall = append(fake.partialArgsForCall, struct {
	}{})
	fake.recordInvocation("Partial", []interface{}{})
	fake.partialMutex.Unlock()
	if fake.PartialStub != nil {
		return fake.PartialStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.partialReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) PartialCallCount() int {
	fake.partialMutex.RLock()
	defer fake.partialMutex.RUnlock()
	return len(fake.partialArgsForCall)
}

func (fake *FakeResourceVersion) PartialCalls(stub func() bool) {
	fake.partialMutex.Lock()
	defer fake.partialMutex.Unlock()
	fake.PartialStub = stub
}

func (fake *FakeResourceVersion) PartialReturns(result1 bool) {
	fake.partialMutex.Lock()
	defer fake.partialMutex.Unlock()
	fake.PartialStub = nil
	fake.partialReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeResourceVersion) PartialReturnsOnCall(i int, result1 bool) {
	fake.partialMutex.Lock()
	defer fake.partialMutex.Unlock()
	fake.PartialStub = nil
	if fake.partialReturnsOnCall == nil {
		fake.partialReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.partialReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeResourceVersion) Reload() (bool, error) {
	fake.reloadMutex.Lock()
	ret, specificReturn := fake.reloadReturnsOnCall[len(fake.reloadArgsForCall)]
	fake.reloadArgsForCall = append(fake.reloadArgsForCall, struct {
	}{})
	fake.recordInvocation("Reload", []interface{}{})
	fake.reloadMutex.Unlock()
	if fake.ReloadStub != nil {
		return fake.ReloadStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.reloadReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeResourceVersion) ReloadCallCount() int {
	fake.reloadMutex.RLock()
	defer fake.reloadMutex.RUnlock()
	return len(fake.reloadArgsForCall)
}

func (fake *FakeResourceVersion) ReloadCalls(stub func() (bool, error)) {
	fake.reloadMutex.Lock()
	defer fake.reloadMutex.Unlock()
	fake.ReloadStub = stub
}

func (fake *FakeResourceVersion) ReloadReturns(result1 bool, result2 error) {
	fake.reloadMutex.Lock()
	defer fake.reloadMutex.Unlock()
	fake.ReloadStub = nil
	fake.reloadReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeResourceVersion) ReloadReturnsOnCall(i int, result1 bool, result2 error) {
	fake.reloadMutex.Lock()
	defer fake.reloadMutex.Unlock()
	fake.ReloadStub = nil
	if fake.reloadReturnsOnCall == nil {
		fake.reloadReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.reloadReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeResourceVersion) ResourceConfig() db.ResourceConfig {
	fake.resourceConfigMutex.Lock()
	ret, specificReturn := fake.resourceConfigReturnsOnCall[len(fake.resourceConfigArgsForCall)]
	fake.resourceConfigArgsForCall = append(fake.resourceConfigArgsForCall, struct {
	}{})
	fake.recordInvocation("ResourceConfig", []interface{}{})
	fake.resourceConfigMutex.Unlock()
	if fake.ResourceConfigStub != nil {
		return fake.ResourceConfigStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.resourceConfigReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) ResourceConfigCallCount() int {
	fake.resourceConfigMutex.RLock()
	defer fake.resourceConfigMutex.RUnlock()
	return len(fake.resourceConfigArgsForCall)
}

func (fake *FakeResourceVersion) ResourceConfigCalls(stub func() db.ResourceConfig) {
	fake.resourceConfigMutex.Lock()
	defer fake.resourceConfigMutex.Unlock()
	fake.ResourceConfigStub = stub
}

func (fake *FakeResourceVersion) ResourceConfigReturns(result1 db.ResourceConfig) {
	fake.resourceConfigMutex.Lock()
	defer fake.resourceConfigMutex.Unlock()
	fake.ResourceConfigStub = nil
	fake.resourceConfigReturns = struct {
		result1 db.ResourceConfig
	}{result1}
}

func (fake *FakeResourceVersion) ResourceConfigReturnsOnCall(i int, result1 db.ResourceConfig) {
	fake.resourceConfigMutex.Lock()
	defer fake.resourceConfigMutex.Unlock()
	fake.ResourceConfigStub = nil
	if fake.resourceConfigReturnsOnCall == nil {
		fake.resourceConfigReturnsOnCall = make(map[int]struct {
			result1 db.ResourceConfig
		})
	}
	fake.resourceConfigReturnsOnCall[i] = struct {
		result1 db.ResourceConfig
	}{result1}
}

func (fake *FakeResourceVersion) Space() atc.Space {
	fake.spaceMutex.Lock()
	ret, specificReturn := fake.spaceReturnsOnCall[len(fake.spaceArgsForCall)]
	fake.spaceArgsForCall = append(fake.spaceArgsForCall, struct {
	}{})
	fake.recordInvocation("Space", []interface{}{})
	fake.spaceMutex.Unlock()
	if fake.SpaceStub != nil {
		return fake.SpaceStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.spaceReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) SpaceCallCount() int {
	fake.spaceMutex.RLock()
	defer fake.spaceMutex.RUnlock()
	return len(fake.spaceArgsForCall)
}

func (fake *FakeResourceVersion) SpaceCalls(stub func() atc.Space) {
	fake.spaceMutex.Lock()
	defer fake.spaceMutex.Unlock()
	fake.SpaceStub = stub
}

func (fake *FakeResourceVersion) SpaceReturns(result1 atc.Space) {
	fake.spaceMutex.Lock()
	defer fake.spaceMutex.Unlock()
	fake.SpaceStub = nil
	fake.spaceReturns = struct {
		result1 atc.Space
	}{result1}
}

func (fake *FakeResourceVersion) SpaceReturnsOnCall(i int, result1 atc.Space) {
	fake.spaceMutex.Lock()
	defer fake.spaceMutex.Unlock()
	fake.SpaceStub = nil
	if fake.spaceReturnsOnCall == nil {
		fake.spaceReturnsOnCall = make(map[int]struct {
			result1 atc.Space
		})
	}
	fake.spaceReturnsOnCall[i] = struct {
		result1 atc.Space
	}{result1}
}

func (fake *FakeResourceVersion) Version() db.Version {
	fake.versionMutex.Lock()
	ret, specificReturn := fake.versionReturnsOnCall[len(fake.versionArgsForCall)]
	fake.versionArgsForCall = append(fake.versionArgsForCall, struct {
	}{})
	fake.recordInvocation("Version", []interface{}{})
	fake.versionMutex.Unlock()
	if fake.VersionStub != nil {
		return fake.VersionStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.versionReturns
	return fakeReturns.result1
}

func (fake *FakeResourceVersion) VersionCallCount() int {
	fake.versionMutex.RLock()
	defer fake.versionMutex.RUnlock()
	return len(fake.versionArgsForCall)
}

func (fake *FakeResourceVersion) VersionCalls(stub func() db.Version) {
	fake.versionMutex.Lock()
	defer fake.versionMutex.Unlock()
	fake.VersionStub = stub
}

func (fake *FakeResourceVersion) VersionReturns(result1 db.Version) {
	fake.versionMutex.Lock()
	defer fake.versionMutex.Unlock()
	fake.VersionStub = nil
	fake.versionReturns = struct {
		result1 db.Version
	}{result1}
}

func (fake *FakeResourceVersion) VersionReturnsOnCall(i int, result1 db.Version) {
	fake.versionMutex.Lock()
	defer fake.versionMutex.Unlock()
	fake.VersionStub = nil
	if fake.versionReturnsOnCall == nil {
		fake.versionReturnsOnCall = make(map[int]struct {
			result1 db.Version
		})
	}
	fake.versionReturnsOnCall[i] = struct {
		result1 db.Version
	}{result1}
}

func (fake *FakeResourceVersion) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.checkOrderMutex.RLock()
	defer fake.checkOrderMutex.RUnlock()
	fake.iDMutex.RLock()
	defer fake.iDMutex.RUnlock()
	fake.metadataMutex.RLock()
	defer fake.metadataMutex.RUnlock()
	fake.partialMutex.RLock()
	defer fake.partialMutex.RUnlock()
	fake.reloadMutex.RLock()
	defer fake.reloadMutex.RUnlock()
	fake.resourceConfigMutex.RLock()
	defer fake.resourceConfigMutex.RUnlock()
	fake.spaceMutex.RLock()
	defer fake.spaceMutex.RUnlock()
	fake.versionMutex.RLock()
	defer fake.versionMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeResourceVersion) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ db.ResourceVersion = new(FakeResourceVersion)
