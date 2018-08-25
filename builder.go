package djson

type builder interface {
	newMapBuilder(key string) builder
	newArrayBuilder(index int) builder
	set(val interface{})
}

type rootBuilder struct {
	// Compose with the map builder since root is a map
	mapBuilder
}

func newRootBuilder() *rootBuilder {
	b := &rootBuilder{
		mapBuilder: mapBuilder{
			m:   map[string]interface{}{},
			key: "root",
		},
	}
	// Set parent to itself to intercept the final set call
	b.parent = b
	return b
}

func (b *rootBuilder) set(val interface{}) {
}

func (b *rootBuilder) get() map[string]interface{} {
	return b.m["root"].(map[string]interface{})
}

type mapBuilder struct {
	m      map[string]interface{}
	key    string
	parent builder
}

func (b *mapBuilder) newMapBuilder(key string) builder {
	if v, ok := b.m[b.key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return &mapBuilder{m: m, key: key, parent: b}
		}
	}
	m := map[string]interface{}{}
	return &mapBuilder{m: m, key: key, parent: b}
}

func (b *mapBuilder) newArrayBuilder(index int) builder {
	if v, ok := b.m[b.key]; ok {
		if a, ok := v.([]interface{}); ok {
			return &arrayBuilder{a: a, index: index, parent: b}
		}
	}
	a := []interface{}{}
	return &arrayBuilder{a: a, index: index, parent: b}
}

func (b *mapBuilder) set(val interface{}) {
	b.m[b.key] = val
	b.parent.set(b.m)
}

type arrayBuilder struct {
	a      []interface{}
	index  int
	parent builder
}

func (b *arrayBuilder) newMapBuilder(key string) builder {
	if len(b.a) >= b.index+1 {
		if m, ok := b.a[b.index].(map[string]interface{}); ok {
			return &mapBuilder{m: m, key: key, parent: b}
		}
	}
	m := map[string]interface{}{}
	return &mapBuilder{m: m, key: key, parent: b}
}

func (b *arrayBuilder) newArrayBuilder(index int) builder {
	if len(b.a) >= b.index+1 {
		if a, ok := b.a[b.index].([]interface{}); ok {
			return &arrayBuilder{a: a, index: index, parent: b}
		}
	}
	var a []interface{}
	return &arrayBuilder{a: a, index: index, parent: b}
}

func (b *arrayBuilder) set(val interface{}) {
	if len(b.a) < b.index+1 {
		add := b.index + 1 - len(b.a)
		b.a = append(b.a, make([]interface{}, add)...)
	}
	b.a[b.index] = val
	b.parent.set(b.a)
}
