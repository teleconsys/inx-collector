package poi

type Parameters struct {
	// HostUrl defines the address exposing the POI API.
	HostUrl string `default:"inx-poi:9687" usage:"the address exposing the POI API"`

	// IsPlugin defines wether the POI host is a POI plugin or a hornet node with an active plugin.
	IsPlugin bool `default:"true" usage:"wether the POI host is a POI plugin or a hornet node with an active plugin"`
}
