package purevpn

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// Generated from JWT payload

type JWT struct {
	Aud                string        `json:"aud"`
	Exp                int64         `json:"exp"`
	Iat                int64         `json:"iat"`
	Iss                string        `json:"iss"`
	Sub                string        `json:"sub"`
	Jti                string        `json:"jti"`
	AuthenticationType string        `json:"authenticationType"`
	Email              string        `json:"email"`
	EmailVerified      bool          `json:"email_verified"`
	ApplicationID      string        `json:"applicationId"`
	Roles              []interface{} `json:"roles"`
	AuthTime           int64         `json:"auth_time"`
	Tid                string        `json:"tid"`
	EntityGrants       []interface{} `json:"entity_grants"`
	User               User          `json:"user"`
}

type User struct {
	Active                             bool          `json:"active"`
	BirthDate                          string        `json:"birthDate"`
	BreachedPasswordLastCheckedInstant int64         `json:"breachedPasswordLastCheckedInstant"`
	BreachedPasswordStatus             string        `json:"breachedPasswordStatus"`
	ConnectorID                        string        `json:"connectorId"`
	Data                               UserData      `json:"data"`
	Email                              string        `json:"email"`
	FirstName                          string        `json:"firstName"`
	ID                                 string        `json:"id"`
	InsertInstant                      int64         `json:"insertInstant"`
	LastLoginInstant                   int64         `json:"lastLoginInstant"`
	LastName                           string        `json:"lastName"`
	LastUpdateInstant                  int64         `json:"lastUpdateInstant"`
	Memberships                        []interface{} `json:"memberships"`
	PasswordChangeRequired             bool          `json:"passwordChangeRequired"`
	PasswordLastUpdateInstant          int64         `json:"passwordLastUpdateInstant"`
	PreferredLanguages                 []interface{} `json:"preferredLanguages"`
	TenantID                           string        `json:"tenantId"`
	TwoFactor                          TwoFactor     `json:"twoFactor"`
	UsernameStatus                     string        `json:"usernameStatus"`
	Verified                           bool          `json:"verified"`
}

type UserData struct {
	AccountCode          string                       `json:"accountCode"`
	BillingType          string                       `json:"billingType"`
	Email                string                       `json:"email"`
	IsMigratedToEmail    int64                        `json:"isMigratedToEmail"`
	MaType               int64                        `json:"maType"`
	Subscription         map[string][]JWTSubscription `json:"subscription"`
	Type                 string                       `json:"type"`
	IsMAAutoLoginAllowed bool                         `json:"isMAAutoLoginAllowed"`
	IsDomeUser           bool                         `json:"isDomeUser"`
}

type JWTSubscription struct {
	Addons           []interface{} `json:"addons"`
	BillingCycle     string        `json:"billingCycle"`
	Expiry           string        `json:"expiry"`
	ExpiryReason     string        `json:"expiryReason"`
	HostingID        string        `json:"hostingId"`
	PaymentGateway   string        `json:"paymentGateway"`
	Plan             string        `json:"plan"`
	ServiceOrigin    string        `json:"service_origin"`
	Status           string        `json:"status"`
	SubscriptionType string        `json:"subscription_type"`
	Vpnusername      string        `json:"vpnusername"`
}

type TwoFactor struct {
	Methods       []interface{} `json:"methods"`
	RecoveryCodes []interface{} `json:"recoveryCodes"`
}

// Simply parse jwt token without any validation
func parseJWT(jwtString string) (*JWT, error) {
	jwt := JWT{}
	parts := strings.Split(jwtString, ".") // [header, payload, signature]
	byt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(byt, &jwt); err != nil {
		return nil, err
	}
	return &jwt, nil
}

func GetUserData(jwtString string) (*UserData, error) {
	if jwt, err := parseJWT(jwtString); err != nil {
		return nil, err
	} else {
		return &jwt.User.Data, nil
	}
}
