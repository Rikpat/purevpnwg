package wireguard

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/Rikpat/purevpnwg/pkg/util"
	"gopkg.in/ini.v1"
)

type WireguardConfig struct {
	Interface
	Peer
}

type Interface struct {
	PrivateKey, Address, DNS, Table string
}

type Peer struct {
	PublicKey, AllowedIPs, Endpoint, PersistentKeepalive string
}

func resolveIpFromHostname(conf *WireguardConfig) {
	// [0]: endpoint [1]: port
	endpointSlice := strings.Split(conf.Peer.Endpoint, ":")
	ips, _ := net.LookupIP(endpointSlice[0])
	conf.Peer.Endpoint = ips[0].To4().String() + ":" + endpointSlice[1]
}

func UpdateConfig(newConfig []byte, config *util.Config) error {
	wgConf := new(WireguardConfig)

	err := ini.MapTo(wgConf, newConfig)
	if err != nil {
		return err
	}

	if config.ResolveIP {
		resolveIpFromHostname(wgConf)
	}

	if _, err := os.Stat(config.WireguardFile); errors.Is(err, os.ErrNotExist) {
		wgConfFile := ini.Empty()
		if err := wgConfFile.ReflectFrom(wgConf); err != nil {
			return err
		}
		return wgConfFile.SaveTo(config.WireguardFile)
	} else if err != nil {
		return err
	}

	wgConfFile, err := ini.Load(config.WireguardFile)
	if err != nil {
		return err
	}
	iface := wgConfFile.Section("Interface")
	iface.Key("Address").SetValue(wgConf.Interface.Address)
	iface.Key("PrivateKey").SetValue(wgConf.Interface.PrivateKey)
	iface.Key("DNS").SetValue(wgConf.Interface.DNS)

	peer := wgConfFile.Section("Peer")
	peer.Key("PublicKey").SetValue(wgConf.Peer.PublicKey)
	peer.Key("Endpoint").SetValue(wgConf.Peer.Endpoint)

	return wgConfFile.SaveTo(config.WireguardFile)
}
