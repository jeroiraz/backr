### Backward Reader

This buffered reader is optimized for reading append-only files with dangling links to previous content.

Chunks of data are read starting at the end but interpretation of data takes is done from left to right. Thus,
this reader can be used as `bufio.Reader` but internal buffering is filled with content at lower offsets.

This backward reader may be used with `size endian` records, where the field storing the size of the record is
at the end.

### Installation

```go
go get github.com/jeroiraz/backr
```


### Example

```go
package main

import (
  "os"
  "github.com/jeroiraz/backr"
)

func main() {
  f, _ := os.Open("file")
  defer f.Close()

  reader, _ := backr.NewFileReader(f)

  buf := make([]byte, 32)
  reader.Read(buf)
}
```