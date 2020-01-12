FROM ubuntu:latest
MAINTAINER Robrotheram <robrotheram@gmail.com>

RUN apt-get update; apt-get install -y ca-certificates imagemagick;  update-ca-certificates
RUN mkdir /app
# Copy the current directory contents into the container at /app
COPY gogallery /app/gogallery
COPY server/config_sample.yml /app/config.yml
RUN chmod +x /app/gogallery
EXPOSE 80
ENV GLLRY_SERVER_PORT ":80"
ENV GLLERY_GALLERY_BASEPATH "/app/pictures"
ENV GLLERY_GALLERY_RENDERERTYPE "imagemagick"
# Set the working directory to /app
WORKDIR /app
ENTRYPOINT ["./gogallery"]
