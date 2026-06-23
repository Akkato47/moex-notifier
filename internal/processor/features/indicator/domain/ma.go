package domain

type MAState struct {
	period int
	values []float64
	sum    float64
}

func NewMAState(period int) *MAState {
	return &MAState{period: period, values: make([]float64, 0, period)}
}

func (m *MAState) Update(price float64) float64 {
	if len(m.values) >= m.period {
		m.sum -= m.values[0]
		m.values = m.values[1:]
	}
	m.values = append(m.values, price)
	m.sum += price
	return m.sum / float64(len(m.values))
}

func (m *MAState) IsReady() bool {
	return len(m.values) >= m.period
}

func (m *MAState) Value() float64 {
	if len(m.values) == 0 {
		return 0
	}
	return m.sum / float64(len(m.values))
}
