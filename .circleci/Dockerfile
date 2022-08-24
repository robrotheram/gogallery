FROM golang:1.19
RUN apt-get update && apt-get install -y git openssh-server tar gzip ca-certificates make curl gnupg docker.io libvips libvips-dev
RUN curl -sL https://deb.nodesource.com/setup_16.x  | bash -
RUN apt-get -y install nodejs
#RUN go get -u github.com/gobuffalo/packr/v2/packr2
