package types

// KeyCount and TwoLevelNestedKeyCount are used to keep track of keys and counts when sorting by count to print
// tables to console.
type KeyCount struct {
	Key   string
	Count int
}

type TwoLevelNestedKeyCount struct {
	TopLevelKey             string
	NestedMapKey            string
	SecondLevelNestedMapKey string
	Count                   int
}
