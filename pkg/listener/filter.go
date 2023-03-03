package listener

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Filter struct {
	Tag            string `json:"tag" validate:"required"`
	PublicKey      string `json:"publicKey,omitempty"`
	Id             string `json:"id,omitempty"`
	BucketName     string `json:"bucketName,omitempty"`
	WithPOI        bool   `json:"withPOI,omitempty"`
	Duration       string `json:"duration,omitempty"`
	Expiration     time.Time
	PublicKeyBytes []byte
}

type StartupFilters struct {
	Filters []Filter `json:"filters"`
}

func NewFilter(tag string, publicKey string, bucketName string, duration string, withPOI bool) (Filter, error) {
	filter := Filter{
		Tag:        tag,
		PublicKey:  publicKey,
		BucketName: bucketName,
		WithPOI:    withPOI,
		Duration:   duration,
	}

	if filter.PublicKey != "" {
		err := filter.setPublicKeyBytes()
		if err != nil {
			return Filter{}, err
		}
	}

	return filter, nil
}

func (f *Filter) setId() {
	f.Id = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", f))))
}

func (f *Filter) setPublicKeyBytes() error {
	publicKeyBytes, err := hex.DecodeString(f.PublicKey)
	if err != nil {
		return err
	}

	f.PublicKeyBytes = publicKeyBytes
	return nil
}

func (f *Filter) setExpiration() error {
	durationParsed, err := time.ParseDuration(f.Duration)
	if err != nil {
		// return an expired filter
		f.Expiration = time.Now()
		return err
	}
	f.Expiration = time.Now().Add(durationParsed)
	return nil
}

func (f *Filter) IsExpired() bool {
	return time.Now().After(f.Expiration)
}

func UnmarshalStartupFilters(filtersString string) ([]Filter, error) {
	var filters StartupFilters

	// unmarshal filters
	err := json.Unmarshal([]byte(filtersString), &filters)
	if err != nil {
		return filters.Filters, err
	}

	// validate filters
	for _, filter := range filters.Filters {
		err = validator.New().Struct(filter)
		if err != nil {
			return filters.Filters, err
		}
		if filter.PublicKey != "" {
			filter.setPublicKeyBytes()
		}
	}
	return filters.Filters, nil
}
