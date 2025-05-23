ARG BUILD_BASE_IMAGE=ubuntu:22.04
FROM ${BUILD_BASE_IMAGE}

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Remove old version of go
RUN rm -rf /usr/local/go

RUN apt-get update && apt-get install -y wget protobuf-compiler \
    curl locales ca-certificates build-essential git vim \
    net-tools sudo cmake build-essential && \
    install -m 0755 -d /etc/apt/keyrings && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN localedef -i en_US -f UTF-8 en_US.UTF-8

# Download docker
RUN curl -k -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc && \
    chmod a+r /etc/apt/keyrings/docker.asc

RUN echo 'deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu jammy stable' > /etc/apt/sources.list.d/docker.list

#RUN apt-get update && apt-get install -y \
#  docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin \
#  git && apt-get clean && rm -rf /var/lib/apt/lists/*

# Download go1.23
RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz && \
    rm go1.23.4.linux-amd64.tar.gz

ARG INSECURE_REGISTRY
RUN echo "INSECURE_REGISTRY is: $INSECURE_REGISTRY" && \
    if [ -n "$INSECURE_REGISTRY" ]; then \
    mkdir -p /etc/docker && \
    echo "{ \
    \"insecure-registries\": [\"$INSECURE_REGISTRY\"] \
    }" > /etc/docker/daemon.json; \
    else \
    echo "INSECURE_REGISTRY is not set"; \
    fi

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
