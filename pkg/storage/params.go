package storage

// ParametersRestAPI contains the definition of the parameters used by the Collector to access the S3 storage
type Parameters struct {
	// Endpoint defines the endpoint for the S3 storage
	Endpoint string `default:"" usage:"the storage endpoint"`

	// AccessId defines the access id for the S3 storage
	AccessKeyID string `default:"" usage:"the access id for the storage"`

	// Password defines the password for the given access id of the S3 storage
	SecretAccessKey string `default:"" usage:"the password for the given access id"`

	// DefaultBucketName sets the default bucket's name
	DefaultBucketName string `default:"shimmer-mainnet-default" usage:"sets the default bucket's name"`

	// DefaultBucketExpirationDays sets the default bucket's expiration days
	DefaultBucketExpirationDays int `default:"30" usage:"sets the default bucket's expiration days"`

	// Region defines the region of the S3 storage
	Region string `default:"eu-south-1" usage:"defines the region of the S3 storage"`

	// ObjectExtension sets the file extension for the object inside the storage
	ObjectExtension string `default:"" usage:"sets the file extension for the object inside the storage"`

	// Secure defines whether the connection to S3 storage should be secure
	Secure bool `default:"true" usage:"whether the connection to storage should be secure"`
}
