# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM golang:1.18 as builder

EXPOSE 3000

RUN apt-get update \
    && apt-get install -y -q build-essential sqlite3 libsqlite3-dev postgresql libpq-dev vim

# Installing Node 12
RUN curl -sL https://deb.nodesource.com/setup_18.x | bash
RUN apt-get update && apt-get install nodejs

# Installing Postgres
RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list' \
    && wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | apt-key add - \
    && apt-get update \
    && apt-get install -y -q postgresql postgresql-contrib libpq-dev\
    && rm -rf /var/lib/apt/lists/* \
    && service postgresql start && \
    # Setting up password for postgres
    su -c "psql -c \"ALTER USER postgres  WITH PASSWORD 'postgres';\"" - postgres

# Installing yarn
RUN npm install -g --no-progress yarn \
    && yarn config set yarn-offline-mirror /npm-packages-offline-cache \
    && yarn config set yarn-offline-mirror-pruning true

# Install golangci
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0

# Installing buffalo binary
RUN wget https://github.com/gobuffalo/cli/releases/download/v0.18.14/buffalo_0.18.14_Linux_x86_64.tar.gz
RUN tar -xvzf buffalo_0.18.14_Linux_x86_64.tar.gz
RUN mv buffalo /usr/local/bin/buffalo

WORKDIR /
RUN go install github.com/gobuffalo/buffalo-pop/v3@latest

RUN mkdir /src
WORKDIR /src

ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/vidlink
WORKDIR /src/vidlink

# this will cache the npm install step, unless package.json changes
ADD package.json .
ADD yarn.lock .yarnrc.yml ./
RUN mkdir .yarn
COPY .yarn .yarn
RUN yarn install
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ADD . .
RUN buffalo build --static -o /bin/app

FROM alpine
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates
RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

WORKDIR /bin/

COPY --from=builder /bin/app .

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/app
