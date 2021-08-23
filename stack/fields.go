package stack

// Fields 域存储接口
type Fields interface {
	// Merge 合并两个域存储
	Merge(Fields) Fields
	// Set 添加域
	Set(string, interface{}) Fields
	// ToKVs 转成kv map
	KVs() map[string]interface{}
	// KVSlice 转成kv slice
	KVsSlice() []interface{}
}

// fields 域存储
type fields map[string]interface{}

// New 域存储初始化
func New() Fields {
	return make(fields, 5)
}

func (fs fields) Merge(fs2 Fields) Fields {
	newFS := New()
	for k, v := range fs.KVs() {
		newFS.Set(k, v)
	}
	for k, v := range fs2.KVs() {
		newFS.Set(k, v)
	}
	return newFS
}

func (fs fields) Set(key string, val interface{}) Fields {
	fs[key] = val
	return fs
}

func (fs fields) KVs() map[string]interface{} {
	return fs
}

func (fs fields) KVsSlice() []interface{} {
	kvsSlice := make([]interface{}, 0, len(fs))
	for key, val := range fs {
		kvsSlice = append(kvsSlice, key, val)
	}
	return kvsSlice
}

// FromKVs 从kvs生成
func FromKVs(kvs map[string]interface{}) Fields {
	return fields(kvs)
}

type multi struct {
	others []Fields
	self   Fields
}

// Multi Multi
func Multi(others ...Fields) Fields {
	return &multi{
		others: others,
		self:   New(),
	}
}

func (fs multi) Merge(fs2 Fields) Fields {
	newFS := New()
	for k, v := range fs.KVs() {
		newFS.Set(k, v)
	}
	for k, v := range fs2.KVs() {
		newFS.Set(k, v)
	}
	return newFS
}

func (fs multi) Set(key string, val interface{}) Fields {
	fs.self.Set(key, val)
	return fs
}

func (fs multi) KVs() map[string]interface{} {
	kvs := make(map[string]interface{})
	for _, other := range fs.others {
		for key, val := range other.KVs() {
			kvs[key] = val
		}
	}
	for key, val := range fs.self.KVs() {
		kvs[key] = val
	}
	return kvs
}

func (fs multi) KVsSlice() []interface{} {
	kvsSlice := make([]interface{}, 0, 20)
	for _, other := range fs.others {
		for key, val := range other.KVs() {
			kvsSlice = append(kvsSlice, key, val)
		}
	}
	for key, val := range fs.self.KVs() {
		kvsSlice = append(kvsSlice, key, val)
	}
	return kvsSlice
}
