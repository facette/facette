FROM debian:jessie

RUN apt-get update && apt-get install --no-install-recommends -y \
    ca-certificates \
    curl \
    librrd-dev \
    pandoc \
    npm \
    nodejs \
    build-essential \
    git-core && \
    ln -s /usr/bin/nodejs /usr/bin/node

RUN curl -s https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH $PATH:/usr/local/go/bin:/go/bin

COPY . /facette
WORKDIR /facette
 
RUN make && make install && \
    mkdir -p /etc/facette && \
    cp docs/examples/facette.json /etc/facette/facette.json

RUN mkdir -p /usr/share/facette && \
    mkdir -p /var/lib/facette && \
    mkdir -p /etc/facette/providers && \
    mkdir -p /var/run/facette && \
    chown -R 1:1 /usr/share/facette /var/lib/facette /etc/facette /var/run/facette

USER 1
EXPOSE 12003
ENTRYPOINT ["facette"]
