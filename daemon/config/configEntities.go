package config

// Config is the main configuration data structure
type Config struct {
	HomeAssistant HomeAssistantConfig      `yaml:"home_assistant"`
	People        map[string]*PeopleConfig `yaml:"people"`
}

// HomeAssistantConfig is the configuration for the Home Assistant platform integration
type HomeAssistantConfig struct {
	IP    string `yaml:"ip"`
	SSL   bool   `yaml:"ssl"`
	Token string `yaml:"token"`
}

// PeopleConfig is the configuration for the Home Assistant platform integration
type PeopleConfig struct {
	FriendlyName string   `yaml:"friendly_name"`
	Devices      []string `yaml:"devices"`
	State        string
	Attributes   map[string]interface{}
}
