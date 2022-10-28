package api

// ParametersRestAPI contains the definition of the parameters used by the Collector HTTP server.
type Parameters struct {
	// BindAddress defines the bind address on which the Collector HTTP server listens.
	BindAddress string `default:"localhost:9030" usage:"the bind address on which the Collector HTTP server listens"`

	// AdvertiseAddress defines the address of the Collector HTTP server which is advertised to the INX Server (optional).
	AdvertiseAddress string `default:"" usage:"the address of the Collector HTTP server which is advertised to the INX Server (optional)"`

	// DebugRequestLoggerEnabled defines whether the debug logging for requests should be enabled
	DebugRequestLoggerEnabled bool `default:"false" usage:"whether the debug logging for requests should be enabled"`
}
