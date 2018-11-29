FROM golang:alpine AS build

RUN apk add --no-cache \
      git \
      curl

RUN mkdir -p /out/usr/local/bin

ENV KUBELETMEIN $GOPATH/src/github.com/4armed/kubeletmein
RUN mkdir -p "$(dirname ${KUBELETMEIN})"
COPY . $KUBELETMEIN

WORKDIR $KUBELETMEIN
RUN go build ./cmd/kubeletmein && cp ./kubeletmein /out/usr/local/bin/kubeletmein

WORKDIR /out
RUN curl -sL https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl -o usr/local/bin/kubectl && \
    chmod +x usr/local/bin/kubectl

FROM google/cloud-sdk:alpine
LABEL maintainer="Marc Wickenden <marc@4armed.com>"

RUN addgroup -S app && adduser -S app -G app

COPY --from=build /out /
USER app
WORKDIR /home/app

CMD [ "kubeletmein" ]
