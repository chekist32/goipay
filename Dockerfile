FROM golang:1.22-alpine AS builder

WORKDIR /app

# Installing system dependencies
RUN apk update && apk add --no-cache make

# Installing golang dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Buidling
COPY . .
RUN make build


FROM alpine AS prod

WORKDIR /app

COPY --from=builder /app/bin/server .
COPY --from=builder /app/config.yml .

EXPOSE 3000

CMD [ "./server" ]