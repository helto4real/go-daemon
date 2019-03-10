package config

// Config is the main configuration data structure
type Config struct {
	HomeAssistant HomeAssistantConfig      `yaml:"home_assistant"`
	Settings      *SettingsConfig          `yaml:"settings"`
	People        map[string]*PeopleConfig `yaml:"people"`
}

// HomeAssistantConfig is the configuration for the Home Assistant platform integration
type HomeAssistantConfig struct {
	IP    string `yaml:"ip"`
	SSL   bool   `yaml:"ssl"`
	Token string `yaml:"token"`
}

type TrackingStateSettingsConfig struct {
	JustArrivedTime  int    `yaml:"just_arrived_time"`
	JustLeftTime     int    `yaml:"just_left_time"`
	HomeState        string `yaml:"home_state"`
	JustLeftState    string `yaml:"just_left_state"`
	JustArrivedState string `yaml:"just_arrived_state"`
	AwayState        string `yaml:"away_state"`
}

// SettingsConfig let you tweak the settings of the daemon
type SettingsConfig struct {
	TrackingSettings TrackingStateSettingsConfig `yaml:"tracking"`
}

// PeopleConfig is the configuration for the Home Assistant platform integration
type PeopleConfig struct {
	FriendlyName string   `yaml:"friendly_name"`
	Devices      []string `yaml:"devices"`
	State        string
	Attributes   map[string]interface{}
}
