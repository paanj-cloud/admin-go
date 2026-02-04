# Paanj Admin SDK for Go

Official Go Admin SDK for Paanj - Server-side administration and management.

[![Go Reference](https://pkg.go.dev/badge/github.com/paanj-cloud/admin-go.svg)](https://pkg.go.dev/github.com/paanj-cloud/admin-go)

## Installation

```bash
go get github.com/paanj-cloud/admin-go@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    admin "github.com/paanj-cloud/admin-go"
)

func main() {
    // Initialize admin client
    paanjAdmin := admin.NewAdmin("your-secret-key", admin.AdminOptions{
        ApiUrl: "https://api1.paanj.com",
        WsUrl:  "wss://ws1.paanj.com",
    })

    // Connect to admin WebSocket
    if err := paanjAdmin.Connect(); err != nil {
        log.Fatal(err)
    }
    defer paanjAdmin.Disconnect()

    fmt.Println("Connected to Paanj Admin!")
}
```

## Features

- ✅ Server-side authentication with secret key
- ✅ Admin WebSocket connections
- ✅ HTTP API access
- ✅ Real-time admin events
- ✅ Automatic reconnection

## API Reference

### Admin Initialization

```go
admin := admin.NewAdmin(secretKey, admin.AdminOptions{
    ApiUrl:               "https://api1.paanj.com",
    WsUrl:                "wss://ws1.paanj.com",
    AutoReconnect:        true,
    ReconnectInterval:    5 * time.Second,
    MaxReconnectAttempts: 10,
})
```

### WebSocket Connection

#### Connect

```go
err := admin.Connect()
```

#### Disconnect

```go
admin.Disconnect()
```

#### Check Connection Status

```go
isConnected := admin.IsConnected()
```

### Event Handling

```go
admin.On("admin.event", func(data interface{}) {
    fmt.Printf("Admin event: %+v\\n", data)
})
```

### HTTP Requests

```go
httpClient := admin.GetHttpClient()
result, err := httpClient.Request("GET", "/api/v1/admin/users", nil)
```

### Subscriptions

```go
err := admin.Subscribe(map[string]interface{}{
    "resource": "conversations",
    "events":   []string{"conversation.created", "conversation.updated"},
})
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    admin "github.com/paanj-cloud/admin-go"
)

func main() {
    // Initialize admin client
    paanjAdmin := admin.NewAdmin("sk_live_your_secret_key", admin.AdminOptions{
        ApiUrl:            "https://api1.paanj.com",
        WsUrl:             "wss://ws1.paanj.com",
        AutoReconnect:     true,
        ReconnectInterval: 5 * time.Second,
    })

    // Connect to admin WebSocket
    if err := paanjAdmin.Connect(); err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    defer paanjAdmin.Disconnect()

    fmt.Println("✅ Connected to Admin WebSocket")

    // Listen for admin events
    paanjAdmin.On("user.created", func(data interface{}) {
        fmt.Printf("📢 New user created: %+v\\n", data)
    })

    paanjAdmin.On("conversation.created", func(data interface{}) {
        fmt.Printf("📢 New conversation: %+v\\n", data)
    })

    // Subscribe to all user events
    err := paanjAdmin.Subscribe(map[string]interface{}{
        "resource": "users",
        "events":   []string{"user.created", "user.updated", "user.deleted"},
    })
    if err != nil {
        log.Printf("Subscription failed: %v", err)
    }

    // Make HTTP API request
    httpClient := paanjAdmin.GetHttpClient()
    users, err := httpClient.Request("GET", "/api/v1/admin/users", nil)
    if err != nil {
        log.Printf("API request failed: %v", err)
    } else {
        fmt.Printf("Users: %+v\\n", users)
    }

    // Keep alive
    fmt.Println("Listening for admin events... (Press Ctrl+C to exit)")
    select {}
}
```

## HTTP API Examples

### GET Request

```go
httpClient := admin.GetHttpClient()
result, err := httpClient.Request("GET", "/api/v1/admin/conversations", nil)
```

### POST Request

```go
data := map[string]interface{}{
    "name": "New Conversation",
}
result, err := httpClient.Request("POST", "/api/v1/admin/conversations", data)
```

### DELETE Request

```go
result, err := httpClient.Request("DELETE", "/api/v1/admin/users/123", nil)
```

## Error Handling

```go
if err := admin.Connect(); err != nil {
    log.Printf("Connection error: %v", err)
    return
}

result, err := httpClient.Request("GET", "/api/v1/admin/users", nil)
if err != nil {
    log.Printf("API error: %v", err)
    return
}
```

## Security Best Practices

1. **Never expose secret keys** in client-side code
2. **Use environment variables** for secret keys
3. **Rotate keys regularly**
4. **Use HTTPS/WSS** in production

```go
import "os"

secretKey := os.Getenv("PAANJ_SECRET_KEY")
admin := admin.NewAdmin(secretKey, admin.AdminOptions{
    ApiUrl: "https://api1.paanj.com",
    WsUrl:  "wss://ws1.paanj.com",
})
```

## License

MIT License - see LICENSE file for details.

## Support

- Documentation: https://docs.paanj.com
- Issues: https://github.com/paanj-cloud/admin-go/issues
- Email: support@paanj.com
