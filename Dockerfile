FROM alpine:3.4

RUN apk --update add openssl jq bash unzip curl

ENV VAULT_VERSION 0.6.4
ENV VAULT_HASH 04d87dd553aed59f3fe316222217a8d8777f40115a115dac4d88fac1611c51a6

RUN curl -sL  https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip > /tmp/vault.zip && \
    echo "${VAULT_HASH}  /tmp/vault.zip" | sha256sum  -c && \
    unzip /tmp/vault.zip && \
    rm /tmp/vault.zip && \
    mv vault /usr/local/bin/vault && \
    chmod +x /usr/local/bin/vault

ADD vault-helper /usr/local/bin/vault-helper
ADD vault-setup /usr/local/bin/vault-setup

ENV VAULT_ADDR=http://127.0.0.1:8200

EXPOSE 8200

ENTRYPOINT ["/usr/local/bin/vault-helper"]

CMD ["dev-server"]

