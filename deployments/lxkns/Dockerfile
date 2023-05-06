# Requires Docker buildx in order to correctly produce the final image with the
# extended file attributes for capabilities still intact.

ARG ALPINE_VERSION=3.17
ARG ALPINE_PATCH=2
ARG GO_VERSION=1.20.4
ARG NODE_VERSION=16

# 1st stage: build the lxkns binary, now this requires cgo and we thus need gcc
# ... and then we also need header files. Oh, well. Caching to the rescue; we
# start with the gcc and header stuff, which is kind of base builder image stuff
# anyway.
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ARG WSDISKPATHS
ARG MOD1=./
ARG MOD2=./
ARG MOD3=./
ARG MOD4=./
ARG MOD5=./
ARG MOD6=./
ARG MOD7=./
ARG MOD8=./
ARG MOD9=./

# General warm-up
RUN apk add --no-cache --update-cache gcc musl-dev libcap
RUN go build std

WORKDIR /ws
# Copy the additionally used modules into the soon-to-be workspace.
COPY --from=bctx1 . ${MOD1}
COPY --from=bctx2 . ${MOD2}
COPY --from=bctx3 . ${MOD3}
COPY --from=bctx4 . ${MOD4}
COPY --from=bctx5 . ${MOD5}
COPY --from=bctx6 . ${MOD6}
COPY --from=bctx7 . ${MOD7}
COPY --from=bctx8 . ${MOD8}
COPY --from=bctx9 . ${MOD9}

# Make sure we have the main module containing a main package to be build...
COPY go.mod go.sum ./lxkns/

# Establish the Go workspace
RUN go work init ${WSDISKPATHS}

WORKDIR /ws/lxkns
# We now try to cache only the dependencies in a separate layer, so we can speed
# up things in case the dependencies do not change. This then reduces the amount
# of fetching and compiling required when compiling the final binary later.
RUN go mod download -x
# And now, finally, we build the lxkns service itself.
COPY api/ ./api/
COPY cmd/ ./cmd/
COPY containerizer/ ./containerizer/
COPY decorator/ ./decorator/
COPY discover/ ./discover/
COPY internal/ ./internal/
COPY log/ ./log/
COPY model/ ./model/
COPY mounts/ ./mounts/
COPY nsioctl/ ./nsioctl/
COPY ops/ ./ops/
COPY plural/ ./plural/
COPY species/ ./species/
COPY *.go ./

RUN go build -v -tags osusergo,netgo -ldflags "-s -w" -o /lxkns ./cmd/lxkns
RUN setcap "cap_sys_admin,cap_sys_chroot,cap_sys_ptrace,cap_dac_read_search,cap_dac_override+ep" /lxkns

RUN go build -v -tags osusergo,netgo -ldflags "-s -w" -o /mntnssandbox ./cmd/mntnssandbox
RUN setcap "cap_sys_admin,cap_sys_chroot,cap_sys_ptrace,cap_dac_read_search,cap_dac_override+ep" /mntnssandbox

# 2nd stage: builds the lxkns web client react application.
FROM node:${NODE_VERSION}-alpine AS reactor
WORKDIR /webapp
ENV PATH /webapp/node_modules/.bin:$PATH
# Cache the dependency hell, so we don't need to recreate it most of the time
# when dependencies don't change.
COPY web/lxkns/package.json \
     web/lxkns/craco.config.js \
     web/lxkns/mdxplugins.js \
     web/lxkns/yarn.lock \
     web/lxkns/tsconfig.json \
        ./
# While not being a true production install in the original sense, this avoids
# installing cypress and the react styleguidist which we all don't need in
# creating the production build.
RUN yarn install --production --network-timeout 1000000000
# Now build the production-optimized web app.
COPY web/lxkns/public/ ./public/
COPY web/lxkns/src/ ./src/
ARG GIT_VERSION
RUN yarn imagebuild

# 3rd and final stage: create the final image containing only the lxkns binary
# and its required shared libraries, as well as the lxkns web app.
FROM alpine:${ALPINE_VERSION}.${ALPINE_PATCH}
COPY --from=builder /lxkns /mntnssandbox /
COPY --from=reactor /webapp/build/ /web/lxkns/build/
ENV PATH /
USER 65534
CMD ["/lxkns"]