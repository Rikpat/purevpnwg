package purevpn

import (
	"encoding/json"
	"fmt"

	"github.com/Rikpat/purevpnwg/pkg/util"
	"github.com/go-rod/rod"
)

// Simplified only to needed things
type Subscription struct {
	ID           string   `json:"id"`
	State        string   `json:"state"`
	VPNUsernames []string `json:"vpnusernames"`
}

func parseSubscription(subscriptionResponse string) (*Subscription, error) {
	sub := Subscription{}
	if err := json.Unmarshal([]byte(subscriptionResponse), &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func GetSubscriptions(page *rod.Page, config *util.Config, token string) (*Subscription, error) {
	if config.Debug {
		fmt.Println("Requesting subscriptions")
	}
	page.MustNavigate(_BASE_URL).MustWaitNavigation()
	res, err := page.Eval(`
		async (authorization, privateKey) => {
			const res = await fetch(
				"/v2/api/subscription",
				{
					method: "POST",
					headers: {
						accept: 'application/json',
						authorization
					}
				}
			)
			if (!res.ok) {
				throw Error(await res.text())
			}
			const json = await res.json()
			if (!json.status) {
				throw Error(JSON.stringify(json))
			}
			// Pick first active subscription, maybe add more logic later
			return JSON.stringify(json.body.find(sub => sub.state === "active"))
	}`, "Bearer "+token)
	if err != nil {
		return nil, err
	}
	return parseSubscription(res.Value.Str())
}

func (sub *Subscription) ToSubscriptionAuth() *util.SubscriptionAuth {
	return &util.SubscriptionAuth{ID: sub.ID, Username: sub.VPNUsernames[0]}
}

// func (subs *Subscriptions) getActiveSubscriptions() (subs []string) {
// 	for _, s := range subs.Subscription[user.Type] {
// 		if s.Status == "active" {
// 			subs = append(subs, s.Vpnusername)
// 		}
// 	}
// 	return subs
// }

// func (user *UserData) SelectSubscription() (*util.SubscriptionAuth, error) {
// 	subs := user.getActiveSubscriptions()
// 	if len(subs) > 1 {
// 		prompt := promptui.Select{
// 			Label: "Select Subscription",
// 			Items: subs,
// 		}

// 		_, result, err := prompt.Run()

// 		if err != nil {
// 			fmt.Printf("Prompt failed %v\n", err)
// 			return nil, err
// 		}

// 		return &util.SubscriptionAuth{Username: result}, nil
// 	}
// 	return &util.SubscriptionAuth{Username: subs[0]}, nil
// }
