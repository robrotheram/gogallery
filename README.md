# GoGallery

A very simple gallery server written in go use your files as the directory structure. No upload /database (well we have a internal K/V for a cache)

Using the golang templating so no massive ball of javascript to edit

Supports PWA (Progressive Web Apps) the is very simple service worker.

Demo at https://gallery.exceptionerror.io 

## Installation

Clone the repo and run dep ensure to get all the dependencies

or use the docker image
```
docker build -t gogallery .
docker run -p 6060:80 -v $(pwd):/app/pictures gogallery
```

## Usage

Edit the config and change the name basepath and base folder that is used for scanning images

#### Admin
The application contains a basic admin section accessed by '/admin' using the username/password combo
The password use the bcrypt algorithum and therefore before using it you needed to created the hash and store in the config.
To generate the password do  './gogallery generate password' 

#### Configuration 
Config can be also edited via environmental variables
 

```
GLLRY_SERVER_PORT
GLLRY_SERVER_WORKERS

GLLRY_DATABASE_BASEURL

GLLRY_GALLERY_NAME
GLLRY_GALLERY_BASEPATH
GLLRY_GALLERY_URL
GALLRY_GALLERY_THEME

GLLRY_ABOUT_ENABLE
GLLRY_ABOUT_INSTAGRAM
GLLRY_ABOUT_TWITTER
GLLRY_ABOUT_FACEBOOK
GLLRY_ABOUT_EMAIL
GLLRY_ABOUT_BLOG
GLLRY_ABOUT_WEBSITE
GLLRY_ABOUT_DESCRIPTION
GLLRY_ABOUT_PHOTOGRAPHER
GLLRY_ABOUT_BACKGROUNDPHOTO
GLLRY_ABOUT_PROFILEPHOTO
GLLRY_ABOUT_FOOTER

GLLRY_ADMIN_ENABLE
GLLRY_ADMIN_USERNAME
GLLRY_ADMIN_PASSWORD

```



## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)
