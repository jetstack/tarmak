FROM ruby:2.3

# apt-get install
RUN apt-get update && apt-get install -y unzip jq && rm -rf /var/lib/apt/lists/*

# install cloudflare ssl
ENV CFSSL_VERSION 1.2
ENV CFSSL_HASH eb34ab2179e0b67c29fd55f52422a94fe751527b06a403a79325fed7cf0145bd
ENV CFSSLJSON_HASH 1c9e628c3b86c3f2f8af56415d474c9ed4c8f9246630bd21c3418dbe5bf6401e
RUN curl -s -L -o /usr/local/bin/cfssl     https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssl_linux-amd64 && \
    curl -s -L -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssljson_linux-amd64 && \
    chmod +x /usr/local/bin/cfssl /usr/local/bin/cfssljson && \
    echo "${CFSSL_HASH}  /usr/local/bin/cfssl" | sha256sum -c && \
    echo "${CFSSLJSON_HASH}  /usr/local/bin/cfssljson" | sha256sum -c

# install packer
ENV PACKER_VERSION 1.0.0
ENV PACKER_HASH ed697ace39f8bb7bf6ccd78e21b2075f53c0f23cdfb5276c380a053a7b906853
RUN curl -sL  https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_amd64.zip > /tmp/packer.zip && \
    echo "${PACKER_HASH}  /tmp/packer.zip" | sha256sum  -c && \
    unzip /tmp/packer.zip && \
    rm /tmp/packer.zip && \
    mv packer /usr/local/bin/packer && \
    chmod +x /usr/local/bin/packer

# install terraform
ENV TERRAFORM_VERSION 0.9.3
ENV TERRAFORM_HASH f34b96f7b7edaf8c4dc65f6164ba0b8f21195f5cbe5b7288ad994aa9794bb607
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform

# install vault
ENV VAULT_VERSION 0.6.5
ENV VAULT_HASH c9d414a63e9c4716bc9270d46f0a458f0e9660fd576efb150aede98eec16e23e
RUN curl -sL  https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip > /tmp/vault.zip && \
    echo "${VAULT_HASH}  /tmp/vault.zip" | sha256sum  -c && \
    unzip /tmp/vault.zip && \
    rm /tmp/vault.zip && \
    mv vault /usr/local/bin/vault && \
    chmod +x /usr/local/bin/vault

# install kubectl
ENV KUBECTL_VERSION 1.5.3
ENV KUBECTL_HASH 9cfc6cfb959d934cc8080c2dea1e5a6490fd29e592718c5b2b2cfda5f92e787e
RUN curl -sL https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl && \
    echo "${KUBECTL_HASH}  /usr/local/bin/kubectl" | sha256sum  -c && \
    chmod +x /usr/local/bin/kubectl

# install rubygems
WORKDIR /work
ADD Gemfile .
ADD Gemfile.lock .
RUN bundle install --path vendor/
