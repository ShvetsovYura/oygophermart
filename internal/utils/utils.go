package utils

import (
	"errors"
	"strconv"
)

func CheckLuhnFromStr(code string) (bool, error) {
	if len(code) < 1 {
		return false, errors.New("input seq has not empty")
	}
	var res = make([]uint8, 0, len(code))

	for _, el := range code {
		s := string(el)
		d, err := strconv.Atoi(s)
		if err != nil {
			return false, err
		}
		if d < 0 || d > 9 {
			return false, errors.New("digit <0 or >9")
		}
		res = append(res, uint8(d))
	}
	isValid := CheckLuhn(res)
	return isValid, nil
}

func CheckLuhn(code []uint8) bool {
	var sum int
	var d1 uint8
	parity := (len(code)) % 2

	for i, d := range code {
		if i%2 == parity {
			d1 = d * 2
			if d1 > 9 {
				d1 = d1 - 9
			}

		} else {
			d1 = d
		}
		sum += int(d1)
	}
	return sum%10 == 0
}
