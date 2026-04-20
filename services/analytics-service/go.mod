module github.com/andev0x/analytics-service

go 1.21.0

require (
	github.com/andev0x/event-driven-order-system/pkg v0.0.0
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/gorilla/mux v1.8.1
	github.com/prometheus/client_golang v1.18.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.18.0
)

require (
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.6
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/andev0x/event-driven-order-system/pkg => ../../pkg
