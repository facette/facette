FROM alpine:latest

LABEL maintainer="Development Team <dev@facette.io>"

COPY . /tmp/build

RUN apk --no-cache add --virtual .build-deps git go make musl-dev nodejs rrdtool-dev && \
    make -C /tmp/build build install && \
    install -D /tmp/build/docs/examples/facette.yaml /etc/facette/facette.yaml && \
    sed -i -r \
        -e 's/listen: localhost:12003/listen: :12003/' \
        /etc/facette/facette.yaml && \
    rm -rf /tmp/build && \
    apk del .build-deps

RUN apk --no-cache add rrdtool

RUN adduser -h /var/lib/facette -S -D -u 1234 facette

USER 1234

EXPOSE 12003

VOLUME /var/lib/facette

ENTRYPOINT ["facette"]

# vim: ts=4 sw=4 et
