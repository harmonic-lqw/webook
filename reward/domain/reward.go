package domain

type Target struct {
	// 给什么打赏
	Biz   string
	BizId int64

	// 要打赏的东西名称
	BizName string

	// 打赏的目标用户
	TarUId int64
}

type Reward struct {
	Id     int64
	SrcUid int64
	Target Target

	Amt    int64
	Status RewardStatus
}

// Completed 是否已经完成，也就是是否处理了支付回调
func (r Reward) Completed() bool {
	return r.Status == RewardStatusFailed || r.Status == RewardStatusPayed
}

type RewardStatus uint8

func (r RewardStatus) AsUint8() uint8 {
	return uint8(r)
}

const (
	RewardStatusUnknown = iota
	RewardStatusInit
	RewardStatusPayed
	RewardStatusFailed
)

type CodeURL struct {
	Rid int64
	URL string
}
