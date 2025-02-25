FROM golang:1.23.4-alpine AS builder

# the default user in golang image is root
WORKDIR /root/Chamael

# copy the chamael source code to the image
COPY . .

# enable cgo
ENV CGO_ENABLED=1

# add the aliyun mirror
RUN echo "https://mirrors.aliyun.com/alpine/v3.21/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.21/community" >> /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev bash

# add the goproxy   
ENV GOPROXY=https://goproxy.cn,direct

# download the dependencies
RUN go mod download

# make the start_all.sh executable
RUN chmod +x start_all.sh && chmod +x start_one.sh

ENTRYPOINT ["./start_all.sh"]