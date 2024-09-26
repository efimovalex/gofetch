# GoHan(dler)s

GoHans is a lightweight, concise, robust and intuitive Go library designed to simplify making HTTP requests. Whether you're fetching data from an API, posting JSON or XML, or working with headers, GoHans offers a streamlined, easy-to-use interface that builds on Go’s native net/http package. With built-in support for TLS configuration, timeouts, retries, and automatic success/error message decoding, GoHans makes managing HTTP communication effortless. GoHans empowers developers to efficiently interact with web services while keeping code clean and concise. Perfect for building scalable and performant applications, GoHans makes HTTP requests in Go effortless.

GoHans is also perfect for use with the Adapter Design Pattern, making it easy to integrate with existing systems or switch between different HTTP clients. This flexibility allows developers to build modular, scalable, and maintainable applications. Whether you're building a microservice or a robust API client, GoHans ensures seamless HTTP communication.

## Usage 
For detailed usage examples, please refer to the [examples](./examples) package and accompanying test cases.

## Client features:

### TLS Configuration for Secure Connections

You can configure the client's TLS settings directly by loading a custom tlsConfig:

```
tlsc := gohans.TLSConfig(ctx, caCertPath, certPath, keyPath)
client := gohans.NewClient(ctx, WithTLSClientConfig(tlsc))
```

### Client Timeout Settings

Specify a custom timeout duration for the client to ensure timely responses:

```
client := gohans.NewClient(ctx, WithTimeout(time.Second)) // Sets a 1-second timeout
```

## Request features

### Expected success and error message variables
GoHans supports automatic decoding of the response body based on the status code. If the expected status code is received, GoHans decodes the success response body. In cases where the status code does not meet expectations, GoHans can automatically decode and handle error messages:

```
successBody := struct {
    Result string `json:"result"`
}{}

errorBody := struct {
    Error string `json:"error"`
}{}

b, err := gohans.NewRequest().
    SetExpectedStatusCode(http.StatusOK). // Sets the expected status code, default is 200 OK
    SetWantedResponseBody(&successBody).  // Decodes success response if expected status code is met
    SetErrorResponseBody(&errorBody).     // Decodes error response if the status code is unexpected
    ...
    .Send(ctx, client)
```

This feature simplifies error handling by providing separate decoding paths for success and error responses, ensuring smooth integration with APIs that follow standard HTTP status code conventions.

### Automatic Retries

GoHans supports automatic retries when the expected status code is not received or in the event of an error:

``` 
b, err := gohans.NewRequest().
   EnableRetries(3). // Retries the request up to 3 times
    ...
   .Send(ctx, client)
```

### JSON encoding & decoding
By default, GoHans is configured to send and receive data in JSON format, eliminating the need to manually set headers:

```
x := struct {
    Query string `json:"query"`
} {
    Query: "get ...."
}
wanted := struct{
    Result string `json:"result"`
}{}
b, err := gohans.NewRequest().
    SetRequestBody(x). // Automatically encodes the body as JSON
     ...
    SetWantedResponseBody(&wanted). // Automatically decodes JSON responses into the 'wanted' variable
    Send(ctx, client)
```

### XML support

GoHans also supports sending and receiving XML requests by setting the appropriate headers:

``` 
b, err := gohans.NewRequest().
    AddHeader("Content-Type", "application/xml"). // Encodes the request body as XML
    AddHeader("Accept", "application/xml").       // Instructs the server to respond with XML
    ...
    SetRequestBody(x). // Encoded as XML based on the "Content-Type" header
    SetWantedResponseBody(&wanted). // Decodes XML responses into the 'wanted' variable
    ...
    .Send(ctx, client)
```

### Authentication 

GoHans also supports setting an authentication token in the headers as a bearer token. For other authentication mechanisms, please use the AddHeader function:

``` 
b, err := gohans.NewRequest().
    SetAuthToken(token).
    ...
   .Send(ctx, client)
```

### Method selection 

Specify the HTTP method for each request:

``` 
b, err := gohans.NewRequest().
    SetMethod(http.MethodGet).
    ...
   .Send(ctx, client)
```

### URLs defined per request

You can define custom URLs on a per-request basis:

``` 
u := url.URL{
    Scheme: "http",
    Host: "localhost:8080",
    Path: "/test/1",
}

b, err := gohans.NewRequest().
    SetURL(u.String()).
    ...
   .Send(ctx, client)
```