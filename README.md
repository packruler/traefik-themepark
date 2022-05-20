# Rewrite Body

This is a fork of [Traefik](https://github.com/traefik)'s [plugin-rewritebody](https://github.com/traefik/plugin-rewritebody)
that is aimed at extending support to handle `gzip` content. This was initially aimed at extending the support for utilizing
[theme.park](https://github.com/traefik/plugin-rewritebody)'s themes but can likely be used for a range of other uses.

## Changes From Original Traefik Plugin

The primary change is to add support for `gzip` content. This brought another potential issue to mind, what about really large
content? This was handled as well.

### Process For Handling Body Content

#### Body Content Requirements

* The header must have `Content-Type` that includes `text`. For example:
  * `text/html`
  * `text/json`
* The header must have `Content-Encoding` header that is supported by this plugin
  * The original plugin supported `Content-Encoding` of `identity` or empty
  * This plugin adds support for `gzip` and `zlib` encoding

#### Processing Paths

* If the either of the previous conditions failes the body is passed on as is and no further processing from this plugin occurs.

* If the `Content-Encoding` is empty or `identity` it is handled in mostly the same manner as the original plugin.

* If the `Content-Encoding` is `gzip` the following process happens:
  * The body content is decompressed by [Go-lang's gzip library](https://pkg.go.dev/compress/gzip)
  * The resulting content is run through the `regex` process created by the original plugin
  * The processed content is then compressed with the same library and returned

## Configuration

### Static

```yaml
pilot:
  token: "xxxx"

experimental:
    plugins:
        rewrite-body:
            moduleName: "github.com/packruler/rewrite-body"
            version: "v0.5.0"
```

### Dynamic

To configure the `Rewrite Body` plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in 
your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). The following example creates
and uses the `rewritebody` middleware plugin to replace all foo occurences by bar in the HTTP response body.

If you want to apply some limits on the response body, you can chain this middleware plugin with the [Buffering middleware](https://docs.traefik.io/middlewares/buffering/) from Traefik.

```yaml
http:
  routers:
    my-router:
      rule: "Host(`localhost`)"
      middlewares: 
        - "rewrite-foo"
      service: "my-service"

  middlewares:
    rewrite-foo:
      plugin:
        rewrite-body:
          # Keep Last-Modified header returned by the HTTP service.
          # By default, the Last-Modified header is removed.
          lastModified: true

          # Rewrites all "foo" occurences by "bar"
          rewrites:
            - regex: "foo"
              replacement: "bar"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
```

## Example theme.park

### Dynamic

```yaml
http:
  routers:
    sonarr-router:
      rule: "Host(`sonarr.example.com`)"
      middlewares: 
        - sonarr-theme
      service: sonarr-service

  middlewares:
    sonarr-theme:
      plugin:
        rewrite-body:
          rewrites:
            - regex: </head>
              replacement: <link rel="stylesheet" type="text/css" href="https://theme-park.dev/css/base/sonarr/{{ env "THEME" }}.css"></head>

  services:
    sonarr-service:
      servers:
        - url: http://localhost:8989
```

You can set an environment variable `THEME` to the name of a theme for easier consistency across apps.

