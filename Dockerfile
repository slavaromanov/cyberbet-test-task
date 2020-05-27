FROM golang:alpine

WORKDIR /app

COPY . .
RUN go build

CMD ./cyberbet-test-task --dump /app/data/storage.gob