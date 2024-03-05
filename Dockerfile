# Build stage
FROM golang:1.22.0-alpine as builder
WORKDIR /app
COPY . .
RUN apk add curl
RUN apk add make

# Download sqlc and unzip
RUN curl -L https://downloads.sqlc.dev/sqlc_1.25.0_linux_amd64.tar.gz --output sqlc_1.25.0_linux_amd64.tar.gz
RUN tar xvzf sqlc_1.25.0_linux_amd64.tar.gz
RUN mv ./sqlc /usr/bin

#Download migrate and unzip
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz --output migrate.linux-amd64.tar.gz
RUN tar xvzf migrate.linux-amd64.tar.gz migrate

# Generate the model files.
RUN make sqlc

RUN go build -o main main.go

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration


EXPOSE 8080
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]