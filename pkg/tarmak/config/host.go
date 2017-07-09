package config

type Host struct {
	ID             string
	Host           string
	HostnamePublic bool
	Hostname       string
	Aliases        []string
	Roles          []string
	User           string
}
