# Rewrite Header

Rewrite header is a middleware plugin for [Traefik](https://traefik.io) which replace a header in the response

## Configuration

### Static

```yaml
pilot:
  token: "xxxx"

experimental:
  plugins:
    rewriteHeaders:
      modulename: "github.com/XciD/traefik-plugin-rewrite-headers"
      version: "v0.0.4"
```

### Dynamic

To configure the Rewrite Head plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). 
The following example creates and uses the rewriteHeaders middleware plugin to modify the Location header

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "rewriteHeaders"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    rewriteHeaders:
      plugin:
        rewriteHeaders:
          rewrites:
            - header: "Location"
              regex: "^http://(.+)$"
              replacement: "https://$1"
```

## Features

- **Header Rewriting**: Rewrite response headers using regex patterns
- **Multiple Rewrites**: Apply multiple header rewrites in a single middleware
- **Validation**: Configuration validation ensures all rewrites have required header and regex fields
- **Standards Compliance**: Proper HTTP response writer interface implementation including support for implicit writes (auto-calls WriteHeader when content is written)
- **Interface Support**: Supports HTTP Hijacker, Flusher, Pusher, and ReaderFrom interfaces

## Error Handling

The plugin validates configuration at initialization and will return errors for:

- Missing or nil configuration
- Missing or nil next handler
- Empty header field in rewrite rules
- Empty regex pattern in rewrite rules
- Invalid regex patterns (compilation errors)

All configuration errors are returned with descriptive messages to aid debugging.

## Recent Improvements

- Added comprehensive config validation with clear error messages
- Fixed implicit response write handling (auto-calls WriteHeader on Write)
- Implemented proper response writer interface delegation
- Added support for HTTP Pusher and ReaderFrom interfaces
- Extended test coverage including validation and edge cases (64.6% coverage)
