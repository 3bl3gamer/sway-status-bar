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
	Networks         []*Network
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

	for _, net := range u.Networks {
		net.isPresent = false
	}
	u.Networks = u.Networks[:0]

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

		net, alreadyExists := u.NetworkByDevName[name]
		if !alreadyExists {
			net = &Network{DevName: name, RxBytes: rxBytes, TxBytes: txBytes}
			u.NetworkByDevName[name] = net
		}

		if alreadyExists {
			net.RxSpeed = utils.MaxInt64(0, rxBytes-net.RxBytes) * int64(time.Second) / timeDelta
			net.TxSpeed = utils.MaxInt64(0, txBytes-net.TxBytes) * int64(time.Second) / timeDelta
		}
		net.RxBytes = rxBytes
		net.TxBytes = txBytes
		net.isPresent = true
		u.Networks = append(u.Networks, net) //re-ordering by /proc/net/dev
	}

	for name, net := range u.NetworkByDevName {
		if !net.isPresent {
			delete(u.NetworkByDevName, name)
		}
	}
	return nil
}
