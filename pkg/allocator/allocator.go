// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package allocator

import (
	"context"
	"fmt"
	"net"

	typesVer "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
)

type Store interface {
	LastReservedIP(ctx context.Context) (net.IP, error)
	Reserve(ctx context.Context, namespace, pod string, requestedIp string) error
}

type IPAllocator struct {
	rangeset     *RangeSet
	store        Store
	allocatedIps []string
}

func NewIPAllocator(s *RangeSet, store Store, allocatedIps []string) *IPAllocator {
	return &IPAllocator{
		rangeset:     s,
		store:        store,
		allocatedIps: allocatedIps,
	}
}

func (a *IPAllocator) InAlreadyAllocate(expectIP string) bool {
	for _, ip := range a.allocatedIps {
		if ip == expectIP {
			return true
		}
	}
	return false
}

// Allocate allocates an IP
func (a *IPAllocator) Allocate(namespace, pod string) (*typesVer.IPConfig, error) {
	var reservedIP *net.IPNet
	var gw net.IP

	iter, err := a.GetIter()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		reservedIP, gw = iter.Next()
		if reservedIP == nil {
			break
		}

		if a.InAlreadyAllocate(reservedIP.IP.String()) {
			continue
		}

		err := a.store.Reserve(ctx, namespace, pod, reservedIP.String())
		if err != nil {
			return nil, err
		}
		break
	}

	if reservedIP == nil {
		return nil, fmt.Errorf("no IP addresses available in range set: %s", a.rangeset.String())
	}

	return &typesVer.IPConfig{
		Address: *reservedIP,
		Gateway: gw,
	}, nil
}

type RangeIter struct {
	rangeset *RangeSet
	// The current range id
	rangeIdx int
	// Our current position
	cur net.IP
	// The IP and range index where we started iterating; if we hit this again, we're done.
	startIP    net.IP
	startRange int
}

// GetIter encapsulates the strategy for this allocator.
// We use a round-robin strategy, attempting to evenly use the whole set.
// More specifically, a crash-looping container will not see the same IP until
// the entire range has been run through.
// We may wish to consider avoiding recently-released IPs in the future.
func (a *IPAllocator) GetIter() (*RangeIter, error) {
	iter := RangeIter{rangeset: a.rangeset}
	// Round-robin by trying to allocate from the last reserved IP + 1
	startFromLastReservedIP := false
	// We might get a last reserved IP that is wrong if the range indexes changed.
	// This is not critical, we just lose round-robin this one time.
	lastReservedIP, err := a.store.LastReservedIP(context.Background())
	if err != nil {
		return nil, err
	}
	startFromLastReservedIP = a.rangeset.Contains(lastReservedIP)
	// Find the range in the set with this IP
	if startFromLastReservedIP {
		for i, r := range *a.rangeset {
			if r.Contains(lastReservedIP) {
				iter.rangeIdx = i
				iter.startRange = i
				// We advance the cursor on every Next(), so the first call
				// to next() will return lastReservedIP + 1
				iter.cur = lastReservedIP
				break
			}
		}
	} else {
		iter.rangeIdx = 0
		iter.startRange = 0
		iter.startIP = (*a.rangeset)[0].RangeStart
	}
	return &iter, nil
}

// Next returns the next IP, its mask, and its gateway. Returns nil
// if the iterator has been exhausted
func (i *RangeIter) Next() (*net.IPNet, net.IP) {
	r := (*i.rangeset)[i.rangeIdx]
	// If this is the first time iterating and we're not starting in the middle
	// of the range, then start at rangeStart, which is inclusive
	if i.cur == nil {
		i.cur = r.RangeStart
		i.startIP = i.cur
		if i.cur.Equal(r.Gateway) {
			return i.Next()
		}
		return &net.IPNet{IP: i.cur, Mask: r.Subnet.Mask}, r.Gateway
	}

	// If we've reached the end of this range, we need to advance the range
	// RangeEnd is inclusive as well
	if i.cur.Equal(r.RangeEnd) {
		i.rangeIdx += 1
		i.rangeIdx %= len(*i.rangeset)
		r = (*i.rangeset)[i.rangeIdx]
		i.cur = r.RangeStart
	} else {
		i.cur = ip.NextIP(i.cur)
	}

	if i.startIP == nil {
		i.startIP = i.cur
	} else if i.rangeIdx == i.startRange && i.cur.Equal(i.startIP) {
		// IF we've looped back to where we started, give up
		return nil, nil
	}

	if i.cur.Equal(r.Gateway) {
		return i.Next()
	}

	return &net.IPNet{IP: i.cur, Mask: r.Subnet.Mask}, r.Gateway
}
