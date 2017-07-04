FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install terraform
ENV TERRAFORM_VERSION 0.9.10
ENV TERRAFORM_HASH 77f0d01182d665f7f3c63c326aa699b452fba043c2e2f9050c4bd114f98a1207
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform
