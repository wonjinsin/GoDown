// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"regexp"
)

func main() {
	r := regexp.MustCompile("\\w\\/([a-zA-Z0-9-_]+).([a-z0-9]+)")
	fmt.Println(r.FindStringSubmatch("https://www.test.kr/test-001.ts?test"))
}
