# gotemplate
A golang preprocessor for generating template instantiations of types and functions

### Install
```
go get -u github.com/redefiance/gotemplate
```

### Example

Create a .go file within your project directory, e.g. `CircularBuffer.go`:

```go
// +gotemplate

package main

type CircularBuffer_T struct {
	data   []T
	curPos int
}

func newCircularBuffer_T(size int) {
	return CircularBuffer_T{data: make([]T, size)}
}

func (b CircularBuffer_T) Push(value T) {
	b.curPos++
	if b.curPos >= len(b.data) {
		b.curPos = 0
	}
	b.data[b.curPos] = value
}

func (b CircularBuffer_T) At(index uint) T {
	pos := b.curPos - index
	for pos < 0 {
		pos += len(b.data)
	}
	return b.data[pos]
}
```

Type and function names must end with `_T`, method names must not.

Now, you can use the new type in another file:

```go
var buf = newCircularBuffer_uint64(10)
```

Run `gotemplate` from your project directory and it will instantiate your templates for every type referenced by the source code in `CircularBuffer_impl.go`.

### Flags

* `-d path`: uses `path` instead of the current working directory to search for templates  
* `-r`: generates templates in all imported packages recursively  

### Limitations

* Only works with 1 Template Parameter  
* Doesn't work with imported types (e.g. `CircularBuffer_os.FileInfo` is not allowed). Wrap imported types into a package-local type if you want to use them.  
* Doesn't work with both value and pointer types at the same time: You have to specify in the template if you want to use `T` or `*T`
