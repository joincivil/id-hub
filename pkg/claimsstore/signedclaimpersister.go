package claimsstore

import (
	"encoding/hex"
	"encoding/json"
	"time"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
)

// SignedClaimPostgres represents the schema for signed claims
type SignedClaimPostgres struct {
	IssuanceDate      time.Time
	Type              claimtypes.CredentialType
	CredentialSubject postgres.Jsonb `gorm:"not null"`
	Issuer            string         `gorm:"not null;index:issuer"`
	Proof             postgres.Jsonb `gorm:"not null"`
	Hash              string         `gorm:"primary_key"`
}

// TableName sets the table name for signed claims
func (SignedClaimPostgres) TableName() string {
	return "signed_claims"
}

// ToCredential converts the db type to the model
func (c *SignedClaimPostgres) ToCredential() (claimtypes.Credential, error) {
	proof := &linkeddata.Proof{}
	err := json.Unmarshal(c.Proof.RawMessage, proof)
	if err != nil {
		return nil, err
	}

	if c.Type == claimtypes.ContentCredentialType {
		credential := &claimtypes.ContentCredential{
			Type:    []claimtypes.CredentialType{claimtypes.VerifiableCredentialType, claimtypes.ContentCredentialType},
			Context: []string{"https://www.w3.org/2018/credentials/v1", "https://id.civil.co/credentials/contentcredential/v1"},
			Issuer:  c.Issuer,
			CredentialSchema: claimtypes.CredentialSchema{
				ID:   "https://id.civil.co/credentials/schemas/v1/metadata.json",
				Type: "JsonSchemaValidator2018",
			},
			Proof: []interface{}{*proof},
		}

		credSubj := &claimtypes.ContentCredentialSubject{}
		err = json.Unmarshal(c.CredentialSubject.RawMessage, credSubj)

		if err != nil {
			return nil, err
		}

		credential.CredentialSubject = *credSubj

		return credential, nil
	} else if c.Type == claimtypes.LicenseCredentialType {
		credential := &claimtypes.LicenseCredential{
			Type: []claimtypes.CredentialType{
				claimtypes.VerifiableCredentialType,
				claimtypes.LicenseCredentialType,
			},
			Context: []string{
				"https://www.w3.org/2018/credentials/v1",
				"https://id.civil.co/credentials/licensecredential/v1",
			},
			Issuer: c.Issuer,
			Proof:  []interface{}{*proof},
		}
		credSubj := make([]interface{}, 0)
		err = json.Unmarshal(c.CredentialSubject.RawMessage, &credSubj)

		if err != nil {
			return nil, err
		}

		credential.CredentialSubject = credSubj
		return credential, nil
	}

	return nil, errors.New("unsupported credential type")
}

func hashCred(cred claimtypes.Credential) (string, error) {
	credJSON, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256(credJSON)
	mHash, err := multihash.EncodeName(hash, "keccak-256")
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mHash), nil
}

// FromContentCredential populates the db type from a model
func (c *SignedClaimPostgres) FromContentCredential(cred *claimtypes.ContentCredential) error {
	c.Issuer = cred.Issuer
	c.IssuanceDate = cred.IssuanceDate
	c.Type = claimtypes.ContentCredentialType
	credSubjJSON, err := json.Marshal(cred.CredentialSubject)
	if err != nil {
		return err
	}
	linkedDataProof, err := cred.FindLinkedDataProof()
	if err != nil {
		return err
	}
	proofJSON, err := json.Marshal(linkedDataProof)
	if err != nil {
		return err
	}
	c.CredentialSubject = postgres.Jsonb{RawMessage: credSubjJSON}
	c.Proof = postgres.Jsonb{RawMessage: proofJSON}

	c.Hash, err = hashCred(cred)
	if err != nil {
		return err
	}

	return nil
}

// FromLicenseCredential populates a signed claim postgres model from a license claim
func (c *SignedClaimPostgres) FromLicenseCredential(cred *claimtypes.LicenseCredential) error {
	c.Issuer = cred.Issuer
	c.IssuanceDate = cred.IssuanceDate
	c.Type = claimtypes.LicenseCredentialType
	credSubjJSON, err := json.Marshal(cred.CredentialSubject)
	if err != nil {
		return err
	}
	linkedDataProof, err := cred.FindLinkedDataProof()
	if err != nil {
		return err
	}
	proofJSON, err := json.Marshal(linkedDataProof)
	if err != nil {
		return err
	}
	c.CredentialSubject = postgres.Jsonb{RawMessage: credSubjJSON}
	c.Proof = postgres.Jsonb{RawMessage: proofJSON}

	c.Hash, err = hashCred(cred)
	if err != nil {
		return err
	}

	return nil
}

// SignedClaimPGPersister persister model for signed claims
type SignedClaimPGPersister struct {
	db *gorm.DB
}

// NewSignedClaimPGPersister returns a new SignedClaimPGPersister
func NewSignedClaimPGPersister(db *gorm.DB) *SignedClaimPGPersister {
	return &SignedClaimPGPersister{
		db: db,
	}
}

// AddCredential takes a credential and adds it to the db
func (p *SignedClaimPGPersister) AddCredential(claim claimtypes.Credential) (string, error) {
	signedClaim := &SignedClaimPostgres{}
	switch val := claim.(type) {
	case *claimtypes.ContentCredential:
		err := signedClaim.FromContentCredential(val)
		if err != nil {
			return "", errors.Wrapf(err, "addcredential.fromcontentcred: hash: %v, sub: %v",
				signedClaim.Hash, string(signedClaim.CredentialSubject.RawMessage))
		}
	case *claimtypes.LicenseCredential:
		err := signedClaim.FromLicenseCredential(val)
		if err != nil {
			return "", errors.Wrapf(err, "addcredential.fromlicencecred: hash: %v, sub: %v",
				signedClaim.Hash, string(signedClaim.CredentialSubject.RawMessage))
		}
	}
	if err := p.db.Create(signedClaim).Error; err != nil {
		return "", errors.Wrapf(err, "addcredential.dbcreate: hash: %v, sub: %v",
			signedClaim.Hash, string(signedClaim.CredentialSubject.RawMessage))
	}
	// Debug remove!
	log.Infof("addcredential.success: hash: %v", signedClaim.Hash)
	return signedClaim.Hash, nil
}

// GetCredentialByHash returns a credential from a hash taken from the associated merkle tree claim
func (p *SignedClaimPGPersister) GetCredentialByHash(hash string) (claimtypes.Credential,
	error) {
	bytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	mHash, err := multihash.EncodeName(bytes, "keccak-256")
	if err != nil {
		return nil, err
	}
	mHashString := hex.EncodeToString(mHash)
	return p.GetCredentialByMultihash(mHashString)
}

// GetCredentialByMultihash returns a credential from a multihash
func (p *SignedClaimPGPersister) GetCredentialByMultihash(mHash string) (claimtypes.Credential,
	error) {
	signedClaim := &SignedClaimPostgres{}
	if err := p.db.Where(&SignedClaimPostgres{Hash: mHash}).First(signedClaim).Error; err != nil {
		return nil, err
	}
	return signedClaim.ToCredential()
}
