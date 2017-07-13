
go_codegen:
	mockgen -imports .=github.com/jetstack/tarmak/pkg/tarmak/interfaces -package=mocks -source=pkg/tarmak/interfaces/interfaces.go > pkg/tarmak/mocks/tarmak.go
	mockgen -package=mocks -source=pkg/tarmak/provider/aws/aws.go > pkg/tarmak/mocks/aws.go
