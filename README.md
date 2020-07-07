![](https://ipstack.com/ipstack_images/ipstack_logo_color.png)

> Locate and identify website visitors by IP 
> ipstack offers one of the leading IP to geolocation APIs and global IP database services worldwide.

(from https://ipstack.com/)
Unofficial Golang [ipstack](https://ipstack.com/) library.

[![GoDoc](https://godoc.org/github.com/qioalice/ipstack?status.svg)](https://godoc.org/github.com/qioalice/ipstack)
[![Go Report Card](https://goreportcard.com/badge/github.com/qioalice/ipstack)](https://goreportcard.com/report/github.com/qioalice/ipstack)

---

# Take it
```
go get github.com/qioalice/ipstack
```
    
# Initialize the package

Just call `Init` function and pass API token as `string`
```go 
ipstack.Init("token")
```
`Init` may return an `error` if you pass an invalid API token or if it can perform first test query (get info about your IP address).

You can handle it, of course, or disable performing first test query (for diff reasons) using constructor parameter `ParamDisableFirstMeCall` (read about it below).

```go
if err := ipstack.Init("token"); err != nil { 
    panic(err) 
}
```

# Want to get info about one IP?

Just call `IP` function and pass IP address info about you want to know as string to the that call.
```go
if res, err := ipstack.IP("8.8.8.8"); err == nil { ... }
```

# Response? Output data? IP info?

When you calling `ipstack.IP` or `ipstack.Me` you will get `*Response` and `error` objects, when `ipstack.IPs` - `[]*Response` and `error`.

Anyway, `Response` contains an info about one requested IP address. This class don't have any methods and just have all public fields that you can check to get all desired info 

But some info can be represented as pointers to the some internal auxiliary response instances: `tResponseCurrency` for `Currency` field, for example, or `tResponseTimeZone` for `TimeZone` field. It was made 'cause sometimes ipstack willn't provide you the whole information about IP - the all that he's can provide: may be you have limitiations on your account, or requested only 2 or 3 fields instead all.

```go
if res, err := ipstack.IP("8.8.8.8"); err == nil {
    fmt.Printf("My IP: %s, My country: %s (%s)\n", res.IP, res.CountryName, res.CountryCode)
}
```

# Bulk queries?

First, be sure that your account supports the bulk queries (starts from _professional_ tariff). You can read check it and read about it [here](https://ipstack.com/product/)

So, just call `IPs` function and pass as many IP addresses as you want but not more than 50 (ipstack limitation).
```go
if ress, err := ipstack.IPs("1.2.3.4", "8.8.8.8", ...); err == nil {
    for _, res := range ress {
        // do smth with each response
    }
}
```

# Your IP address?

When you initialize package or create `Client` object (read about it below) if you don't disable _first test query_, you already can get instant info about your IP address. It already stored in the `Client` object.

So, you have `Me` function. If `Client` object already have info about your IP, `Me` just return it. Want a fresh info and overwrite cache? `Me` takes unnecessary `bool` argument, which makes the `Me` function delete cached info and request the fresh if it will be `true`.

```go
me, err := ipstack.Me() // get cached info
me, err = ipstack.Me(true) // fetch the fresh info and save it to the cache
```

# What is `Client`? And what's diff between `New` and `Init`

1. **There is no functions. Only methods.**
When you writing `ipstack.IP(...)` it means `ipstack.DefaultClient.IP(...)`, and as you can guess, `Init` just initializes (calls `New` and stores the result) `DefaultClient` variable. Yes, each package level function (except `New` and `Init` of course) just an alias to the some method of `DefaultClient` object that has `Client` type.

2. **What is `Client`?**
Almost wrapper over `http.Client` object. To be more precision, the flexible way to perform the diff requests with one token (diff fields, diff checking error, diff allocating memory, etc).

3. **First `New` call includes the `Init` call.**
`New` creates a new `Client` instance with specified parameters. And if `ipstack.DefaultClient` object is nil, also store it to that.

So. Wanna use a few `Client`s with the diff tokens? I'm not against.

```go
cli1, err1 := ipstack.New("token1") // 'cli1' also will be stored to the 'ipstack.DefaultClient'
cli2, err2 := ipstack.New("token2")
// ipstack.IP(...) call is the same as cli1.IP(...) (cause 'cli1' created by first calling of 'New')
```

# Constructor (`New`, `Init`) arguments (`tClientParam`)

When you create a `Client` object using `New` function or initialize package level default client using `Init` method, you can pass arguments to the that functions.

I called them "constructors". So, for example you can pass API token as `string` directly or wrap it to the `tClientParam` object using `ParamToken` function.

They allows you to specify behaviour what you want.

```go
ipstack.Init(ipstack.ParamToken("token"))
```

`tClientParam` arguments always overwrite arguments that has been passed as not `tClientParam`. Thus, in the example below, default client will be initialized with `"token1"`

```go
ipstack.Init(ipstack.ParamToken("token1"), "token2")
```

Find your favorite parametrizers:

| Parametrizer<br>Arguments | Description |
| --- | :--- |
| `ParamToken`<br>`string` | The ipstack API token, `Client` will be created with. And each request from this `Client` will perform with.
| `ParamDisableFirstMeCall`<br>none | Disables the **first test query** while creating the `Client` object. **First test query** is the request info about **your IP** address. If this request will be successfull, it means that `Client` instance created and initialized successfully. And result of that request stores to the internal cache and will be available using `Me` method instantly.
| `ParamUseHTTPS`<br>`bool`| Switches the schema of Web API requests. `true` means "use **HTTPS**" and `false` means "use **HTTP**" respectively.<br>**Warning!** You can use HTTPS only on a non-free tariffs! You can check it and read about it [here](https://ipstack.com/product/).
| `ParamFields`<br>`string...` | Specify what kinds of IP's info you want to get from ipstack. You can use predefined constants which starts from `Field` word and pass constants only of that fields, what kind info you want know.<br>**Warning!** Some fields requires diff tariff plans. You can check it and read about it [here](https://ipstack.com/product/).
| `ParamEnableSecurity`<br>`bool` | Enables security module.<br>**Warning!** Security module requires diff tariff plans.


And, for example, it looks like:

```go
cli, err := ipstack.New(
    ipstack.ParamToken("token"), // specify token
    ipstack.ParamDisableFirstMeCall(), // disable first test query
    ipstack.ParamUseHTTPS(true), // force use HTTPS instead HTTP
    ipstack.ParamFields(ipstack.FieldCountryName, ipstack.FieldCurrency), // get only country name and currency info
    ipstack.ParamEnableSecurity(true) // enables security module (account must support this)
)
```

# Errors and API errors (`tError`)

Each `IP`, `IPs` or `Me` (or `New`/`Init` with enabled first test query) may return an error object (as only return argument or as a second, depends by method) (only if you're not calling `tRequest` methods directly, but more about that below).

In Golang, each error object represents by `error` interface, and error object will be represented by that type.

But you can get **API error**. 

**API error** is the special error type which signals that technically request will be successfully sent and response will be successfully received and decoded. But contains not an info about requested IP(s) but error message.

In that case you must handle that error as you want. That error represents by `tError` class, that as you guess, implements the `error` interface.

At your disposal `APIError` function that takes an argument of `error` type and returns the `*tError` not nil object if `error` is `tError` object, and nil if `error` isn't (and it doesn't matter `error` is nil or not). 

So, now you have `tError` object. Just check "code", "type", "info" entities to know what kind of error is occur. Use `Code`, `Type` and `Info` methods for that. 

And at your service, you can call any of these methods from `nil` `tError` object. It means, that if `APIError` will return `nil`, you can chain directly `Code` or `Type` or `Info`, w/o checking whether returned object is nil or not.

If `tError` object is `nil`, `Code` will return `0`, and `Type` and `Info` an empty `""` string. 

```go
res, err := ipstack.IP("8.8.8.8")
if err := ipstack.APIError(err); err.Code() != 0 {
    // API error is occurred
    panic(err.Type() + " " + err.Info())
}
if err != nil {
    // Unknown error is occurred
    panic(err)
}

```

# `R` method and `tRequest`, `tResponse` classes

Want more flexibility? Get it! <br>
Worry about each allocated byte? Save it! <br>
Want another JSON decoder or/and API error checker? Use it!

Each request to the ipstack Web API represents by `tRequest` object. Each response by `tResponse`. No exceptions.

Moreover, `Client` doesn't contain any behaviour definitions but contains the default request object (`tReqeust` type), and each call 'IP', 'IPs' or 'Me' method of 'Client' class it's just calling the same methods of default `tRequest` object. But it can be modified! Even already! Constructor parametrizers have the almost same logic.

##### What you should know?

1. **`tRequest` methods return `tResponse` objects. Always.**
This is a lower level. The price for more flexibility. `tResponse` objects contains RAW not decoded JSON data by default (`RawData` field) and error object (`Error` field) that represents some request or response error. But it guarantees, that if `RawData` is nil, `Error` isn't and vice versa.

2. **When you get `tRequest` object, all what you've done with it, willn't apply to the `Client` object from which you got `tRequest` object.**
You can request some different fields for one request, or made it with HTTPS instead HTTP, why not? And any change of behaviour don't saved anywhere except `tReqeust` object you're working with. But by default this is just a copy of default request of `Client`.

3. **Don't forgot whether response contains API error.**
As you know, `tResponse` object contains not decoded JSON RAW data as `RawData` field. But API may return encoded JSON error. You must check it. Or use internal function `CheckError`. So, `Client` methods `IP`, `IPs`, `Me`, if you'd see, just calling the same methods of `tRequest` and then checks error using `CheckError` method of `tResponse` and decode JSON using `DecodeTo` method (of `tResponse` too). You can use `CheckError` method, or do it the way you want.

4. **How to decode JSON?**
This is the finish step. May be checking error and decoding JSON in your logic is the one step, but I prefer to split these steps.
<br>So, you can use `DecodeTo` method, that receives only one argument - the destination object. By default it just calls the `json.Unmarshal` function with `tResponse.RawData` and received destination argument. But you can decode as you want - by custom JSON decoder, with the saving each unneccessary byte, with writing a very RAM-efficiency algorithm.

##### How use it?

1. **Call `R` method of some `Client` object or call `R` package function.**
You'll get `tRequest` object (copy of base client's request object) with which you can do the next steps.

2. **Change behaviour of `tRequest`.**
You can use any of described below `tRequest` method to change its behaviour.

3. **Perform request.**
Call `IP`, `IPs` or `Me` method, save the result (`tResponse` instance).

4. **Check error and decode JSON response.**
Use `CheckError`, `DecodeTo` methods of `tResponse` or use your personal way.

```go
// 'cli' will always perform requests over HTTPS and with default set of fields
cli, err := ipstack.New("token", ipstack.ParamUseHTTPS(true)) 
// but we want perform a few queries over HTTP and with only one field - ip's country name

// we can save cretated 'tRequest' object and then perform queries
req := cli.R().UseHTTPS(false).Fields(ipstack.FieldCountryName) // 'req' type is '*tRequest'
resp1 := req.IP("8.8.8.8") // 'resp' type is '*tResponse', query over HTTP and with only one field
resp2 := req.IP("1.2.3.4") // the same as above
resp3 := cli.IP("1.2.3.5") // will be over HTTPS and with default fields

// or don't save 'tRequest' object and perform query right away
// but in this way we should change behaviour to the desired each time
resp1 = cli.R().UseHTTPS(false).Fields(ipstack.FieldCountryName).IP("8.8.8.8")
resp2 = cli.R().UseHTTPS(false).Fields(ipstack.FieldCountryName).IP("1.2.3.4")
resp3 = cli.IP("1.2.3.5") // will be over HTTPS and with default fields

// then we must check errors
if err := resp1.CheckError(); err != nil {
    if err := ipstack.APIError(err); err.Code != 0 {
        // Some API error, check 'err.Code(), err.Type(), err.Info()'
    } else {
        // Some another error
    }
}
// and decode JSON (for example to the map)
mresp := map[string]interface{}{}
if err := resp1.DecodeTo(&mresp); err != nil {
    // Decode error
}

// or may be you want check error and decode JSON manually over each byte?
for _, b := range resp1.RawData { ... } // 'resp1.RawData' type is '[]byte'
```

# To Do

**The current version of this library is beta.**
- There is no **tests** (they're exists, but I don't commit them now, I'll do it later, when I'll documented them and make them more cute).
- Also, there is no **examples**. I tried to write as much code examples as I can, but anyway, there's no complex example. One or more.
- And, of course, may be I'll add some **another functional** to the libarary, but atm I dunno what else I can add.
<br>But. **I definitely won't break the package API.**

# License

qioalice@gmail.com, Alice Qio, 2019.
<br>MIT
