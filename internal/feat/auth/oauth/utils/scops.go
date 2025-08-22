package oauth

import "slices"

type Scops struct {
	scopsSlice []string
}

func (s Scops) Array() []string {
	return s.scopsSlice
}

func (s Scops) Equal(o Scops) bool {
	return slices.Equal(s.scopsSlice, o.scopsSlice)
}

func (s Scops) EqualArray(o []string) bool {
	return s.Equal(*NewScops(o))
}

func NewScops(s []string) *Scops {
	scopsSlice := slices.Clone(s)
	slices.Sort(scopsSlice)
	return &Scops{
		scopsSlice: scopsSlice,
	}
}
