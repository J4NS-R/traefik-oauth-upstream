# Upstream OAuth - Traefik Middleare

This middleware adds OAuth headers to your requests so that for the upstream (service) the request is  
OAuth-authenticated. With other middleware you can configure any kind of downstream (client) authentication 
(E.g., [Basic Auth](https://doc.traefik.io/traefik/middlewares/http/basicauth/)) or leave it open to the
internet! (not recommended)

After the client has signed in, tokens are kept cached and are automatically refreshed. 

## Typical flow

```mermaid
sequenceDiagram
  participant B as Downstream client
  participant O as OAuth Provider
  participant P as Traefik OAuth Plugin
  participant U as Upstream server

  alt First ever request
    B->>P: Plain request
    P->>B: 302
    B->>O: Auth request
    O->>B: Success redirect
    B->>+P: OAuth callback
    Note right of P: Token & refresh token stored
    P->>-B: Redirect back to original request
  end

  alt Token still valid
    B->>+P: Plain request
    Note right of P: Bearer token added
    P->>-U: Authorised request
    U->>P: Response
    P->>B: Response
  end

  alt Token expired
    B->>P: Plain request
    P->>O: Refresh token
    O->>+P: Refreshed tokens
    Note right of P: Tokens updated and bearer added
    P->>-U: Authorised request
    U->>P: Response
    P->>B: Response
  end
  
```

## Config

You can set up different upstream OAuths by configuring different middlewares, or you can configure one middleware
and reuse it with multiple routers/services.

[Config example](docker-dev/dynamic-config-example/oauthup.yml)

## Development

This repo is GitPod friendly.
