FROM golang:1.19.5-alpine3.16 as builder
WORKDIR /build
COPY . .
RUN apk add --update --no-cache gcc && go build -o BiliBackup main.go

FROM alpine:3.15.4
WORKDIR /root
COPY --from=builder /build/BiliBackup .
COPY init.sh rclone.conf /root/
RUN apk add --update --no-cache rclone ffmpeg && \
mkdir -p /root/.config/rclone && \
mv rclone.conf /root/.config/rclone/rclone.conf
ENTRYPOINT [ "./init.sh" ]
