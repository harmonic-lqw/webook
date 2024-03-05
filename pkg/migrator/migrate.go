package migrator

type Entity interface {
	// ID 返回 ID，数据迁移完全围绕 ID 来进行
	ID() int64
	// CompareTo 交给具体的 Entity 实现者来负责
	CompareTo(dst Entity) bool
}
