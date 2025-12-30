package request

// InstallAddOnRequest represents the request payload to install an add-on
type InstallAddOnRequest struct {
	Code       string `json:"code" example:"12345678" validate:"required"`
	Name       string `json:"name" example:"AnkiConnect" validate:"required"`
	Version    string `json:"version" example:"1.0.0"`
	ConfigJSON string `json:"config_json" example:"{}"`
}

// UpdateAddOnConfigRequest represents the request payload to update add-on config
type UpdateAddOnConfigRequest struct {
	ConfigJSON string `json:"config_json" example:"{\"port\": 8765}" validate:"required"`
}

// ToggleAddOnRequest represents the request payload to enable/disable an add-on
type ToggleAddOnRequest struct {
	Enabled bool `json:"enabled"`
}

