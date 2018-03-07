FROM alpine:3.6

RUN apk --update add openssl jq bash unzip curl

ENV VAULT_VERSION 0.9.5
ENV VAULT_HASH f6dbc9fdac00598d2a319c9b744b85bf17d9530298f93d29ef2065bc751df099

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
