module gitlab.jiangxingai.com/edgenode/dhcp-backend

go 1.13

replace gitlab.jiangxingai.com/applications/base-modules/internal-sdk/go-utils => ./go-utils

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/coredns/coredns v1.6.5
	github.com/coreos/etcd v3.3.17+incompatible // indirect
	github.com/gorilla/mux v1.7.3
	github.com/jinzhu/gorm v1.9.11 // indirect
	github.com/lrh3321/wps v0.0.1 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	gitlab.jiangxingai.com/applications/base-modules/internal-sdk/go-utils v0.0.0-00010101000000-000000000000
	go.etcd.io/etcd v3.3.17+incompatible
	go.mongodb.org/mongo-driver v1.1.3
)
