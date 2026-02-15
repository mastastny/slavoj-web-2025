FROM golang:1.24-alpine AS builder

WORKDIR /workdir

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bin/server .

FROM alpine:latest AS runner

WORKDIR /rundir

RUN apk --no-cache add ca-certificates

COPY --from=builder /workdir/bin/server ./server
COPY --from=builder /workdir/static ./static
COPY --from=builder /workdir/init.sql ./init.sql

EXPOSE 8080

ENTRYPOINT ["./server"]
