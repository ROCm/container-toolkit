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

# Install Go 1.26. The checksum comes from https://go.dev/dl/ and changes
# with the version.
ARG GO_VERSION=1.26.11
ARG GO_SHA256=9849f4e04ee35303714e34fbc3ac8ab44ea3e448a95aa182f91e3b9b79c710ec
RUN wget -q -O /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" && \
    echo "${GO_SHA256}  /tmp/go.tar.gz" | sha256sum -c - && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

#RUN localedef -i en_US -f UTF-8 en_US.UTF-8

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
