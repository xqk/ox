package ocolor

import (
	"fmt"
	"testing"
)

func TestBlue(t *testing.T) {
	fmt.Println(Blue("hello "))
	fmt.Println(Blue("hello ", "world"))
}

func TestGreen(t *testing.T) {
	fmt.Println(Green("hello "))
	fmt.Println(Green("hello ", "world"))
}

func TestRed(t *testing.T) {
	fmt.Println(Red("hello "))
	fmt.Println(Red("hello ", "world"))
}
