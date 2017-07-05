FROM debian:jessie-slim
 
MAINTAINER Development Team <dev@facette.io>
 
ENV GO_VERSION=1.8.3 NODE_VERSION=7.10.0 PREFIX=/usr BUILD_TAGS=builtin_assets
 
RUN echo "deb http://deb.debian.org/debian jessie-backports main" >>/etc/apt/sources.list && \
    apt-get update && \
    apt-get install --no-install-recommends -y -t jessie-backports \
        build-essential \
        ca-certificates \
        curl \
        git \
        librrd-dev \
        pandoc \
        xz-utils && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
 
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
 
RUN curl -s -L https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | \
    tar -C /usr/local -xvzf -
 
ENV PATH=${PATH}:/usr/local/go/bin
 
RUN curl -s -L https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz | \
    tar -C /usr/local --transform "flags=r;s/^node-v${NODE_VERSION}-linux-x64/node/" -xvJf -
 
ENV PATH=${PATH}:/usr/local/node/bin
 
COPY . /facette

WORKDIR /facette 

RUN make && \
    make install && \
    install -D docs/examples/facette.yaml /etc/facette/facette.yaml && \
    useradd -r -m -u 12003 -s /usr/sbin/nologin -d /var/lib/facette facette && \
    sed -i -r \
        -e 's/listen: localhost:12003/listen: :12003/' \
        -e 's/path: data.db/path: \/var\/lib\/facette\/data.db/' \
        -e 's/assets_dir: assets/assets_dir: \/usr\/share\/facette\/assets/' \
        /etc/facette/facette.yaml && \
    rm -rf /facette

WORKDIR /

VOLUME /var/lib/facette

EXPOSE 12003

USER 12003

ENTRYPOINT ["facette"]
 
# vim: ts=4 sw=4 et
