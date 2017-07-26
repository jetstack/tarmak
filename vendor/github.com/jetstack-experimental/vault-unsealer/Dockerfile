FROM alpine:3.6

RUN apk add --update ca-certificates

COPY vault-unsealer_linux_amd64 /usr/local/bin/vault-unsealer

ENTRYPOINT ["/usr/local/bin/vault-unsealer"]
