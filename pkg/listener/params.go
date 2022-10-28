package listener

// ParametersListener contains the definition of the parameters used by the Listener
type Parameters struct {
	// Filters is a json string which sets startup filters
	Filters string `default:"" usage:"startup filters from env or config.json in a string format"`
}
