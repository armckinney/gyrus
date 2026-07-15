FROM ghcr.io/armckinney/go:1.25.0

ARG WORKSPACE=/workspaces/template-go
ENV WORKSPACE=${WORKSPACE}

# make non-interactive for build
ENV DEBIAN_FRONTEND=noninteractive

RUN mkdir /tmp/container \
    && cd /tmp/container \
    # indirect installs
    && apt update \
    # installs
    && apt install -y \
        postgresql-client \
    # cleanup
    && apt autoremove -y \
    && apt clean -y \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /tmp/container

WORKDIR ${WORKSPACE}

# install go dependencies
COPY go.mod go.sum ./
RUN go mod download

# restore dialog for interactive use
ENV DEBIAN_FRONTEND=dialog
