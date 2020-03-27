package urlio_test

import (
	"log"

	"github.com/mashiike/urlio"
)

func ExampleNewReader() {
	reader, err := urlio.NewReader(
		urlio.MustParse("https://www.google.com/"),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer reader.Close()
}
