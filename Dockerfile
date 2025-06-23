FROM golang:1.24.4-alpine AS base

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o adda-backend

EXPOSE 8080

CMD [ "/build/adda-backend" ]
