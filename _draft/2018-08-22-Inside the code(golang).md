

### Summary
Implement one thing within 50 lines by golang just for fun.

### Build & Run

```sh
# install golang
brew install golang

# build and run source code
go run xxx.go

```
### Test
1. [Reverse words](http://yangjunsss.github.io/_codes/reverse_str.go)
* [Reverse list](http://yangjunsss.github.io/_codes/reverse_list.go)
* [Go through BTree](http://yangjunsss.github.io/_codes/gothrough_btree.go)
* [Point vs Value method](http://yangjunsss.github.io/_codes/usepfunc.go)
* [sliceTricks]
* [2 BTree equivalent]
* [2 BTree subtree equivalent]
* [bubble sort]
* [select sort]
* [fast sort]
* [shell sort]
* [bucket sort]

### Language overview
1. Only for, no while
* Entry point is main function within main package
* go run/build source code
* short variable declaration :=, only be used inside functions
* multiple values function return
* use struct, no class
* don't need to break for each switch case, switch with no value means switch true, fallthrough uses through multiple case
* method for pointer receiver or value receiver
* structure has no constructors method
* new(X) is the same as &X{}
* Composition better than inheritance
* operator precedence differences
* Go doesn't support overloading
* The 'case' blocks in 'swtich' break by default
* Pointers versus Values, if you don't know, use pointer, value is a great way to make data immutable
* iteration over maps isn't ordered
* cyclical imports isn't allow
* types and functions are visible outside of a package by first uppercase letter
* go get always points to master/head, otherwise, use godep to solve revision
* interface is composition
* go preferred way to deal with errors is through return values, not exceptions
* use defer xxx.xxx
* panic is like throwing an exception,, recover is like catch
* initialized IF, if [init]; condition {}
* conversion use .(TYPE)
* string use bytes, runes which are unicode code point
* function are first-class types
* goroutines is similar to a thread but it's scheduled by Go, not the OS
* use channel/sys.Mutex/sys.RWMutex to resolve synchronization and coordination goroutines
* buffer channel for not blocking client and storing more data
* mangle multiple channels and drop messages using select

### Common mistake
1. Strings can't be "nil"
* Array function arguments:Arrays in Go are values, so pass arrays to functions the funcions get a copy of the original array data
* Unexpected values in range clauses
* Accessing non-existing map keys
* Strings are immutable
* Built-in data structure operations are not synchronized
* Iterating a Map will be random
* Methods with value receivers can't change the original value
* JSON encoder adds a newline character
* Use UTF-8 or byte[] for JSON string value
* Use == operator to compare struct variables if each struct field can be compared with the == operator,if any field is not comparable then it will result in compile error
* recover() only works when it's done in a deferred function
* The value generated in range are copies of the collection elements, but you can change the value when you use pointer within the collection
* When you reslice a slice, the new slice will reference the original slice
* Slice data corruption, use copy or full slice to alloc new slice
* append return a new collection, is different with reslice
* break without a label only gets you out of inner switch/select
* iteration variable in for are reused in each iteration passed to goroutines
* deferred calls are executed at the end of function not the iteration
* func destroyed but not wait goroutine release
* The first use of iota doesn't always start with zero
* interface and map element is not addressable means you can't call pointer method directly
* stack and head variables: you don't always know if your variable is allocated on statck or heap. The compiler picks the location to store the variable based on its size and the result of "escape analysis", so it's ok to return references to local variables, which is not ok in other language like C or C++
* GOMAXPROCS, the max value used to be 256 in v1.10,1024 in v1.9
* Can't call c function with variable arguments
* Can't assign slice directly to interfaceSlice due to cannot the memory size

### Tips
