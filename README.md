# go-congen

Go HTTP controller generator. (see [example](example))

Can be used as CLI or library.

Utility:
* parses HTML file
* detects all forms
* generates endpoints and stubs

By-default, if no response sent, 303 See Other will be sent
to client.

## Installation

MacOS

    brew install reddec/tap/go-congen

From binary - check [releases](https://github.com/reddec/go-congen/releases) section

As CLI from source

    go get github.com/reddec/congen/cmd/...

## Features

### Multiple forms with same action

It supports multiple forms with same action and different fields set.
Fields will be merged to one structure:

**index.html**

```html
<form action="something" method="post">
    <input type="text" name="field1"/>
</form>
<form action="something" method="post">
    <input type="text" name="field2"/>
</form>
```

**gen.go**

```go
// ...

type Controller interface {
    // ...
	DoSomething(writer http.ResponseWriter, request *http.Request, params SomethingParams) error
    // ...
}
// ...
type SomethingParams struct {
	Field1 string
	Field2 string
}
// ...
```

### Multiple fields in the same form

In case of (un)intentional definition of fields with same names within one form
only first field will be used.