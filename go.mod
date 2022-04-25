module github.com/forbole/bdjuno

go 1.15

require (
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/cosmos/cosmos-sdk v0.42.9
	github.com/desmos-labs/juno v0.0.0-20210923065451-3b09c4e72f19
	github.com/go-co-op/gocron v0.3.3
	github.com/gogo/protobuf v1.3.3
	github.com/jmoiron/sqlx v1.2.1-0.20200324155115-ee514944af4b
	github.com/lib/pq v1.9.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/onomyprotocol/onomy v0.0.4
	github.com/pelletier/go-toml v1.9.5
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/rs/zerolog v1.23.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.14
	github.com/ziutek/mymysql v1.5.4 // indirect
	google.golang.org/grpc v1.43.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/tendermint/tendermint => github.com/forbole/tendermint v0.34.13-0.20210820072129-a2a4af55563d

replace github.com/cosmos/cosmos-sdk => github.com/onomyprotocol/onomy-sdk v0.42.10-0.20211228140704-1a3046991600
