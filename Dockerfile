FROM ruby:2.3

# apt-get install
RUN apt-get update && apt-get install -y unzip && rm -rf /var/lib/apt/lists/*

# install kubectl
ENV KUBECTL_VERSION 1.5.3
ENV KUBECTL_HASH 9cfc6cfb959d934cc8080c2dea1e5a6490fd29e592718c5b2b2cfda5f92e787e
RUN curl -sL https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl && \
    echo "${KUBECTL_HASH}  /usr/local/bin/kubectl" | sha256sum  -c && \
    chmod +x /usr/local/bin/kubectl

# install terraform
ENV TERRAFORM_VERSION 0.8.8
ENV TERRAFORM_HASH 403d65b8a728b8dffcdd829262b57949bce9748b91f2e82dfd6d61692236b376
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform

# install packer
ENV PACKER_VERSION 0.12.3
ENV PACKER_HASH d11c7ff78f546abaced4fcc7828f59ba1346e88276326d234b7afed32c9578fe
RUN curl -sL  https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_amd64.zip > /tmp/packer.zip && \
    echo "${PACKER_HASH}  /tmp/packer.zip" | sha256sum  -c && \
    unzip /tmp/packer.zip && \
    rm /tmp/packer.zip && \
    mv packer /usr/local/bin/packer && \
    chmod +x /usr/local/bin/packer

# install rubygems
WORKDIR /work
ADD Gemfile .
ADD Gemfile.lock .
RUN bundle install --path vendor/

# add terraform
ADD . .
