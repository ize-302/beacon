FROM golang:1.26.6-alpine AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@latest

RUN apk add make

COPY go.mod go.sum ./

RUN go mod download

COPY . .

CMD ["make", "dev"]


FROM golang:1.26.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/main ./cmd/api/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/simulator/main ./cmd/simulator/


FROM alpine:3.22 AS prod

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/main /app/main
COPY --from=build /out/simulator/main /app/simulator/main

# Pulled from a release asset rather than the repo: the file is LFS-tracked, and build
# contexts that clone without fetching LFS objects (Railway) would ship the pointer text
# instead of the PBF. The simulator opens this path relative to WORKDIR.
ADD --chmod=644 \
  https://github.com/ize-302/beacon/releases/download/map-data/lagos.osm.pbf \
  /app/cmd/simulator/map_data/lagos.osm.pbf

EXPOSE 8080

CMD ["/app/main"]
