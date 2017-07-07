FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install terraform
ENV TERRAFORM_VERSION 0.9.11
ENV TERRAFORM_HASH 804d31cfa5fee5c2b1bff7816b64f0e26b1d766ac347c67091adccc2626e16f3
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform
