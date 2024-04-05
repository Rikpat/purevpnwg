package purevpnwg

import (
	"fmt"

	"github.com/Rikpat/purevpnwg/pkg/purevpn"
	"github.com/Rikpat/purevpnwg/pkg/util"
	"github.com/Rikpat/purevpnwg/pkg/wireguard"
)

type FullCmd struct {
}

func (r *FullCmd) Run(ctx *Context) error {
	page, cookies := purevpn.Login(ctx.Config.Username, ctx.Config.Password)
	defer page.MustClose()

	if ctx.Config.Debug {
		fmt.Println("Successfully Logged in")
	}

	token := util.FilterCookies(cookies, "fa_token")
	if len(token) == 0 {
		return fmt.Errorf("no token in cookies")
	}

	userData, err := purevpn.GetUserData(token[0].Value)
	if err != nil {
		return err
	}

	if ctx.Config.Debug {
		fmt.Println("Successfully parsed user data")
	}

	if ctx.Config.Subscription.Username == "" {
		if ctx.Config.Subscription, err = userData.SelectSubscription(); err != nil {
			return err
		}
	}

	if ctx.Config.Debug {
		fmt.Printf("Selected subscription %v\n", ctx.Config.Subscription.Username)
	}

	if err := ctx.Config.Subscription.GetEncryptPassword(page, token[0].Value); err != nil {
		return err
	}

	if ctx.Config.Debug {
		fmt.Println("Successfully got subscription password")
		fmt.Printf("ctx.Config: %v\n", ctx.Config)
	}

	server, err := purevpn.GetWireguardServer(page, ctx.Config, token[0].Value)

	if err == nil {
		if ctx.Config.Debug {
			fmt.Printf("Got wireguard server: %v\n", server)
		}
		err = wireguard.UpdateConfig([]byte(server), ctx.Config)
		if ctx.Config.Debug && err == nil {
			fmt.Printf("Created wireguard file at %v\n", ctx.Config.WireguardFile)
		}
	}
	return err
}
