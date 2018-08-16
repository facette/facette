FROM alpine:latest

LABEL maintainer="Development Team <dev@facette.io>"

COPY . /root/go/src/facette.io/facette

RUN apk --no-cache add --virtual .build-deps git go make musl-dev nodejs rrdtool-dev yarn && \
    GOBIN=/usr/local/bin go get github.com/jteeuwen/go-bindata/... && \
    make -C /root/go/src/facette.io/facette build install && \
    install -D /root/go/src/facette.io/facette/docs/examples/facette.yaml /etc/facette/facette.yaml && \
    sed -i -r \
        -e "s|listen: localhost:12003|listen: :12003|" \
        -e "s|path: var/data.db|path: /var/lib/facette/data.db|" \
        -e "s|path: var/cache|path: /var/cache/facette|" \
        /etc/facette/facette.yaml && \
    rm -rf /root/go && \
    apk del .build-deps

RUN apk --no-cache add ca-certificates rrdtool

RUN adduser -h /var/lib/facette -S -D -u 1234 facette

USER 1234

EXPOSE 12003

VOLUME /var/lib/facette

ENTRYPOINT ["facette"]

# vim: ts=4 sw=4 et
