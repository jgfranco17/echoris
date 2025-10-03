package system

type HealthStatus struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

type ServiceInfo struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Uptime      string `json:"uptime"`
}

type BasicErrorInfo struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
