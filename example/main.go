package example

import "fmt"

type God struct {
	Name string
}

var OGod = newObservable_God(God{"Zeus"})

func init() {
	observer := OGod.Observe(func(deity God) {
		fmt.Printf("All hail to %s, our new god!\n", deity.Name)
	})

	OGod.Set(God{"Flying Spaghetti Monster"})

	observer.Close()
}
