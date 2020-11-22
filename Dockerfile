FROM golang:1.15.5-buster AS builder
WORKDIR /build
COPY . ./
RUN go build -o smsgatemock

FROM debian:buster
WORKDIR /app
COPY --from=builder /build/smsgatemock /app/smsgatemock
COPY .env /app/.env
RUN apt update && apt install -y ca-certificates
RUN mkdir /app/dbdata
CMD ["./smsgatemock"]
EXPOSE 8811
