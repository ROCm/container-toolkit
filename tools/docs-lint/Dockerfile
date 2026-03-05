# Minimal image for running docs lint (markdownlint + pyspelling) locally.
# Use: make docs-lint-docker (from repo root)
ARG BUILD_BASE_IMAGE=ubuntu:22.04
FROM ${BUILD_BASE_IMAGE}

ENV DEBIAN_FRONTEND=noninteractive

# Node.js 22.x (markdownlint-cli2 requires Node >= 20)
RUN apt-get update && apt-get install -y curl ca-certificates && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y \
    make \
    python3 \
    python3-pip \
    aspell \
    aspell-en \
    nodejs \
    git \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN npm install markdownlint-cli2 --global
RUN pip3 install pyspelling

WORKDIR /repo

ENTRYPOINT ["make"]
CMD ["docs-lint"]
