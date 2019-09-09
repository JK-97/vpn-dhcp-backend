package main

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func hello() error {

	cli, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer cli.Close()

	devices, err := cli.Devices()
	if err != nil {
		return err
	}

	if len(devices) > 0 {
		for _, dev := range devices {
			fmt.Println(dev)
		}

	} else {
		key, err := wgtypes.GeneratePrivateKey()
		err = cli.ConfigureDevice("wg0", wgtypes.Config{
			PrivateKey: &key,
		})
		if err != nil {
			return err
		}

		fmt.Println(cli.Device("wg0"))

	}

	return nil
}
