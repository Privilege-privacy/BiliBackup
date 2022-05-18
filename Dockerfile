FROM golang:1.18.2-alpine3.15 as builder
WORKDIR /test
COPY . .
RUN go build -o main main.go

FROM alpine:3.15.4
WORKDIR /root
COPY --from=builder /test/main .
COPY init.sh rclone.conf /root/
RUN apk add --update --no-cache rclone ffmpeg && \
mkdir -p /root/.config/rclone && \
mv rclone.conf /root/.config/rclone/rclone.conf
ENTRYPOINT [ "./init.sh" ]
