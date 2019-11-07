package idhubmain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/golang/glog"

	ctime "github.com/joincivil/go-common/pkg/time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/urfave/cli"
)

// RunCLI runs the idhub cli command
func RunCLI() error {
	app := cli.NewApp()
	app.Name = "idhubcli"
	app.Usage = "idhubcli"
	app.Version = "0.1"
	app.Commands = []cli.Command{
		*cmdGenerateDID(),
		*cmdGenerateNewKey(),
		*cmdGenerateGqlCreds(),
	}

	return app.Run(os.Args)
}

// cmdGenerateDID generates a base DID for an identity given their public key and public key
// type.  It will generate a new DID for this identity.
func cmdGenerateDID() *cli.Command {
	pubKeyTypeFlag := cli.StringFlag{
		Name:     "pktype, t",
		Usage:    "Sets the public key type for the initial public key",
		Required: true,
	}
	pubKeyFileFlag := cli.StringFlag{
		Name:     "pkfile, f",
		Usage:    "Set the full path to the public key file to use for the initial public key",
		Required: true,
	}
	storeFlag := cli.BoolFlag{
		Name:     "store, s",
		Usage:    "Store the new DID to the data store",
		Required: false,
	}
	storeHostFlag := cli.StringFlag{
		Name:     "host, o",
		Usage:    "Hostname of the Postgresql store",
		Value:    "localhost",
		Required: false,
	}
	storePortFlag := cli.StringFlag{
		Name:     "port, p",
		Usage:    "Port of the Postgresql store",
		Value:    "5423",
		Required: false,
	}
	storeDbnameFlag := cli.StringFlag{
		Name:     "dbname, d",
		Usage:    "DB name of the Postgresql store",
		Value:    "development",
		Required: false,
	}
	storeUsernameFlag := cli.StringFlag{
		Name:     "user, u",
		Usage:    "User of the Postgresql store",
		Required: false,
	}
	storePasswordFlag := cli.StringFlag{
		Name:     "password, w",
		Usage:    "Password of the Postgresql store",
		Required: false,
	}

	cmdFn := func(c *cli.Context) error {
		pkfile := c.String("pkfile")
		pktype := c.String("pktype")

		store := c.Bool("store")
		host := c.String("host")
		port := c.Int("port")
		dbname := c.String("dbname")
		user := c.String("user")
		pwd := c.String("password")

		var persister did.Persister
		if store {
			grm, err := NewGormPostgres(GormPostgresConfig{
				Host:     host,
				Port:     port,
				Dbname:   dbname,
				User:     user,
				Password: pwd,
			})
			if err != nil {
				log.Errorf("Error initializing GORM: err: %v", err)
			} else {
				grm.AutoMigrate(&did.PostgresDocument{})
				persister = did.NewPostgresPersister(grm)
			}
		}

		_, err := did.GenerateDIDCli(linkeddata.SuiteType(pktype), pkfile, persister)
		return err
	}

	return &cli.Command{
		Name:    "generatedid",
		Aliases: []string{"g"},
		Usage:   "Generates a simple new DID with an initial public key",
		Flags: []cli.Flag{
			pubKeyTypeFlag,
			pubKeyFileFlag,
			storeFlag,
			storeHostFlag,
			storePortFlag,
			storeDbnameFlag,
			storeUsernameFlag,
			storePasswordFlag,
		},
		Action: cmdFn,
	}
}

func generateKeys(privKey *ecdsa.PrivateKey) (string, string, string, error) {
	bys := crypto.FromECDSA(privKey)
	// hex keys do not have 0x prefix
	privKeyHex := hex.EncodeToString(bys)

	bys = crypto.FromECDSAPub(&privKey.PublicKey)
	// hex keys do not have 0x prefix
	pubKeyHex := hex.EncodeToString(bys)

	ethAddr := crypto.PubkeyToAddress(privKey.PublicKey)

	return privKeyHex, pubKeyHex, ethAddr.String(), nil
}

func cmdGenerateNewKey() *cli.Command {
	storeFlag := cli.BoolFlag{
		Name:  "store, s",
		Usage: "Save the keys to pub.hex.key / priv.hex.key files",
	}

	cmdFn := func(c *cli.Context) error {
		privKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		priv, pub, eth, err := generateKeys(privKey)
		if err != nil {
			return err
		}

		fmt.Printf("-- ECDSA, SECP256K1 HEX --\n\n")
		fmt.Printf("-- PRIVATE KEY --\n")
		fmt.Printf("%v\n", priv)
		fmt.Printf("-- PRIVATE KEY --\n\n")
		fmt.Printf("-- PUBLIC KEY --\n")
		fmt.Printf("%v\n", pub)
		fmt.Printf("-- PUBLIC KEY --\n\n")
		fmt.Printf("-- ETH --\n")
		fmt.Printf("%v\n", eth)
		fmt.Printf("-- ETH --\n")

		if c.Bool("store") {
			err = ioutil.WriteFile("pub.hex.key", []byte(pub), 0600)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile("priv.hex.key", []byte(priv), 0600)
			if err != nil {
				return err
			}

			fmt.Printf("\n\nWrote key files\n")
		}
		return nil
	}

	return &cli.Command{
		Name:    "generatekey",
		Aliases: []string{"k"},
		Usage:   "Generates a new private and public key to represent an identity",
		Flags: []cli.Flag{
			storeFlag,
		},
		Action: cmdFn,
	}
}

func cmdGenerateGqlCreds() *cli.Command {
	privKeyHexFlag := cli.StringFlag{
		Name:     "privatekey, k",
		Usage:    "Sets the private key to use to sign",
		Required: true,
	}
	didKeyFlag := cli.StringFlag{
		Name:     "did, d",
		Usage:    "Sets the DID to use when generating the key. (Optional)",
		Required: false,
	}

	cmdFn := func(c *cli.Context) error {
		thedid := c.String("did")
		key := c.String("privatekey")

		k, err := crypto.HexToECDSA(key)
		if err != nil {
			return err
		}

		reqTs := ctime.CurrentEpochSecsInInt()
		sig, err := auth.SignEcdsaRequestMessage(k, thedid, reqTs)
		if err != nil {
			return err
		}

		fmt.Printf("did:\n%v\n", thedid)
		fmt.Printf("reqTs:\n%v\n", reqTs)
		fmt.Printf("signature:\n%v\n", sig)

		return nil
	}

	return &cli.Command{
		Name:    "generategqlcreds",
		Aliases: []string{"q"},
		Usage:   "Generates a signature and ts from a given DID for GraphQL access",
		Flags: []cli.Flag{
			privKeyHexFlag,
			didKeyFlag,
		},
		Action: cmdFn,
	}

}
