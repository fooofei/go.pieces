package main


import (
"fmt"
	"math"
)

//type ErrNegativeSqrt float64


func Sqrt(x float64) (float64, error) {
	if(x <0){
		return 0, fmt.Errorf("cannot Sqrt negative number: %v",x )
	}
	return math.Sqrt(x),nil
}

func main() {
	fmt.Println(Sqrt(2))
	fmt.Println(Sqrt(-2))
}

