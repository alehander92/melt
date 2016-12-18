# melt

melt compiles to Go.

I like the concurrency and ecosystem of Go and I like to use it for 
writing tools and web services.
However I don't like a lot of the language design choices.
melt fixes:

* Type system: Adds generics and sum types to the type system while remaining compatible to the Go ecosystem
* Error handling: Keeps Go error model but improves a lot the syntax for it
* Expresiveness: Easier with the error syntax

People usually defend Go choices like "it's simpler, everybody can grasp it in 2 hours". I love simplicity, but I think in that case it's on the expense of the power of the language. Also I find it hard to believe that many of the same people that are ready to invest huge amounts of their time to learn Git/Frameworks/Vim, can't take a little more time to learn the nuances of a more powerful language

# Records

```go
struct Message:
	Type  MessageType
	Value [8]byte
```

is equivalent to

```go
type Message struct {
	Type  MessageType
	Value [8]byte
}
```

```go
struct Stack<T>:
	Length 		int
	Capacity 	int
	Head 		*Node<T>
```

defines a generic struct. It's compiled to a go non-generic struct for each variation of the used concrete types in main.

E.g. if you used `Stack<int>` and `Stack<Vector<string>>` with safe_name=false you'll have

```go
type StackOfInt struct {
	Length 		int
	Capacity	int
	Head 		*NodeOfInt
}

type StackOfVectorOfString struct {
	Length 		int
	Capacity 	int
	Head 		*NodeOfVectorOfString
}
```

# Syntax:

Melt syntax is close to, but not the same as Go:

It's currently indentation-based, but that can change. We can easily 
use Go-style braces, but actually for now I prefer this difference, so
you can easily say if you're editing melt or go code.

# Optimized error syntax:

Error syntax in Go has those goals:

* Signify that a call can return an error
* Deal with this error in the calling function

The problem is, it forces you to do it immediately on the next 2-3 lines.
This makes composability and chaining of calls almost impossible very often and
it creates problems.

* Spaghetti code: You mix your error handling and your logic in a similar way to the php-html chaos in 90s
* It's actually not working for functions that return only an error

Functions that return another object are most often called like `obj, err := f()`
Your code show that it returns an error, even if you ignore it, you have to write `_`
Functions that don't return anything except an error are often just written `f()` and
Go doesn't even warn you you might miss it. 

Melt solution is:

* Functions always return only a value
* They can return an error, but you have to always signify this with '!' after their name
* You can and you actually *HAVE TO* handle their errors, but in `on` handlers after their calls, where you have the `$err` error variable

That satisfies Go requirements:

* Signify that a call can return an error
* Deal with this error in the calling function (improved)

And also separates them nice and clean

You can have your logic neatly expressed on several lines and *then* deal with
a possible error.

Of course, you can still do it after it like in normal Go.

# Generics

It's pretty simple, you can define generic functions, interfaces and structs
using <T>
