# Request client



## Usage 

```golang 
package main

import (
    "github.com/efimovalex/gofetch"
)

func main() {
    ... 
    // Load the TLS configuration (optional)
    tlsConf, err := tls.TLSConfig(ctx, ca, cert, key, true)
    if err != nil {
        logger.Error("Failed to create TLS config", "error", err)

        return
    }
    // Create a new client with the TLS configuration
    client := NewClient(ctx, WithTLSClientConfig(tlsConf))

    // Create a new request
    
    
}
```