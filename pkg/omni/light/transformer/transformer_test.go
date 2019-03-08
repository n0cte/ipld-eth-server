// VulcanizeDB
// Copyright © 2018 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package transformer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/vulcanizedb/pkg/fakes"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/retriever"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/transformer"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/contract"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/parser"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/poller"
)

var _ = Describe("Transformer", func() {
	var fakeAddress = "0x1234567890abcdef"

	Describe("SetEvents", func() {
		It("Sets which events to watch from the given contract address", func() {
			watchedEvents := []string{"Transfer", "Mint"}
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetEvents(fakeAddress, watchedEvents)
			Expect(t.WatchedEvents[fakeAddress]).To(Equal(watchedEvents))
		})
	})

	Describe("SetEventAddrs", func() {
		It("Sets which account addresses to watch events for", func() {
			eventAddrs := []string{"test1", "test2"}
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetEventArgs(fakeAddress, eventAddrs)
			Expect(t.EventArgs[fakeAddress]).To(Equal(eventAddrs))
		})
	})

	Describe("SetMethods", func() {
		It("Sets which methods to poll at the given contract address", func() {
			watchedMethods := []string{"balanceOf", "totalSupply"}
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetMethods(fakeAddress, watchedMethods)
			Expect(t.WantedMethods[fakeAddress]).To(Equal(watchedMethods))
		})
	})

	Describe("SetMethodAddrs", func() {
		It("Sets which account addresses to poll methods against", func() {
			methodAddrs := []string{"test1", "test2"}
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetMethodArgs(fakeAddress, methodAddrs)
			Expect(t.MethodArgs[fakeAddress]).To(Equal(methodAddrs))
		})
	})

	Describe("SetStartingBlock", func() {
		It("Sets the block range that the contract should be watched within", func() {
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetStartingBlock(fakeAddress, 11)
			Expect(t.ContractStart[fakeAddress]).To(Equal(int64(11)))
		})
	})

	Describe("SetCreateAddrList", func() {
		It("Sets the block range that the contract should be watched within", func() {
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetCreateAddrList(fakeAddress, true)
			Expect(t.CreateAddrList[fakeAddress]).To(Equal(true))
		})
	})

	Describe("SetCreateHashList", func() {
		It("Sets the block range that the contract should be watched within", func() {
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetCreateHashList(fakeAddress, true)
			Expect(t.CreateHashList[fakeAddress]).To(Equal(true))
		})
	})

	Describe("Init", func() {
		It("Initializes transformer's contract objects", func() {
			blockRetriever := &fakes.MockLightBlockRetriever{}
			firstBlock := int64(1)
			blockRetriever.FirstBlock = firstBlock

			parsr := &fakes.MockParser{}
			fakeAbi := "fake_abi"
			parsr.AbiToReturn = fakeAbi

			pollr := &fakes.MockPoller{}
			fakeContractName := "fake_contract_name"
			pollr.ContractName = fakeContractName

			t := getFakeTransformer(blockRetriever, parsr, pollr)
			t.SetEvents(fakeAddress, []string{"Transfer"})

			err := t.Init()

			Expect(err).ToNot(HaveOccurred())

			c, ok := t.Contracts[fakeAddress]
			Expect(ok).To(Equal(true))

			Expect(c.StartingBlock).To(Equal(firstBlock))
			Expect(c.LastBlock).To(Equal(int64(-1)))
			Expect(c.Abi).To(Equal(fakeAbi))
			Expect(c.Name).To(Equal(fakeContractName))
			Expect(c.Address).To(Equal(fakeAddress))
		})

		It("Fails to initialize if first and most recent block numbers cannot be fetched from vDB headers table", func() {
			blockRetriever := &fakes.MockLightBlockRetriever{}
			blockRetriever.FirstBlockErr = fakes.FakeError
			t := getFakeTransformer(blockRetriever, &fakes.MockParser{}, &fakes.MockPoller{})
			t.SetEvents(fakeAddress, []string{"Transfer"})

			err := t.Init()

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakes.FakeError))
		})

		It("Does nothing if watched events are unset", func() {
			t := getFakeTransformer(&fakes.MockLightBlockRetriever{}, &fakes.MockParser{}, &fakes.MockPoller{})

			err := t.Init()

			Expect(err).ToNot(HaveOccurred())

			_, ok := t.Contracts[fakeAddress]
			Expect(ok).To(Equal(false))
		})
	})

})

func getFakeTransformer(blockRetriever retriever.BlockRetriever, parsr parser.Parser, pollr poller.Poller) transformer.Transformer {
	return transformer.Transformer{
		Parser:           parsr,
		BlockRetriever:   blockRetriever,
		Poller:           pollr,
		HeaderRepository: &fakes.MockLightHeaderRepository{},
		Contracts:        map[string]*contract.Contract{},
		WatchedEvents:    map[string][]string{},
		WantedMethods:    map[string][]string{},
		ContractStart:    map[string]int64{},
		EventArgs:        map[string][]string{},
		MethodArgs:       map[string][]string{},
		CreateAddrList:   map[string]bool{},
		CreateHashList:   map[string]bool{},
	}
}