package purevpn

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Rikpat/purevpnwg/pkg/util"
	"github.com/Rikpat/purevpnwg/pkg/wireguard"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Change if changed in the future
const _BASE_URL = "https://my.purevpn.com/"

var _PATH, _ = launcher.LookPath()
var _LAUNCHER = launcher.New().Bin(_PATH).Leakless(false).MustLaunch()

func Login(username, password string) (*rod.Page, []*proto.NetworkCookie) {
	browser := rod.New().Timeout(time.Minute).ControlURL(_LAUNCHER).MustConnect()

	page := stealth.MustPage(browser)

	page.MustNavigate(_BASE_URL).MustWaitNavigation()

	page.MustElement(`input#loginId`).MustWaitVisible().MustInput(username)
	page.MustElement(`input#password`).MustInput(password).MustType(input.Enter)

	page.MustWaitIdle()

	cookies := page.Browser().MustGetCookies()

	return page, cookies
}

func InitPage() (*rod.Page, error) {
	cookies, err := util.ReadCookies("cookies.gob")
	if err != nil {
		return nil, err
	}
	if util.AreCookiesExpired(cookies) {
		return nil, fmt.Errorf("cookies are expired, run purevpnwg login")
	}
	browser := rod.New().Timeout(time.Minute).ControlURL(_LAUNCHER).MustConnect().MustSetCookies(cookies...)

	page := stealth.MustPage(browser)

	page.MustNavigate(_BASE_URL).MustWaitNavigation()

	return page, nil
}

func GetToken(page *rod.Page, uuid string) (string, error) {
	page.MustNavigate(_BASE_URL).MustWaitNavigation()
	res, err := page.Eval(`
		async (uuid) => {
			const res = await fetch("/v2/api/fusionauth/auto-login", {
				method: "POST",
				body: new URLSearchParams({uuid}).toString(),
				headers: {
					'content-type': 'application/x-www-form-urlencoded',
					'accept': 'application/json'
				}
			})
			if (!res.ok) {
				throw Error(await res.text())
			}
			const json = await res.json()
			if (!json.status) {
				throw Error(json)
			}
			return json.body.token
		}
	`, uuid)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return "", err
	}
	return res.Value.Str(), nil
}

func GetWireguardServer(page *rod.Page, config *util.Config, token string) (string, error) {
	privateKey, publicKey, err := wireguard.GenerateKeyPair()
	if err != nil {
		return "", err
	}
	if config.Debug {
		fmt.Printf("Generated keys, publicKey: %v, privateKey: %v\n", publicKey, privateKey)
	}

	params := url.Values{}
	params.Add("sUsername", config.Subscription.Username)
	params.Add("sPassword", config.Subscription.Password)
	params.Add("sCountrySlug", config.Server.Country)
	params.Add("sDeviceType", config.Device)
	params.Add("sClientPublicKey", publicKey)
	params.Add("iCityId", strconv.Itoa(config.Server.City))
	params.Add("sSubsId", config.Subscription.ID)
	// There's some port forwarding logic for this variable, if you need this, feel free to create a PR
	// const hasPF = currentSubscriptionData?.add_ons?.some(addOn => [addonConstants.port_forwarding_with_dedicated_ip, addonConstants.port_forwarding].includes(addOn.code))  ? true : false;
	// const hasPFTTag = serverType.includes("PF")  ? true : false;
	// const natServerBool = (hasPF && hasPFTTag)? "0" : "1";
	params.Add("natServer", "1")

	if config.Debug {
		fmt.Printf("Requesting wireguard server with params: %v\n", params)
	}

	page.MustNavigate(_BASE_URL).MustWaitNavigation()
	res, err := page.Eval(`
		async (body, authorization, privateKey) => {
			const res = await fetch(
				"/v2/api/wireguard/get-wg-server",
				{
					method: "POST",
					body,
					headers: {
						'content-type': 'application/x-www-form-urlencoded',
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
			return json.body[0].wireguard_configuration
				.replace("{clientPrivateKey}", privateKey)
				.replace(/Endpoint=.*:51820/, "Endpoint="+json.body[0].protocol_ip+":51820")
	}`, params.Encode(), "Bearer "+token, privateKey)
	if err != nil {
		return "", err
	}

	return res.Value.Str(), nil
}
