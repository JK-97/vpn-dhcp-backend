$env:GOOS = "linux"
$env:ARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -tags netgo -installsuffix netgo -ldflags "-s"
# go build -v -a -tags netgo -installsuffix netgo -ldflags "-s" ./cmd/keygen
Remove-Item env:ARCH
Remove-Item env:GOOS
Remove-Item env:CGO_ENABLED
