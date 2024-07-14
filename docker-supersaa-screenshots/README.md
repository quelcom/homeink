# Docker-supersaa-screenshots

## Build and run

```
docker build -t supersaa . && docker run --rm -v /home/quelcom/docker/supersaa/out:/tmp/screenshot supersaa
```

## Tests

```
./run-tests.sh
```

Note: The script will spawn a `chromedp/headless-shell` container and map the host port 9222 to the same port in the container. The Go tests will also run a test HTTP server in port 8000 to serve a static HTML copy of the Supersää website, which was generated with [monolith](https://github.com/Y2Z/monolith) using this command:

```
monolith -j https://www.is.fi/supersaa/suomi/oulu/643492/ -B -d .googletagmanager.com -o supersaa_oulu_$(date +"%Y-%m-%d_%H%M").html
```

