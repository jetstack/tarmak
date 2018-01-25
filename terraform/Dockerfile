# Copyright Jetstack Ltd. See LICENSE for details.
FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install terraform
ENV TERRAFORM_VERSION 0.11.2
ENV TERRAFORM_HASH f728fa73ff2a4c4235a28de4019802531758c7c090b6ca4c024d48063ab8537b
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform

# install terraform plugins
WORKDIR /terraform
ADD providers.tf /terraform
RUN terraform init
RUN rm providers.tf
