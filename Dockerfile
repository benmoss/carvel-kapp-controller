# --- deps image ---
FROM --platform=$BUILDPLATFORM golang:1.17.9 AS deps
WORKDIR /workspace
COPY hack/deps.* go.* .
COPY vendor vendor
ARG TARGETOS TARGETARCH
RUN go run -mod=vendor ./deps.go -dest=/usr/local/bin -os=${TARGETOS} -arch=${TARGETARCH}

# --- build image ---
FROM --platform=$BUILDPLATFORM golang:1.17.9 AS build
ARG KCTRL_VER=development
ARG TARGETARCH=amd64
ARG TARGETOS=linux
WORKDIR /workspace
COPY . .
# helpful ldflags reference: https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=vendor -ldflags="-X 'main.Version=$KCTRL_VER' -buildid=" -trimpath -o kapp-controller ./cmd/main.go

# --- run image ---
FROM photon:4.0 AS run

# Install openssh for git
# TODO(bmo): why do we need sed?
RUN tdnf install -y git openssh-clients sed

# Create the kapp-controller user in the root group, the home directory will be mounted as a volume
RUN echo "kapp-controller:x:1000:0:/home/kapp-controller:/usr/sbin/nologin" > /etc/passwd
# Give the root group write access to the openssh's root bundle directory
# so we can rename the certs file with our dynamic config, and append custom roots at runtime
RUN chmod g+w /etc/pki/tls/certs

COPY --from=deps /usr/local/bin/* .

COPY --from=build /workspace/kapp-controller .

# Copy the ca-bundle so we have an original
RUN cp /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt.orig

# Run as kapp-controller by default, will be overridden to a random uid on OpenShift
USER 1000
ENV PATH="/:${PATH}"
ENTRYPOINT ["/kapp-controller"]
