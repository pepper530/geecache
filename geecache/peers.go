package geecache

// 根据传入的 key选择相应的节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 从相应的 group中查找缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
