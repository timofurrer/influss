# influss

Minimalist self-hosted read-it-later server that just produces an RSS feed.

## Host influss

influss can easily be self-hosted. It's a single Go binary and currently
supports storing the clipped websites in a file system directory.

```shell
docker run -d --name influss -p 8080:8080 -v $(pwd)/influss-store:/store ghcr.io/timofurrer/influss/influss
```

Running influss with `docker compose` makes configuration a little bit easier:

```yaml
services:
  influss:
    image: ghcr.io/timofurrer/influss/influss:1.0.28
    ports:
      - "8080:8080"
    volumes:
      - store:/var/lib/influss/store
    command:
      - '--listen-addr=:8080'
      - '--use-local-store=true'
      - '--local-store-dir=/var/lib/influss/store'
      - '--feed-title=Read it later (Influss)'
      - '--feed-author-name=<your name>'
      - '--feed-author-email=<your email>'
      - '--feed-link=<external URL>/clips'
      - '--feed-description=Read it later RSS feed produced by Influss'

volumes:
  store:
```

influss does not come with any means of authentication.
Therefore, we recommend running it behind a reverse proxy that
also takes care of authentication. We recommend [caddy](https://caddyserver.com/).

Configure simple HTTP basic auth with your reverse proxy for influss:

```caddyfile
influss.<your domain> {
        basicauth /* {
                <username> <caddy hashed password>
        }
        reverse_proxy localhost:8080
}
```

Use [`caddy hash-password`](https://caddyserver.com/docs/command-line#caddy-hash-password)
to hash the password for the `Caddyfile` configuration.

## Configure RSS reader

Configure your RSS reader to point to `influss.<your domain>/clips` and optionally
use the HTTP basic auth credentials.

We recommend [miniflux](https://miniflux.app/) as the RSS reader.

## Install browser extension

influss currently only provides a Firefox Add-on.
It's still in review by Mozilla as of writing this.

However, you may already install the add-on from the [`extensions`](./extensions)
directory.

### Configure

To configure the browser extension go to its settings and configure
the endpoint, username and password.
