package testinits

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// InitDIDService creates the did service for stesting
func InitDIDService(db *gorm.DB) (*did.Service, *ethuri.Service) {
	persister := ethuri.NewPostgresPersister(db)
	ethURIService := ethuri.NewService(persister)
	return did.NewService([]did.Resolver{ethURIService}), ethURIService
}

// AddDID adds a did to the db
func AddDID(ethURI *ethuri.Service, cService *claims.Service) (*didlib.DID, *ecdsa.PrivateKey, error) {
	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c"
	userDID, err := didlib.Parse(userDIDs)

	if err != nil {
		return nil, nil, err
	}

	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		return nil, nil, err
	}
	pubKey := secKey.Public().(*ecdsa.PublicKey)

	pubBytes := crypto.FromECDSAPub(pubKey)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(userDID)
	docPubKey.Controller = did.CopyDID(userDID)
	didDoc, err := ethuri.InitializeNewDocument(userDID, docPubKey, false, true)
	if err != nil {
		return nil, nil, err
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		return nil, nil, err
	}

	err = cService.CreateTreeForDIDWithPks(&didDoc.ID,
		[]*ecdsa.PublicKey{pubKey})
	if err != nil {
		return nil, nil, err
	}

	return userDID, secKey, nil
}
