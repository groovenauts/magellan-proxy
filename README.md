# magellan-proxy - MAGELLAN HTTP request adaptor

## How to use magellan-proxy in your Docker image

### Add magellan-proxy in your image

(In your Dockerfile)
```
ADD https://github.com/groovenauts/magellan-proxy/releases/download/v0.0.1/magellan-proxy_linux-amd64 /usr/app/magellan-proxy
RUN chmod +x /usr/app/magellan-proxy
```

### Change CMD of your container

Prepend "magellan-proxy" before your application's commandline.

ex1)
```
CMD ["bundle", "exec", "puma", "--port", "80"]
↓
CMD ["/usr/app/magellan-proxy", "bundle", "exec", "puma", "--port", "80"]
```

ex2)
```
CMD bundle exec rails server production --port 80
↓
CMD /usr/app/magellan-proxy bundle exec rails server production --port 80
```

You can specify port and concurrency number.

```
CMD ["/usr/app/magellan-proxy", "--port", "8080", "--num", "5", "bundle", "exec", "puma", "-t", "5:5", "--port", "8080"]
```

## Installation

```
go get github.com/groovenauts/magellan-proxy
```

## How to biuld

At first, install gom.

```
go get github.com/mattn/gom
```

### Build for current platform

```
gom build
```

### Build for Docker image (Linux/amd64)

```
./scripts/build.sh
```

## License

MIT
See LICENSE.txt.
