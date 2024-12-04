package common

type NetworkIDKey string

const (
	SteamNetworkIDKey     NetworkIDKey = "steam"
	FaceItNetworkIDKey    NetworkIDKey = "faceit"
	BattleNetNetworkIDKey NetworkIDKey = "battlenet"
)

type Network struct {
	ID          NetworkIDKey `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	WebSite     string       `json:"website"`
}

var (
	Steam = &Network{
		ID:   SteamNetworkIDKey,
		Name: "Steam",
	}
	BattleNet = &Network{
		ID:          BattleNetNetworkIDKey,
		Name:        "BattleNet",
		Description: "Blizzard Entertainment's online gaming service",
	}

	// TODO: rever redes ou inverter/configurar em db ???
)
