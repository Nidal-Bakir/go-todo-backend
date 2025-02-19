package utils

type PhoneNumber struct {
	CounterCode string
	Number      string
}

func (p PhoneNumber) ToString() string {
	return p.CounterCode + p.Number
}

func (p PhoneNumber) IsValid() bool {
	// TODO: implement this somehow!!
	return true
}
