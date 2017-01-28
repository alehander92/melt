```ruby
func Filter?<T>(test? T -> bool, sequence Sequence<T>) []T:
	result = [] as []T
	for item in sequence:
	    if test(item):
	    	result.push(item)

	escalate test
	return result
```

## Generics

Generic functions are defined using a declaration of the generic args and a normal signature.
The generic args look like `<T, U>`, `<T>` etc. They aren't needed the call.

```ruby
a := make([]A, 20)
i := make([]string, 20)
l := Filter(isWeird, a)
m := Filter!(isBig!, i)
escalate filter
puts(l)
puts(m)

```ruby
{
	"filter": {
		{"T": Basic("A"), "sequence": Slice(Basic("A")), "error:0": ErrorHint('')},
		{"T": Basic('B'), "sequence": Slice(Basic("B")), "error:0": ErrorHint('!')}
	}
}
```

```ruby
func Filter0(test: func(A) bool, sequence: []A) []A {
	result := []A{}
	for _, item := range sequence {
		if test(item) {
			result = append(result, item)
		}
	}
	return result
}

func Filter1(test: func(B) (bool, error), sequence: []B) ([]B, error) {
	result := []B{}
	for _, item := range sequence {
		testVar, testErr := test(item)
		if testErr != nil {
			return testErr
		}
		if testVar {
			result = append(result, testVar)
		}
	}
}
```

