# Copyright Jetstack Ltd. See LICENSE for details.
FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install airworthy
ENV AIRWORTHY_VERSION 0.2.0
ENV AIRWORTHY_HASH 2d69cfe0b92f86481805c28d0b8ae47a8ffa6bb2373217e7c5215d61fc9efa1d
RUN curl -sLo /usr/local/bin/airworthy https://github.com/jetstack/airworthy/releases/download/${AIRWORTHY_VERSION}/airworthy_${AIRWORTHY_VERSION}_linux_amd64 && \
  echo "${AIRWORTHY_HASH}  /usr/local/bin/airworthy" | sha256sum -c && \
  chmod 755 /usr/local/bin/airworthy

# install terraform
ENV TERRAFORM_VERSION 0.11.2
RUN airworthy -v download \
  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
  --sha256sums https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS \
  --signature-binary https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig && \
  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  chmod +x terraform && \
  mv terraform /usr/local/bin/terraform

# install terraform plugins
WORKDIR /terraform

ENV TERRAFORM_PROVIDER_AWSTAG_VERSION 0.3.0-rc1
RUN airworthy -v download \
  https://github.com/jetstack/tarmak/releases/download/release-${TERRAFORM_PROVIDER_AWSTAG_VERSION}/terraform-provider-awstag_${TERRAFORM_PROVIDER_AWSTAG_VERSION}_linux_amd64.zip \
  --sha256sums https://github.com/jetstack/tarmak/releases/download/release-${TERRAFORM_PROVIDER_AWSTAG_VERSION}/terraform-provider-awstag_${TERRAFORM_PROVIDER_AWSTAG_VERSION}_SHA256SUMS \
  --signature-binary https://github.com/jetstack/tarmak/releases/download/release-${TERRAFORM_PROVIDER_AWSTAG_VERSION}/terraform-provider-awstag_${TERRAFORM_PROVIDER_AWSTAG_VERSION}_SHA256SUMS.sig && \
  unzip terraform-provider-awstag_linux_amd64.zip && \
  rm terraform-provider-awstag_linux_amd64.zip && \
  chmod +x terraform-provider-awstag && \
  mv terraform-provider-awstag /terraform/terraform-provider-awstag

ADD providers.tf /terraform
RUN terraform init
RUN rm providers.tf
