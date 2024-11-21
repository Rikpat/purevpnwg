package purevpn

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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
const _BASE_URL = "https://my.purevpn.com/v2/dashboard/security-tools"

var _PATH, _ = launcher.LookPath()
var _LAUNCHER = launcher.New().Bin(_PATH).Leakless(false).MustLaunch()

func Login(username, password string) (*rod.Page, []*proto.NetworkCookie) {
	browser := rod.New().Timeout(time.Minute).ControlURL(_LAUNCHER).MustConnect()

	page := stealth.MustPage(browser)

	page.MustNavigate(_BASE_URL).MustWaitNavigation()

	page.MustElement(`input#loginId`).MustWaitVisible().MustInput(username)
	page.MustElement(`input#password`).MustInput(password).MustType(input.Enter)

	page.MustWaitIdle()

	cookies := page.MustCookies()
	cookiesJSON, err := json.Marshal(cookies)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal cookies: %v", err))
	}

	customCookies, err := util.UnmarshalCookies(cookiesJSON)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal cookies: %v", err))
	}

	return page, customCookies
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
	params.Add("iCityId", config.Server.City)

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

func GetManualServer(page *rod.Page, config *util.Config, token string) (string, error) {
	params := url.Values{}
	params.Add("sUsername", config.Subscription.Username)
	params.Add("sCountrySlug", "NL")
	params.Add("iCityId", "2902")
	params.Add("sProtocolSlug1", "UDP")
	params.Add("aTagsFilter", "")

	if config.Debug {
		fmt.Printf("Requesting ovpn config with params: %v\n", params)
	}

	page.MustNavigate(_BASE_URL).MustWaitNavigation()
	res, err := page.Eval(`
        async (body, authorization) => {
            const res = await fetch(
                "/v2/api/manual/get-mc-server",
                {
                    method: "POST",
                    body: new URLSearchParams(body).toString(),
                    headers: {
                        'content-type': 'application/x-www-form-urlencoded',
                        'accept': 'application/json',
                        'authorization': authorization
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
            return json.body.configuration
        }
    `, params.Encode(), "Bearer "+token)
	if err != nil {
		return "", err
	}

	fmt.Printf("Adjusting auth-user-pass string in config\n")
	configStr := res.Value.Str()
	configStr = strings.Replace(configStr, "auth-user-pass", "auth-user-pass /vpn/login.conf", 1)
	// Remove extra spaces at the start of the lines
	configStr = strings.ReplaceAll(configStr, "\n ", "\n")

	return configStr, nil
}
