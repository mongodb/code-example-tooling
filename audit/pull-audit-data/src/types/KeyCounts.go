package types

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
