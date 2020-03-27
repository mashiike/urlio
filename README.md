# urlio

![ci](https://github.com/mashiike/urlio/workflows/Go/badge.svg)
[![Documentation](https://godoc.org/github.com/mashiike/urlio?status.svg)](http://godoc.org/github.com/mashiike/urlio)
[![Go Report Card](https://goreportcard.com/badge/github.com/mashiike/urlio)](https://goreportcard.com/report/github.com/mashiike/urlio)


package urlio aims to provide io.Reader/io.Writer for resources corresponding to URLs.

**Note**: As of v0.0.0, it is still a Reader only implementation.

## Usage
see details in [godoc](http://godoc.org/github.com/mashiike/urlio)
```go
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mashiike/urlio"
)

func main() {
	reader, _ := urlio.NewReader("s3://example.com/example.txt")
	defer reader.Close()
	bytes, _ := ioutil.ReadAll(reader)
	fmt.Println(string(bytes))
}
```

## LICENSE
MIT
