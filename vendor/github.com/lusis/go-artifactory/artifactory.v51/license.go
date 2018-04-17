package artifactory

import (
	"encoding/json"
)

// LicenseInformation represents the json response from artifactory for license information
type LicenseInformation struct {
	LicenseType  string `json:"type"`
	ValidThrough string `json:"validThrough"`
	LicensedTo   string `json:"licensedTo"`
}

// GetLicenseInformation returns license information from Artifactory
func (c *Client) GetLicenseInformation() (LicenseInformation, error) {
	o := make(map[string]string)
	var l LicenseInformation
	d, e := c.Get("/api/system/license", o)
	if e != nil {
		return l, e
	}
	err := json.Unmarshal(d, &l)
	return l, err
}
