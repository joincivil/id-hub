package idhubmain

import (
	"github.com/ethereum/go-ethereum"
	log "github.com/golang/glog"
	"github.com/iden3/go-iden3-core/db"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/pubsub"
	"github.com/joincivil/id-hub/pkg/utils"
)

func initDidService(resolvers []did.Resolver) *did.Service {
	return did.NewService(resolvers)
}

func initClaimsService(treeStore *claimsstore.PGStore, signedClaimStore *claimsstore.SignedClaimPGPersister,
	didService *did.Service, rootService *claims.RootService, dlock lock.DLock) (*claims.Service, error) {
	return claims.NewService(treeStore, signedClaimStore, didService, rootService, dlock)
}

func initJWTClaimService(didJWTService *didjwt.Service,
	jwtPersister *claimsstore.JWTClaimPGPersister,
	claimService *claims.Service, natsService *pubsub.NatsService) *claims.JWTService {
	return claims.NewJWTService(didJWTService, jwtPersister, claimService, natsService)
}

func initRootService(config *utils.IDHubConfig, ethHelper *eth.Helper,
	treeStore db.Storage, persister *claimsstore.RootCommitsPGPersister) (*claims.RootService, error) {
	if config.RootCommitsAddress == "" {
		log.Errorf("No root commits address set, disabling root commits access")
		return nil, nil
	}

	rootCommitter, err := claims.NewRootCommitter(
		ethHelper,
		ethHelper.Blockchain.(ethereum.TransactionReader),
		config.RootCommitsAddress,
	)
	if err != nil {
		return nil, err
	}
	return claims.NewRootService(treeStore, rootCommitter, persister)
}

func initServices(db *gorm.DB, config *utils.IDHubConfig) (*claims.JWTService, *did.Service, *claims.Service, *didjwt.Service) {
	// DID init
	// Universal Resolver
	resolver, err := initHTTPUniversalResolver(config)
	if err != nil {
		log.Fatalf("error initializing universal resolver")
	}
	// EthURI Resolver
	ethURIResolver, err := initEthURIResolver(db)
	if err != nil {
		log.Fatalf("error initializing ethuri resolver")
	}

	sc, err := initializeNats(config)
	if err != nil {
		log.Fatalf("error initializing nats: %v", err)
	}

	// TODO(PN): Adding ethuri resolver during transition of enterprise clients
	// to other DID methods. Once this occurs, should remove it.
	didService := initDidService([]did.Resolver{resolver, ethURIResolver})
	didJWTService := didjwt.NewService(didService)

	// Claims init
	treePersister := initTreePersister(db)
	signedClaimPersister := initSignedClaimPersister(db)
	rootPersister := initRootClaimPersister(db)
	jwtClaimPersister := initJWTClaimPersister(db, didJWTService)
	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("error initializing eth helper: %v", err)
	}
	rootService, err := initRootService(config, ethHelper, treePersister, rootPersister)
	if err != nil {
		log.Fatalf("error initializing root service: %v", err)
	}
	dlock := initDLock(config)
	claimsService, err := initClaimsService(
		treePersister,
		signedClaimPersister,
		didService,
		rootService,
		dlock,
	)
	if err != nil {
		log.Fatalf("error initializing claims service")
	}

	jwtService := initJWTClaimService(
		didJWTService,
		jwtClaimPersister,
		claimsService,
		sc,
	)

	return jwtService, didService, claimsService, didJWTService
}
