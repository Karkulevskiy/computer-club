FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .

COPY . .

RUN go build -o task.exe ./src/cmd/main.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/task.exe /build/task.exe
COPY --from=builder /build/test_file.txt /build/test_file.txt

ENTRYPOINT ["/build/task.exe"]