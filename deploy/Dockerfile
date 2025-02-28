FROM golang:1.23.3-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG ENV_PREFIX

RUN CGO_ENABLED=0 go build -o /cmd/app/honeybee /app/cmd/app/main.go

CMD ["/cmd/app/honeybee"]