FROM golang:1.13-alpine as builder
COPY . /build
WORKDIR /build
RUN ls -la .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -mod=vendor -o=server cmd/server/main.go

FROM scratch
COPY --from=builder /build/server /opt/
WORKDIR /opt/
CMD ["./server"]
