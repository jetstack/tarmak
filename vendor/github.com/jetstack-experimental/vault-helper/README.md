# Vault Dev Environment - Setup

Start a development vault server
```
$ vault server -dev
```

Note the root token :
```
Root Token: ########-####-####-####-#############
```

Export root token to environment variable:
```
$ export VAULT_TOKEN=########-####-####-####-############
```

Export vault address (local) :
```
$ export VAULT_ADDR=http://127.0.0.1:8200
```


# Building Golang Vault Helper

Ensure Go path has been set.

Get dep - dependency management tool for go
```
$ go get -u github.com/golang/dep/cmd/dep
```

Use dep ensure
```
$ dep ensure
```

Build golang vault helper (-o vault-helper-golang is required due to name conflict):
```
$ go build -o vault-helper-golang
```
