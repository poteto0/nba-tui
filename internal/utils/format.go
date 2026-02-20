package utils

import "fmt"

func PtrToIntStr(i *int) string {
	if i == nil {
		return "0"
	}
	return fmt.Sprintf("%d", i)
}

func PtrToFloatStr2f(f *float64) string {
	if f == nil {
		return "0"
	}
	return fmt.Sprintf("%.0f", *f)
}

func PtrToPctStr(f *float64) string {
	if f == nil {
		return "0.0"
	}
	return fmt.Sprintf("%.1f", *f*100)
}

func FormatScore(score int) string {
	s := fmt.Sprintf("%d", score)
	if len(s) == 1 {
		return " " + s + " "
	}
	if len(s) == 2 {
		return " " + s
	}
	return s
}
