FROM alpine:3.6

RUN apk --update add openssl jq bash unzip curl

ENV VAULT_VERSION 0.6.5
ENV VAULT_HASH c9d414a63e9c4716bc9270d46f0a458f0e9660fd576efb150aede98eec16e23e

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

