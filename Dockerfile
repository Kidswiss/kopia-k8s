FROM golang:1.17 as build
ENV CGO_ENABLED=0
ENV KOPIA_VERSION=0.12.1

WORKDIR /kopia-k8s
RUN wget "https://github.com/kopia/kopia/releases/download/v${KOPIA_VERSION}/kopia-${KOPIA_VERSION}-linux-x64.tar.gz" && \
    tar xvfz kopia-${KOPIA_VERSION}-linux-x64.tar.gz
ADD ./ .
RUN VERSION=$(git describe --tags --always) && \
    go build -v -a -ldflags "-X main.version=$VERSION"

FROM alpine
ENV KOPIA_VERSION=0.12.1
RUN apk --no-cache add ca-certificates
COPY --from=build /kopia-k8s/kopia-k8s /kopia-k8s
COPY --from=build /kopia-k8s/kopia-${KOPIA_VERSION}-linux-x64 /usr/local/bin
ENTRYPOINT ["/kopia-k8s"]
