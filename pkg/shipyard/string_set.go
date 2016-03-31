package shipyard

//StringSet a set of strings
type StringSet struct {
	set map[string]bool
}

//NewStringSet create a new string set
func NewStringSet() *StringSet {
	set := &StringSet{
		set: make(map[string]bool),
	}

	return set
}

//Add add to the set.  Return true if it was new, false if it already existed
func (set *StringSet) Add(value string) bool {
	_, found := set.set[value]
	set.set[value] = true
	return !found //False if it existed already
}

//AsSlice get the set as a slice
func (set *StringSet) AsSlice() *[]string {
	keys := make([]string, len(set.set))

	i := 0
	for k := range set.set {
		keys[i] = k
		i++
	}

	return &keys
}
