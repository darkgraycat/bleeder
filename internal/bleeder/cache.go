package bleeder

// A two layered generic cache
type Cache[T any] struct {
	data map[string]map[string]T
}

// Create new Cache instance
func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		data: make(map[string]map[string]T),
	}
}

// Store new value into cache
func (c *Cache[T]) Set(name, key string, value T) {
	if _, ok := c.data[name]; !ok {
		c.data[name] = make(map[string]T)
	}
	c.data[name][key] = value
}

// Get value from cache
func (c *Cache[T]) Get(name, key string) T {
	var none T
	if set, ok := c.data[name]; ok {
		if value, ok := set[key]; ok {
			return value
		}
	}
	return none
}

// Remove from cache by name:key
func (c *Cache[T]) DelKey(name, key string) {
	if set, ok := c.data[name]; ok {
		delete(set, key)
	}
}

// Remove from cache by name
func (c *Cache[T]) DelName(name string) {
	delete(c.data, name)
}
