package domain

type MetricCounterScore struct {
	Key          string
	ScoreCounter int
}

type MetricCountersScores []MetricCounterScore

func (mcs MetricCountersScores) Keys() []string {
	keys := make([]string, len(mcs))
	for _, m := range mcs {
		keys = append(keys, m.Key)
	}
	return keys
}

func (mcs MetricCountersScores) Counters() []int {
	counters := make([]int, len(mcs))
	for i, m := range mcs {
		counters[i] = m.ScoreCounter
	}
	return counters
}

type MetricCountFizzBuzz struct {
	Key     string          `json:"-"`
	Score   int             `json:"counter"`
	Request FizzBuzzRequest `json:"request"`
}
