FROM golang:1.13-alpine as builder
COPY . /build/
WORKDIR /build
RUN ls -la .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -mod=vendor -o=emulator main.go

FROM scratch
COPY --from=builder /build/emulator /opt/
COPY routes /opt/routes
WORKDIR /opt/
CMD ["./emulator"]
