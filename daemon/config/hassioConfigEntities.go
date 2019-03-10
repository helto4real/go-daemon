package config

type HassioOptionsConfig struct {
	LogLevel string                      `json:"log_level"`
	Settings *HassioOptionSettingsConfig `json:"settings"`
	Persons  []HassioOptionPerson        `json:"persons"`
}

type HassioOptionTrackerStatesConfig struct {
	JustArrivedTime  int    `json:"just_arrived_time"`
	JustLeftTime     int    `json:"just_left_time"`
	HomeState        string `json:"home_state"`
	JustLeftState    string `json:"just_left_state"`
	JustArrivedState string `json:"just_arrived_state"`
	AwayState        string `json:"away_state"`
}

// HassioOptionSettingsConfig let you tweak the settings of the daemon
type HassioOptionSettingsConfig struct {
	TrackingSettings *HassioOptionTrackerStatesConfig `json:"tracking"`
}

type HassioOptionPerson struct {
	ID           string   `json:"id"`
	FriendlyName string   `json:"friendly_name"`
	Devices      []string `json:"devices"`
}
