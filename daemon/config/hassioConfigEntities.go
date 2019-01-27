package config

type HassioOptionsConfig struct {
	LogLevel string               `json:"log_level"`
	Persons  []HassioOptionPerson `json:"persons"`
}

type HassioOptionPerson struct {
	ID           string   `json:"id"`
	FriendlyName string   `json:"friendly_name"`
	Devices      []string `json:"devices"`
}
