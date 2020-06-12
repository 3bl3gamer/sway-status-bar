package networks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sway_status_bar/utils"
	"time"
)

type Network struct {
	DevName   string
	TxBytes   int64
	RxBytes   int64
	TxSpeed   int64
	RxSpeed   int64
	isPresent bool
}

type NetworksUnit struct {
	ActiveNetworks   []*Network
	NetworkByDevName map[string]*Network
	SkipFunc         func(string) bool
	lastUpdateAt     time.Time
}

func (u *NetworksUnit) Update(iter int64) error {
	if iter%3 != 0 {
		return nil
	}
	if u.NetworkByDevName == nil {
		u.NetworkByDevName = make(map[string]*Network)
	}

	now := time.Now()
	timeDelta := int64(now.Sub(u.lastUpdateAt))
	u.lastUpdateAt = now

	buf, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return err
	}

	u.ActiveNetworks = u.ActiveNetworks[:0]
	lines := strings.Split(strings.TrimSpace(string(buf)), "\n")
	for _, line := range lines[2:] {
		nameEndIndex := strings.IndexByte(line, ':')
		if nameEndIndex == -1 {
			return errors.New("no name separator ':' in: " + line)
		}

		name := strings.TrimSpace(line[:nameEndIndex])
		if u.SkipFunc != nil && u.SkipFunc(name) {
			continue
		}

		var rxBytes, txBytes, dummy int64
		_, err := fmt.Sscanf(line[nameEndIndex+1:], "%d %d%d%d%d%d%d%d %d",
			&rxBytes,
			&dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy,
			&txBytes)
		if err != nil {
			return err
		}

		net, ok := u.NetworkByDevName[name]
		if !ok {
			net = &Network{DevName: name, RxBytes: rxBytes, TxBytes: txBytes}
			u.NetworkByDevName[name] = net
		}

		if net.isPresent {
			net.RxSpeed = utils.MaxInt64(0, rxBytes-net.RxBytes) * int64(time.Second) / timeDelta
			net.TxSpeed = utils.MaxInt64(0, txBytes-net.TxBytes) * int64(time.Second) / timeDelta
		} else {
			net.RxSpeed = 0
			net.TxSpeed = 0
		}
		net.RxBytes = rxBytes
		net.TxBytes = txBytes
		net.isPresent = true
		u.ActiveNetworks = append(u.ActiveNetworks, net) //re-ordering by /proc/net/dev
	}

	for _, net := range u.NetworkByDevName {
		net.isPresent = false
	}
	for _, net := range u.ActiveNetworks {
		net.isPresent = true
	}
	return nil
}
