FROM golang:1.24.4-alpine AS base

FROM base AS dev

WORKDIR /app
VOLUME /private
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download 
CMD ["air"]


FROM base AS builder


WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o adda-backend

FROM scratch AS prod

WORKDIR /prod
VOLUME /private
COPY --from=builder /build/adda-backend ./
EXPOSE 8080
CMD [ "/prod/adda-backend" ]
