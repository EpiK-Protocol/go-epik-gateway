module github.com/EpiK-Protocol/go-epik-data

go 1.16

require (
	github.com/EpiK-Protocol/go-epik v1.0.1-0.20211007091417-a53d087e8df2
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef
	github.com/bramvdbogaerde/go-scp v1.1.0
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-bitfield v0.2.4
	github.com/filecoin-project/go-data-transfer v1.5.0
	github.com/filecoin-project/go-fil-markets v1.3.0
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-multistore v0.0.3
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.7
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/libp2p/go-libp2p-metrics v0.1.0
	github.com/libp2p/go-libp2p-protocol v0.1.0
	github.com/libp2p/go-libp2p-pubsub v0.4.2-0.20210212194758-6c1addf493eb
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cast v1.3.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vesoft-inc/nebula-go/v2 v2.5.1
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/filecoin-project/specs-actors/v2 => github.com/EpiK-Protocol/go-epik-actors/v2 v2.4.0-alpha.0.20211007091141-d6b2892aaedc
