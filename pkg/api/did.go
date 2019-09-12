package api

import (
	context "context"
	"time"

	log "github.com/golang/glog"

	"github.com/golang/protobuf/ptypes"
	pwrap "github.com/golang/protobuf/ptypes/wrappers"
	didlib "github.com/ockam-network/did"
	"google.golang.org/appengine/log"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/did"
	gapi "github.com/joincivil/id-hub/pkg/generated/api"
)

// NewDidImplementedServer is a convenience function that returns a new DidImplementedServer
// given it's dependencies
func NewDidImplementedServer(didService *did.Service) *DidImplementedServer {
	return &DidImplementedServer{
		didService: didService,
	}
}

// DidImplementedServer implements the DidServer interface
type DidImplementedServer struct {
	didService *did.Service
}

// Get implements the Get func in the DidServer interface
func (d *DidImplementedServer) Get(ctx context.Context, req *gapi.DidGetRequest) (*gapi.DidGetResponse, error) {
	requestedDid := req.Did
	if requestedDid == "" {
		return nil, status.Error(codes.InvalidArgument, "did is empty string")
	}

	doc, err := d.didService.GetDocument(requestedDid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get document for did")
	}

	if doc == nil {
		return nil, status.Error(codes.NotFound, "document is not found")
	}

	pbDoc, err := DocToPbDoc(doc)
	if err != nil {
		return nil, errors.Wrap(err, "error converting doc to pbdoc in get")
	}

	return &gapi.DidGetResponse{
		Doc: pbDoc,
	}, nil
}

// Save implements Save func in the DidServer interface
func (d *DidImplementedServer) Save(ctx context.Context, req *gapi.DidSaveRequest) (
	*gapi.DidSaveResponse, error) {
	// TODO(PN): Auth needed here. Only DID owner (or authorized users in DID doc)
	// can save and update.

	theDid := req.Did
	if theDid == "" {
		return nil, status.Error(codes.InvalidArgument, "did is empty string")
	}

	// Validate the DID
	if !did.ValidDid(theDid) {
		return nil, status.Error(codes.InvalidArgument, "did is invalid")
	}

	// Validate all the PKs in the public key list
	var pk *did.DocPublicKey
	pkMap := map[string]int{}
	for _, pbPk := range req.Doc.PublicKeys {
		pk = PbPublicKeyToPublicKey(pbPk)
		if !did.ValidDocPublicKey(pk) {
			return nil, status.Errorf(codes.InvalidArgument, "pk is invalid: %v", pk.ID.String())
		}
		pkMap[pk.ID.String()] = 1
	}

	// Validate all the PKs in the authentications key list
	var auth *did.DocAuthenicationWrapper
	var authDid string
	for _, pbAuth := range req.Doc.Authentications {
		auth = PbAuthToAuth(pbAuth)
		authDid = auth.DocPublicKey.ID.String()
		if auth.IDOnly {
			_, ok := pkMap[authDid]
			if !ok {
				return nil, status.Errorf(codes.InvalidArgument, "auth pk is not in public keys: %v", authDid)
			}
		} else if !did.ValidDocPublicKey(&auth.DocPublicKey) {
			return nil, status.Errorf(codes.InvalidArgument, "auth pk is invalid: %v", authDid)
		}
	}

	doc, err := d.didService.GetDocument(theDid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get document for did")
	}

	// If doc is not found for DID, create a new one with a first key
	if doc == nil {
		// Validate the first public key, ignore all others for now.
		// If valid, then save it
	}

	// If doc is found for DID, add the new key

	// doc := PbDocToDoc(req.Doc)
	// err := d.didService.SaveDocument(doc)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, "error saving did document")
	// }

	return nil, nil
}

// DocToPbDoc converts a core DID document to Protobufs DID document
func DocToPbDoc(doc *did.Document) (*gapi.DidDocument, error) {
	var err error
	pbDoc := &gapi.DidDocument{}

	pbDoc.Id = doc.ID.String()
	pbDoc.Context = doc.Context

	if doc.Controller != nil {
		pbDoc.Controller = doc.Controller.String()
	}
	if doc.Created != nil {
		pbDoc.Created = convertTimeToPbTimestamp(doc.Created)
	}
	if doc.Updated != nil {
		pbDoc.Updated = convertTimeToPbTimestamp(doc.Updated)
	}

	if doc.PublicKeys != nil {
		pbKeys := make([]*gapi.DidDocPublicKey, len(doc.PublicKeys))
		for ind, pk := range doc.PublicKeys {
			pbKeys[ind] = PublicKeyToPbPublicKey(&pk)
		}
		pbDoc.PublicKeys = pbKeys
	}

	if doc.Authentications != nil {
		pbAuths := make([]*gapi.DidDocAuthentication, len(doc.Authentications))
		for ind, a := range doc.Authentications {
			pbAuths[ind] = AuthToPbAuth(&a)
		}
		pbDoc.Authentications = pbAuths
	}

	if doc.Services != nil {
		var pbServ *gapi.DidDocService
		pbServs := make([]*gapi.DidDocService, len(doc.Services))
		for ind, a := range doc.Services {
			pbServ, err = ServiceToPbService(&a)
			if err != nil {
				return nil, errors.Wrap(err, "error converting service to pb service")
			}
			pbServs[ind] = pbServ
		}
		pbDoc.Services = pbServs
	}

	if doc.Proof != nil {
		pbDoc.Proof = LdProofToPbLdProof(doc.Proof)
	}

	return pbDoc, nil
}

