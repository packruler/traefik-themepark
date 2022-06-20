# `theme.park` Traefik Plugin

Apply themes from [theme.park](https://theme-park.dev) to supported applications.

All credit for design and styling to the various contributors to
[https://github.com/GilbN/theme.park](https://github.com/GilbN/theme.park)!

## Features

Here is a list of features: (current [x], planned [ ], and potential `?`)

* [x] Support for all supported themes and apps in [theme.park](https://theme-park.dev)
* [x] Supports service side compression:
  - [x] `gzip` - gzip
  - [x] `deflate` - zlib
  -  ?  `br` - brotli (currently unsupported in [Yaegi](https://github.com/traefik/yaegi) for Traefik plugins)
* [x] Limits the HTTP queries which are touched by plugin to improve performance
* [x] Updates requests to limit requests' `Accept-Encoding` to include only supported systems

## Configuration

### Static

```yaml
pilot:
  token: "xxxx"

experimental:
    plugins:
        themepark:
            moduleName: "github.com/packruler/traefik-themepark"
            version: "v1.1.0"
```

### Dynamic

To configure the `theme.park` plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in
your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). The following example creates
and uses the `themepark` middleware plugin to replace all foo occurences by bar in the HTTP response body.

The `app` and `theme` values passed to the plugin should match the `<style>` tag provided in the `Installation` section.
The example below would match `Installation` section having the following `<style>` tag.

`<link rel="stylesheet" type="text/css" href="https://theme-park.dev/css/base/sonarr/dark.css">`



```yaml
http:
  routers:
    my-router:
      rule: "Host(`localhost`)"
      middlewares:
        - "sonarr-dark"
      service: "my-service"

  middlewares:
    sonarr-dark:
      plugin:
        themepark:
          # The name of the supported application listed on https://docs.theme-park.dev/themes.
          app: sonarr

          # The name of the supported theme listed on https://docs.theme-park.dev/theme-options/ or https://docs.theme-park.dev/community-themes/
          theme: dark
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
```

## How Does This Work?

This is an extension of the [rewrite-body](https://github.com/packruler/rewrite-body)
plugin I created based on Traefik's [plugin-rewritebody](https://github.com/traefik/plugin-rewritebody)
to add support for compressed content.

That said, this plugin is more focused on `theme.park` support and allows more targetted
middleware logic. This means the overhead added by the plugin's logic is very limited.
You can read more about that in the [process](#process) section.

### Process

#### Supported Requests

For any updates to be attempted the following conditions must be met by the incoming request:

- `Accept` header must include `text/html`
- HTTP `Method` must be `GET`

These conditions are intended to drastically limit the HTTP queries that are touched by this plugin.
At this time these conditions properly cover all tested applications.

#### Supported Responses

Assuming [Supported Request](#supported-requests) conditions have been met, the following conditions must
be met by the resulting response:

- `Content-Type` must be `text/html`
- `Content-Encoding` must be a support compression (`gzip`, `deflate`, or `identity`)
