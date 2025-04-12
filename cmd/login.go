package purevpnwg

import (
	"fmt"

	"github.com/Rikpat/purevpnwg/pkg/purevpn"
	"github.com/Rikpat/purevpnwg/pkg/util"
)

type LoginCmd struct {
}

func (r *LoginCmd) Run(ctx *Context) error {
	page, cookies := purevpn.Login(ctx.Config.Username, ctx.Config.Password)
	defer page.MustClose()

	token := util.FilterCookies(cookies, "fa_token")
	if len(token) == 0 {
		return fmt.Errorf("no token in cookies")
	}
	config, err := util.ReadConfig()
	if err != nil {
		return err
	}

	userData, err := purevpn.GetUserData(token[0].Value)
	if err != nil {
		return err
	}

	config.UUID = userData.AccountCode
	if config.Subscription == nil || config.Subscription.Username == "" || config.Subscription.Password == "" || ctx.Config.Subscription.ID == "" {
		if sub, err := purevpn.GetSubscriptions(page, ctx.Config, token[0].Value); err == nil {
			ctx.Config.Subscription = sub.ToSubscriptionAuth()
		} else {
			return err
		}

		if err := config.Subscription.GetEncryptPassword(page, token[0].Value); err != nil {
			return err
		}

	}
	err = util.WriteConfig(config)
	if err != nil {
		return err
	}

	err = util.WriteCookies("cookies.gob", util.FilterCookies(cookies, "fusionauth"))
	if err == nil {
		fmt.Println("Successfully logged in and stored cookies")
	}
	return err
}
