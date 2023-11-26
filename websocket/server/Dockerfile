##
## Build
##
FROM golang:1.18-bullseye AS build


WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o ws-server .
WORKDIR /dist
RUN cp /build/ws-server .

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /
COPY --from=build /dist/ws-server /
EXPOSE 8080
ENTRYPOINT ["/ws-server"]