// PbDocToDoc converts a Protobuf DID document to core DID document
func PbDocToDoc(doc *gapi.DidDocument) *did.Document {
	return nil
}

// PublicKeyToPbPublicKey converts a core public key to a protobuf public key
func PublicKeyToPbPublicKey(pk *did.DocPublicKey) *gapi.DidDocPublicKey {
	pbPk := &gapi.DidDocPublicKey{}

	pbPk.Id = pk.ID.String()
	pbPk.Type = string(pk.Type)
	if pk.Controller != nil {
		pbPk.Controller = pk.Controller.String()
	}

	pbPk.PublicKeyPem = pk.PublicKeyPem
	pbPk.PublicKeyJwk = pk.PublicKeyJwk
	pbPk.PublicKeyHex = pk.PublicKeyHex
	pbPk.PublicKeyBase64 = pk.PublicKeyBase64
	pbPk.PublicKeyBase58 = pk.PublicKeyBase58
	pbPk.PublicKeyMultibase = pk.PublicKeyMultibase
	pbPk.EthereumAddress = pk.EthereumAddress

	return pbPk
}

// PbPublicKeyToPublicKey converts a protobuf public key to core public key
func PbPublicKeyToPublicKey(pbPk *gapi.DidDocPublicKey) *did.DocPublicKey {
	pk := &did.DocPublicKey{}

	d, err := didlib.Parse(pbPk.Id)
	if err != nil {
		log.Errorf("Did not parse DID properly: err: %v", err)
	}

	pk.ID = d
	pk.Type = did.LDSuiteType(pbPk.Type)

	if pbPk.Controller != "" {
		d, err = didlib.Parse(pbPk.Controller)
		if err != nil {
			log.Errorf("Did not parse controller DID properly: err: %v", err)
		}
		pk.Controller = d
	}

	pk.PublicKeyPem = pbPk.PublicKeyPem
	pk.PublicKeyJwk = pbPk.PublicKeyJwk
	pk.PublicKeyHex = pbPk.PublicKeyHex
	pk.PublicKeyBase64 = pbPk.PublicKeyBase64
	pk.PublicKeyBase58 = pbPk.PublicKeyBase58
	pk.PublicKeyMultibase = pbPk.PublicKeyMultibase
	pk.EthereumAddress = pbPk.EthereumAddress

	return pk
}

// AuthToPbAuth converts a core auth to a protobuf auth
func AuthToPbAuth(auth *did.DocAuthenicationWrapper) *gapi.DidDocAuthentication {
	pbAuth := &gapi.DidDocAuthentication{}

	pbAuth.PublicKey = PublicKeyToPbPublicKey(&auth.DocPublicKey)
	pbAuth.IdOnly = auth.IDOnly

	return pbAuth
}

// PbAuthToAuth converts a protobuf auth to a core auth struct
func PbAuthToAuth(pbAuth *gapi.DidDocAuthentication) *did.DocAuthenicationWrapper {
	pb := &did.DocAuthenicationWrapper{}

	pb.DocPublicKey = *PbPublicKeyToPublicKey(pbAuth.PublicKey)
	pb.IDOnly = pbAuth.IdOnly

	return pb
}

// ServiceToPbService converts a core service to a protobuf auth
func ServiceToPbService(serv *did.DocService) (*gapi.DidDocService, error) {
	pbServ := &gapi.DidDocService{}

	pbServ.Id = serv.ID.String()
	pbServ.Type = serv.Type
	pbServ.Description = serv.Description

	if serv.ServiceEndpointURI != nil {
		wrappedStr := &pwrap.StringValue{Value: *serv.ServiceEndpointURI}
		a, err := ptypes.MarshalAny(wrappedStr)
		if err != nil {
			return nil, errors.Wrap(err, "error marshalling any for string value")
		}
		pbServ.ServiceEndpoint = a
	}

	// XXX(PN): Not sure how to handle JSON-LD here using any.Any.
	// Perhaps need another proto type?
	// if serv.ServiceEndpointLD != nil {
	// }

	return pbServ, nil
}

// LdProofToPbLdProof converts a core linked data proof to a protobuf linked data
// proof
func LdProofToPbLdProof(proof *did.LinkedDataProof) *gapi.LinkedDataProof {
	ldp := &gapi.LinkedDataProof{}

	ldp.Type = proof.Type
	ldp.Creator = proof.Creator
	ldp.Created = convertTimeToPbTimestamp(&proof.Created)
	ldp.ProofValue = proof.ProofValue
	if proof.Domain != nil {
		ldp.Domain = *proof.Domain
	}
	if proof.Nonce != nil {
		ldp.Nonce = *proof.Domain
	}

	return ldp
}

func convertTimeToPbTimestamp(t *time.Time) *timestamp.Timestamp {
	nt := &timestamp.Timestamp{}
	nt.Seconds = t.Unix()
	nt.Nanos = int32(t.UnixNano())
	return nt
}

// func convertPbTimestampToTime(t *timestamp.Timestamp) *time.Time {
// 	tm := time.Unix(t.Seconds, 0)
// 	return &tm
// }
