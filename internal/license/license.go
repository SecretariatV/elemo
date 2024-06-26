package license

import (
	"errors"
	"time"

	"github.com/goccy/go-json"

	"github.com/hyperboloide/lk"
	"github.com/rs/xid"
)

const (
	DefaultValidityPeriod = 365 // default validity period in days

	FeatureComponents        Feature = "components"         // allow components
	FeatureCustomStatuses    Feature = "custom_statuses"    // allow custom statuses
	FeatureCustomFields      Feature = "custom_fields"      // allow custom fields
	FeatureMultipleAssignees Feature = "multiple_assignees" // allow multiple assignees per task
	FeatureReleases          Feature = "releases"           // allow releases

	QuotaDocuments     Quota = "documents"     // number of documents
	QuotaNamespaces    Quota = "namespaces"    // number of namespaces
	QuotaOrganizations Quota = "organizations" // number of organizations
	QuotaProjects      Quota = "projects"      // number of projects
	QuotaRoles         Quota = "roles"         // number of roles
	QuotaUsers         Quota = "users"         // number of users
)

var (
	ErrLicenseExpired          = errors.New("license expired")                     // license is expired
	ErrLicenseInvalid          = errors.New("invalid or expired license provided") // license is expired
	ErrLicenseInvalidSignature = errors.New("invalid license signature")           // license signature is invalid
	ErrNoLicense               = errors.New("no license provided")                 // no license provided

	// DefaultFeatures is the default set of features for a license.
	DefaultFeatures = []Feature{
		FeatureComponents,
		FeatureCustomStatuses,
		FeatureCustomFields,
		FeatureMultipleAssignees,
		FeatureReleases,
	}

	// DefaultQuotas is the default set of quotas for a license.
	DefaultQuotas = map[Quota]uint32{
		QuotaDocuments:     10,
		QuotaNamespaces:    1,
		QuotaOrganizations: 1,
		QuotaProjects:      10,
		QuotaRoles:         5,
		QuotaUsers:         5,
	}
)

// Feature represents a license feature.
type Feature string

// Quota represents a license quota.
type Quota string

// License represents the license information.
//
// The license information is parsed from the license key. The license key is
// generated by the license server. The license is valid for a specific
// organization for a given period.
//
// The license is validated locally by the hardcoded public key. The public key
// is used to verify the signature of the license key.
type License struct {
	ID           xid.ID           `json:"id"`           // license id
	Email        string           `json:"email"`        // organization email
	Organization string           `json:"organization"` // organization name
	Quotas       map[Quota]uint32 `json:"quotas"`       // quotas
	Features     []Feature        `json:"features"`     // features
	ExpiresAt    time.Time        `json:"expires_at"`   // expiration time
}

// Valid validates the license fields and returns false if any of the
// required license fields are missing, quotas does not meet minimum
// expectations, or the license is expired.
func (l *License) Valid() bool {
	if len(l.Features) == 0 || len(l.Quotas) == 0 {
		return false
	}

	for _, quota := range []Quota{QuotaDocuments, QuotaNamespaces, QuotaOrganizations, QuotaProjects, QuotaRoles, QuotaUsers} {
		if val, ok := l.Quotas[quota]; !ok || val == 0 {
			return false
		}
	}

	return l.ID != xid.NilID() &&
		l.Email != "" &&
		l.Organization != "" &&
		!l.Expired()
}

// Expired returns true if the license is expired.
func (l *License) Expired() bool {
	return l.ExpiresAt.Before(time.Now().UTC())
}

// HasFeature returns true if the license has the given feature.
func (l *License) HasFeature(feature Feature) bool {
	for _, f := range l.Features {
		if f == feature {
			return true
		}
	}

	return false
}

// WithinThreshold returns true if the given value is within the threshold.
func (l *License) WithinThreshold(quota Quota, value int) bool {
	return int(l.Quotas[quota])-value >= 0
}

// NewLicense validates the license key.
//
// If the license is valid, the validated license is returned. Otherwise, an
// error is returned.
func NewLicense(licenseKey, pubKey string) (*License, error) {
	var err error

	key, err := lk.PublicKeyFromB32String(pubKey)
	if err != nil {
		return nil, errors.Join(ErrLicenseInvalid, err)
	}

	licenseData, err := lk.LicenseFromB32String(licenseKey)
	if err != nil {
		return nil, errors.Join(ErrLicenseInvalid, err)
	}

	if isLicenseValid, err := licenseData.Verify(key); err != nil || !isLicenseValid {
		if err != nil {
			return nil, errors.Join(ErrLicenseInvalid, err)
		}

		return nil, errors.Join(ErrLicenseInvalidSignature, err)
	}

	license := new(License)
	if err = json.Unmarshal(licenseData.Data, license); err != nil {
		return nil, errors.Join(ErrLicenseInvalid, err)
	}

	if !license.Valid() {
		if license.Expired() {
			return nil, ErrLicenseExpired
		}

		return nil, ErrLicenseInvalid
	}

	return license, nil
}
