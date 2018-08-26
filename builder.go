package djson

type mapBuilderFactory interface {
	newMapBuilder(key string) builder
}

type arrayBuilderFactory interface {
	newArrayBuilder(index int) builder
}

type setter interface {
	set(val interface{})
}

type builder interface {
	mapBuilderFactory
	arrayBuilderFactory
	setter
}

type rootBuilder struct {
	m map[string]interface{}
}

func newRootBuilder(m map[string]interface{}) *rootBuilder {
	return &rootBuilder{
		m: m,
	}
}

func (b *rootBuilder) newMapBuilder(key string) builder {
	return &mapBuilder{m: b.m, key: key, parent: b}
}

func (b *rootBuilder) set(val interface{}) {
	// Set nothing, map is passed by reference.
}

type mapBuilder struct {
	m      map[string]interface{}
	key    string
	parent setter
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
	parent setter
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
