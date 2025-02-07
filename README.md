# golibmagic

A pure Go implementation of the libmagic library, forked from [itchio/wizardry](https://github.com/itchio/wizardry).  This fork addresses long-standing maintenance issues and bugs in the original project.

## About

`golibmagic` provides file type detection based on magic numbers. It's a port of the functionality found in the widely-used `libmagic` C library, but implemented natively in Go, eliminating the need for cgo dependencies. This makes it easier to cross-compile and deploy, and generally more Go-friendly.

It contains:

  * A parser, which turn magic rule files into an AST
  * An interpreter, which identifies a target by following
  the rules in the AST
  * A compiler, which generates go code to follow the
  rules in the AST


The original `itchio/wizardry` repository has been unmaintained for several years, leading to accumulated bugs and a lack of support for newer magic database formats. This fork aims to:

* **Refactor the code:** Improve code readability, maintainability, and performance.
* **Fix bugs:** Address known issues and improve overall stability.
* **Maintain compatibility:**  Strive to maintain compatibility with the core libmagic functionality and magic database format.
* **Provide up-to-date support:** Keep the library up-to-date with the latest magic database definitions and Go best practices.

## Installation

```bash
go get [github.com/postfix/golibmagic]
```

## Usage

```go 
package main

import (
        "fmt"
        "github.com/postfix/golibmagic"
        "log"
)

func main() {
        m, err := golibmagic.New()
        if err != nil {
                log.Fatal(err)
        }
        defer m.Close()

        file := "path/to/your/file"
        fileType, err := m.LookupFile(file)
        if err != nil {
                log.Fatal(err)
        }

        fmt.Printf("File type: %s\n", fileType)


    // Alternatively, use a byte slice:
    data := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff} // Example GIF header
    dataType, err := m.Lookup(data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Data type: %s\n", dataType)

}
```
## License

wizardry is released under the MIT license, see the
`LICENSE` file for details.

