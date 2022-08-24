FROM ubuntu:20.04
RUN apt-get update; apt-get install -y libvips && rm -rf /var/lib/apt/lists/*
RUN mkdir /app
# Copy the current directory contents into the container at /app
COPY gogallery /app/gogallery
COPY config_sample.yml /app/config.yml
RUN chmod +x /app/gogallery
EXPOSE 80
ENV GLLRY_SERVER_PORT ":80"
ENV GLLRY_GALLERY_BASEPATH "/app/pictures"
ENV GLLRY_GALLERY_THEME "default"
WORKDIR /app
ENTRYPOINT ["./gogallery", "--config", "/app/config.yml",  "serve"]
