# Copyright Jetstack Ltd. See LICENSE for details.
FROM alpine:3.8

RUN apk --update add openssl jq bash unzip curl

ENV VAULT_VERSION 0.9.6
ENV VAULT_HASH 3f1f346ff7aaf367fed6a3e83e5a07fdc032f22860585e36c3674f9ead08dbaf

RUN curl -sL  https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip > /tmp/vault.zip && \
    echo "${VAULT_HASH}  /tmp/vault.zip" | sha256sum  -c && \
    unzip /tmp/vault.zip && \
    rm /tmp/vault.zip && \
    mv vault /usr/local/bin/vault && \
    chmod +x /usr/local/bin/vault

ADD vault-helper_linux_amd64 /usr/local/bin/vault-helper

ENV VAULT_ADDR=http://127.0.0.1:8200

EXPOSE 8200

ENTRYPOINT ["/usr/local/bin/vault-helper"]

CMD []
