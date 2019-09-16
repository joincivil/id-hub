module github.com/joincivil/id-hub

require (
	github.com/ethereum/go-ethereum v1.8.29-0.20190620093831-25c3282cf126
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/uuid v1.1.0
	github.com/graph-gophers/graphql-go v0.0.0-20190828000412-6bd6fd61c647 // indirect
	github.com/iden3/go-iden3-core v0.0.7-0.20190819135112-6add852d2a3e
	github.com/jinzhu/gorm v1.9.10
	github.com/joincivil/go-common v0.0.0-20190809201706-92cb751fe113
	github.com/karalabe/usb v0.0.0-20190819132248-550797b1cad8 // indirect
	github.com/ockam-network/did v0.1.3
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	github.com/urfave/cli v1.21.0
)

replace github.com/iden3/go-iden3 v0.0.7-0.20190819135112-6add852d2a3e => github.com/iden3/go-iden3-core v0.0.7-0.20190819135112-6add852d2a3e
