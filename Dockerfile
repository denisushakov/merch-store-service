FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CONFIG_PATH=configs/values_local.yaml

RUN go build -o /build ./cmd/merch-store

EXPOSE 8080

CMD ["/build"]
