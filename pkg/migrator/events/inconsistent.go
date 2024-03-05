package events

type InconsistentEvent struct {
	ID int64
	// 取值为 SRC 表示以源表为准，取值为 DST 表示以目标表为准
	Direction string
	// 标记什么原因引起的不一致
	Type string
}

const (
	// InconsistentEventTypeBaseMissing 校验的源数据，缺失
	InconsistentEventTypeBaseMissing = "base_missing"
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺失
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNotEqual 不相等
	InconsistentEventTypeNotEqual = "not_equal"
)
