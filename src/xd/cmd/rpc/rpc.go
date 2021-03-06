package rpc

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"xd/lib/bittorrent/swarm"
	"xd/lib/config"
	"xd/lib/log"
	"xd/lib/rpc"
)

var formatUnits = map[int]string{
	0: "B",
	1: "KB",
	2: "MB",
	3: "GB",
}

func formatRate(rate float32) string {
	r := uint32(rate)
	idx := 0
	for r > 1024 {
		r /= 1024
		idx++
	}
	return fmt.Sprintf("%d %s/s", r, formatUnits[idx])
}

func Run() {
	fname := "torrents.ini"
	if len(os.Args) > 1 {
		fname = os.Args[1]
	}
	cfg := new(config.Config)
	err := cfg.Load(fname)
	if err != nil {
		log.Errorf("error: %s", err)
		return
	}
	log.SetLevel(cfg.Log.Level)
	u := url.URL{
		Host: cfg.RPC.Bind,
		Path: rpc.RPCPath,
	}
	c := rpc.NewClient(u.String())
	var list swarm.TorrentsList
	list, err = c.ListTorrents()
	if err != nil {
		log.Errorf("rpc error: %s", err)
		return
	}
	var globalTx, globalRx float32

	var torrents swarm.TorrentStatusList
	sort.Stable(&list.Infohashes)

	for _, ih := range list.Infohashes {
		var status swarm.TorrentStatus
		status, err = c.SwarmStatus(ih)

		if err != nil {
			log.Errorf("rpc error: %s", err)
			return
		}

		torrents = append(torrents, status)

	}
	sort.Stable(&torrents)
	for _, status := range torrents {
		var tx, rx float32
		fmt.Printf("%s [%s]\n", status.Name, status.Infohash)
		sort.Stable(&status.Peers)
		for _, peer := range status.Peers {
			fmt.Printf("%s tx=%s rx=%s\n", peer.ID, formatRate(peer.TX), formatRate(peer.RX))
			tx += peer.TX
			rx += peer.RX
		}
		fmt.Printf("\n%s tx=%s rx=%s\n", status.State, formatRate(tx), formatRate(rx))
		fmt.Println()
		globalRx += rx
		globalTx += tx
	}
	fmt.Println()
	fmt.Printf("%d torrents: tx=%s rx=%s\n", list.Infohashes.Len(), formatRate(globalTx), formatRate(globalRx))
}
