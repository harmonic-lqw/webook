package migrator

type Entity interface {
	// ID 返回 ID，数据迁移完全围绕 ID 来进行
	// 实体的 id 就是事件的 id 就是 dao的id
	ID() int64
	// CompareTo 交给具体的 Entity 实现者来负责
	CompareTo(dst Entity) bool
}
