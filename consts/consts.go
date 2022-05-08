package consts

const (
	DelayBucket    = "DelayBucket"
	ReservedBucket = "ReservedBucket"
	FailedBucket   = "FailedBucket"
	FinishBucket   = "FinishBucket"
)

const (
	State_Delay = iota
	State_Ready
	State_Reserved
	State_Deleted
)
