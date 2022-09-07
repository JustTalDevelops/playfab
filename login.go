package playfab

import (
	"context"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"strings"
	"time"
)

const (
	// minecraftTitleID represents the PlayFab title ID for Minecraft: Bedrock Edition.
	minecraftTitleID = "20ca2"
	// minecraftDefaultSDK represents the usual SDK sent by the Minecraft client.
	minecraftDefaultSDK = "XPlatCppSdk-3.6.190304"
	// minecraftUserAgent represents the usual user agent sent by the Minecraft client.
	minecraftUserAgent = "libhttpclient/1.0.0.0"
)

// infoRequestParameters represent parameters that can be modified to request more information from the PlayFab API.
// By default, these are hardcoded to whatever the Minecraft client usually requests.
type infoRequestParameters struct {
	CharacterInventories bool `json:"GetCharacterInventories"`
	CharacterList        bool `json:"GetCharacterList"`
	PlayerProfile        bool `json:"GetPlayerProfile"`
	PlayerStatistics     bool `json:"GetPlayerStatistics"`
	TitleData            bool `json:"GetTitleData"`
	UserAccountInfo      bool `json:"GetUserAccountInfo"`
	UserData             bool `json:"GetUserData"`
	UserInventory        bool `json:"GetUserInventory"`
	UserReadOnlyData     bool `json:"GetUserReadOnlyData"`
	UserVirtualCurrency  bool `json:"GetUserVirtualCurrency"`
	PlayerStatisticNames any  `json:"PlayerStatisticNames"`
	ProfileConstraints   any  `json:"ProfileConstraints"`
	TitleDataKeys        any  `json:"TitleDataKeys"`
	UserDataKeys         any  `json:"UserDataKeys"`
	UserReadOnlyDataKeys any  `json:"UserReadOnlyDataKeys"`
}

// loginRequest is a request sent by the client to the PlayFab API to obtain a temporary login token.
type loginRequest struct {
	CreateAccount         bool                  `json:"CreateAccount"`
	EncryptedRequest      any                   `json:"EncryptedRequest"`
	InfoRequestParameters infoRequestParameters `json:"InfoRequestParameters"`
	PlayerSecret          any                   `json:"PlayerSecret"`
	TitleID               string                `json:"TitleId"`
	XboxToken             string                `json:"XboxToken"`
}

// loginResponse is a response sent by the PlayFab API to a login request.
type loginResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   struct {
		SessionTicket   string `json:"SessionTicket"`
		PlayFabID       string `json:"PlayFabId"`
		NewlyCreated    bool   `json:"NewlyCreated"`
		SettingsForUser struct {
			NeedsAttribution bool `json:"NeedsAttribution"`
			GatherDeviceInfo bool `json:"GatherDeviceInfo"`
			GatherFocusInfo  bool `json:"GatherFocusInfo"`
		} `json:"SettingsForUser"`
		LastLoginTime     time.Time `json:"LastLoginTime"`
		InfoResultPayload struct {
			AccountInfo struct {
				PlayFabID string    `json:"PlayFabId"`
				Created   time.Time `json:"Created"`
				TitleInfo struct {
					DisplayName        string    `json:"DisplayName"`
					Origination        string    `json:"Origination"`
					Created            time.Time `json:"Created"`
					LastLogin          time.Time `json:"LastLogin"`
					FirstLogin         time.Time `json:"FirstLogin"`
					IsBanned           bool      `json:"isBanned"`
					TitlePlayerAccount struct {
						ID         string `json:"Id"`
						Type       string `json:"Type"`
						TypeString string `json:"TypeString"`
					} `json:"TitlePlayerAccount"`
				} `json:"TitleInfo"`
				PrivateInfo struct {
				} `json:"PrivateInfo"`
				XboxInfo struct {
					XboxUserID      string `json:"XboxUserId"`
					XboxUserSandbox string `json:"XboxUserSandbox"`
				} `json:"XboxInfo"`
			} `json:"AccountInfo"`
			UserInventory           []any `json:"UserInventory"`
			UserDataVersion         int   `json:"UserDataVersion"`
			UserReadOnlyDataVersion int   `json:"UserReadOnlyDataVersion"`
			CharacterInventories    []any `json:"CharacterInventories"`
			PlayerProfile           struct {
				PublisherID string `json:"PublisherId"`
				TitleID     string `json:"TitleId"`
				PlayerID    string `json:"PlayerId"`
				DisplayName string `json:"DisplayName"`
			} `json:"PlayerProfile"`
		} `json:"InfoResultPayload"`
		EntityToken struct {
			EntityToken     string    `json:"EntityToken"`
			TokenExpiration time.Time `json:"TokenExpiration"`
			Entity          struct {
				ID         string `json:"Id"`
				Type       string `json:"Type"`
				TypeString string `json:"TypeString"`
			} `json:"Entity"`
		} `json:"EntityToken"`
		TreatmentAssignment struct {
			Variants  []any `json:"Variants"`
			Variables []any `json:"Variables"`
		} `json:"TreatmentAssignment"`
	} `json:"data"`
}

// entityData contains data about the entity, such as the entity ID or entity type.
type entityData struct {
	ID         string `json:"Id"`
	Type       string `json:"Type"`
	TypeString string `json:"TypeString,omitempty"`
}

// entityTokenRequest is sent by the client to the PlayFab API to request an entity token for the session.
type entityTokenRequest struct {
	Entity entityData `json:"Entity"`
}

// entityTokenResponse is a response sent by the PlayFab API to an entityTokenRequest.
type entityTokenResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   struct {
		EntityToken     string     `json:"EntityToken"`
		TokenExpiration time.Time  `json:"TokenExpiration"`
		Entity          entityData `json:"Entity"`
	} `json:"data"`
}

// acquireLoginToken acquires the temporary login token that will be used to acquire the entity token, using the Xbox
// Live token.
func (p *PlayFab) acquireLoginToken() error {
	token, err := p.src.Token()
	if err != nil {
		return err
	}
	t, err := auth.RequestXBLToken(context.Background(), token, "rp://playfabapi.com/")
	if err != nil {
		return err
	}

	var resp loginResponse
	if err = p.request(fmt.Sprintf("Client/LoginWithXbox?sdk=%s", minecraftDefaultSDK), loginRequest{
		CreateAccount: true,
		InfoRequestParameters: infoRequestParameters{
			PlayerProfile:   true,
			UserAccountInfo: true,
		},
		TitleID:   strings.ToUpper(minecraftTitleID),
		XboxToken: fmt.Sprintf("XBL3.0 x=%v;%v", t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash, t.AuthorizationToken.Token),
	}, &resp); err != nil {
		return err
	}

	p.id = resp.Data.PlayFabID
	p.token = resp.Data.EntityToken.EntityToken
	return nil
}

// acquireEntityToken acquires the entity token that will be used for the rest of the session, and updates the PlayFab
// instance with the new token.
func (p *PlayFab) acquireEntityToken() error {
	var resp entityTokenResponse
	if err := p.request(fmt.Sprintf("Authentication/GetEntityToken?sdk=%s", minecraftDefaultSDK), entityTokenRequest{Entity: entityData{
		ID:   p.id,
		Type: "master_player_account",
	}}, &resp); err != nil {
		return err
	}

	p.token = resp.Data.EntityToken
	return nil
}
