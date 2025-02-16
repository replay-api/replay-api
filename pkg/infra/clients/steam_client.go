package clients

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	steamEntities "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamClient struct {
	HttpClient *http.Client
}

func NewSteamClient() *SteamClient {
	return &SteamClient{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: time.Second * 30,
			},
			Timeout: time.Second * 10,
		},
	}
}

// deprecated
func (c *SteamClient) Details(token string) (*steamEntities.SteamUser, error) {
	res, err := c.HttpClient.Get("https://api.steampowered.com/ISteamUserOAuth/GetTokenDetails/v1/?access_token=" + token)
	if err != nil {
		slog.Error("Failed to get token details", err)
		return nil, err
	}
	defer res.Body.Close()

	var steamUser steamEntities.SteamUser
	json.NewDecoder(res.Body).Decode(&steamUser)

	return &steamUser, nil
}
