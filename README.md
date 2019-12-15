# UpN'Go

This app is currently a work in progress. The ultimate goal is to build a
personal hosted file uploader, so that you can easily share files.

## The `tus` protocol

This app uses [`tus`](https://tus.io) as its protocol, which allows it to
support resumable file uploads.

## Deploying the app

The `docker-compose.yml` file contains everything needed to deploy the app
quickly to any server. You only need to have Docker and `docker-compose`
installed.

These are the only commands you'll need to type:

```sh
# clone the repository
$ git clone https://github.com/alarsyo/upngo
$ cd upngo

# choose a domain name for the app and launch it
$ export UPNGO_HOSTNAME=example.com
$ docker-compose up -d
```

This automatically spawns a Caddy container acting as a reverse proxy to your
app, to serve it over HTTPS with a [Let's Encrypt](https://letsencrypt.org/)
certificate.

This certificate will be stored in `$HOME/.caddy` outside the container, so make
sure you don't have anything conflicting there.

## Testing the setup by uploading a file

The app is still a work in progress, but you can test that your setup is working
by using `tus-py-client`, a Python client library for the `tus` protocol.

First, get the `tus-py-client` on your machine:

```sh
# create a python virtual environment to install the tus client
$ python -m venv venv
$ source venv/bin/activate
$ pip install tuspy
```

Then write this inside a `up.py` file:

```py
import sys

from tusclient import client
from tusclient.storage import filestorage

storage = filestorage.FileStorage('tus_save_file')
my_client = client.TusClient('https://example.com/files/')

uploader = my_client.uploader(sys.argv[1], metadata={"filename": sys.argv[1]}, store_url=True, url_storage=storage)
uploader.upload()
print(uploader.get_url())
```

Finally, run:

```sh
$ python up.py MY_TEST_FILE.txt
```

That's it! Your file was uploaded to the server, and you can download it by
opening the url printed by the script.

## Why UpN'Go

I'm just bad at naming things. That's an **up**loader. Written in **Go**.
