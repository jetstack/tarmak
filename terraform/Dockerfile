FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install terraform
ENV TERRAFORM_VERSION 0.10.6
ENV TERRAFORM_HASH fbb4c37d91ee34aff5464df509367ab71a90272b7fab0fbd1893b367341d6e23
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
