FROM golang:1.18.4 as builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o waffel ./main/main.go

FROM alpine:latest
COPY --from=builder /src/waffel /
CMD /waffel