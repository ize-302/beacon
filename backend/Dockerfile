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

# the simulator opens this at a path relative to WORKDIR
COPY --from=build /app/cmd/simulator/map_data /app/cmd/simulator/map_data

EXPOSE 8080

CMD ["/app/main"]
