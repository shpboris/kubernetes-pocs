# syntax=docker/dockerfile:experimental
# ---
FROM golang:1.19 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /work
COPY . /work

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/admission-webhook .

# ---
FROM scratch AS run

COPY --from=build /work/bin/admission-webhook /usr/local/bin/

CMD ["admission-webhook"]
