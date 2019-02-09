FROM alpine:latest

MAINTAINER Robrotheram <robrotheram@gmail.com>



# Copy the current directory contents into the container at /app
COPY . /app
RUN chmod +x /app/gogallery
EXPOSE 80
ENV GLLRY_SERVER_PORT ":80"
ENV GLLERY_GALLERY_BASEPATH "/app/pictures"
# Set the working directory to /app
WORKDIR /app

ENTRYPOINT ["./awesomeProject"]
