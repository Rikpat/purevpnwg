package util

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

type PartitionKey struct {
	Value interface{}
}

func (pk *PartitionKey) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		pk.Value = str
		return nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		pk.Value = obj
		return nil
	}

	return fmt.Errorf("partitionKey is neither a string nor an object")
}

type CustomNetworkCookie struct {
	Name         string       `json:"name"`
	Value        string       `json:"value"`
	Domain       string       `json:"domain"`
	Path         string       `json:"path"`
	Expires      float64      `json:"expires"`
	Size         int          `json:"size"`
	HttpOnly     bool         `json:"httpOnly"`
	Secure       bool         `json:"secure"`
	Session      bool         `json:"session"`
	SameSite     string       `json:"sameSite"`
	PartitionKey PartitionKey `json:"partitionKey"`
}

func UnmarshalCookies(data []byte) ([]*proto.NetworkCookie, error) {
    var customCookies []*CustomNetworkCookie
    if err := json.Unmarshal(data, &customCookies); err != nil {
        return nil, err
    }

    var cookies []*proto.NetworkCookie
    for _, customCookie := range customCookies {
        cookie := &proto.NetworkCookie{
            Name:     customCookie.Name,
            Value:    customCookie.Value,
            Domain:   customCookie.Domain,
            Path:     customCookie.Path,
            Expires:  proto.TimeSinceEpoch(customCookie.Expires),
            Size:     customCookie.Size,
            Secure:   customCookie.Secure,
            Session:  customCookie.Session,
            SameSite: proto.NetworkCookieSameSite(customCookie.SameSite),
        }
        cookies = append(cookies, cookie)
    }

    return cookies, nil
}

func FilterCookies(cookies []*proto.NetworkCookie, namePrefix string) (filtered []*proto.NetworkCookie) {
    for _, cookie := range cookies {
        if strings.HasPrefix(cookie.Name, namePrefix) {
            filtered = append(filtered, cookie)
        }
    }
    return filtered
}

func AreCookiesExpired(cookies []*proto.NetworkCookie) bool {
    now := proto.TimeSinceEpoch(time.Now().Unix())
    for _, cookie := range cookies {
        if cookie.Expires != -1 && cookie.Expires < now {
            return true
        }
    }
    return false
}

func WriteCookies(filePath string, cookies []*proto.NetworkCookie) error {
    file, err := os.Create(filePath)
    if err == nil {
        encoder := gob.NewEncoder(file)
        err = encoder.Encode(cookies)
    }
    file.Close()
    return err
}

func ReadCookies(filePath string) ([]*proto.NetworkCookie, error) {
    cookies := []*proto.NetworkCookie{}
    file, err := os.Open(filePath)
    if err == nil {
        decoder := gob.NewDecoder(file)
        err = decoder.Decode(&cookies)
    }
    file.Close()
    return cookies, err
}