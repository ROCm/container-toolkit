ARG FROM=almalinux
ARG VERSION=latest

FROM ${FROM}:${VERSION}

RUN yum install -y gcc gcc-c++ \
    libtool libtool-ltdl \
    make cmake \
    git \
    pkgconfig \
    sudo \
    automake autoconf \
    wget \
    yum-utils rpm-build && \
    yum clean all

# Remove old version of go
RUN rm -rf /usr/local/go

# Install Go 1.23
RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz 
RUN tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
RUN rm go1.23.4.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

#RUN localedef -i en_US -f UTF-8 en_US.UTF-8

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
