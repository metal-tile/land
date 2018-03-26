gofmt -w ./*.go
gofmt -w ./dqn/*.go
gofmt -w ./firedb/*.go

golint ./*.go
golint ./dqn/*.go
golint ./firedb/*.go

go vet ./*.go
go vet ./dqn/*.go
go vet ./firedb/*.go
