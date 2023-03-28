package service

//go:generate ../.deps/mockgen -destination mock/fizzbuzz_service.go -source fizzbuzz_service.go

import (
	"strconv"

	"go.uber.org/zap"
)

type FizzBuzzService interface {
	SimpleFizzBuzz(limit, firstMod, sndMod int, firsStr, sndStr string) []string
}

type fizzBuzzService struct {
	logger *zap.Logger
}

func NewFizzBuzzService(logger *zap.Logger) FizzBuzzService {
	return &fizzBuzzService{
		logger: logger,
	}
}

func (fbs *fizzBuzzService) SimpleFizzBuzz(limit, firstMod, sndMod int, firsStr, sndStr string) []string {
	res := make([]string, limit)
	for i := 0; i < limit; i++ {
		nb := i + 1
		m1 := nb % firstMod
		m2 := nb % sndMod
		if m1 == 0 && m2 == 0 {
			res[i] = firsStr + sndStr
		} else if m1 == 0 {
			res[i] = firsStr
		} else if m2 == 0 {
			res[i] = sndStr
		} else {
			res[i] = strconv.Itoa(nb)
		}
	}
	return res
}
