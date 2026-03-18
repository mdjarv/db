package schema

import (
	"context"
	"sync"
)

// CachedInspector wraps an Inspector and caches results per schema+table key.
// Thread-safe via sync.RWMutex.
type CachedInspector struct {
	inner Inspector
	mu    sync.RWMutex

	tables      map[string][]Table      // key: schema
	columns     map[string][]ColumnInfo // key: schema.table
	indexes     map[string][]Index      // key: schema.table
	constraints map[string][]Constraint // key: schema.table
	foreignKeys map[string][]ForeignKey // key: schema.table
}

// NewCachedInspector returns a caching wrapper around the given Inspector.
func NewCachedInspector(inner Inspector) *CachedInspector {
	return &CachedInspector{
		inner:       inner,
		tables:      make(map[string][]Table),
		columns:     make(map[string][]ColumnInfo),
		indexes:     make(map[string][]Index),
		constraints: make(map[string][]Constraint),
		foreignKeys: make(map[string][]ForeignKey),
	}
}

func tableKey(schema, table string) string {
	return schema + "." + table
}

// Tables returns cached table list, querying the inner Inspector on cache miss.
func (c *CachedInspector) Tables(ctx context.Context, schema string) ([]Table, error) {
	c.mu.RLock()
	if cached, ok := c.tables[schema]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	result, err := c.inner.Tables(ctx, schema)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.tables[schema] = result
	c.mu.Unlock()
	return result, nil
}

// Columns returns cached column info, querying the inner Inspector on cache miss.
func (c *CachedInspector) Columns(ctx context.Context, schema, table string) ([]ColumnInfo, error) {
	key := tableKey(schema, table)

	c.mu.RLock()
	if cached, ok := c.columns[key]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	result, err := c.inner.Columns(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.columns[key] = result
	c.mu.Unlock()
	return result, nil
}

// Indexes returns cached index info, querying the inner Inspector on cache miss.
func (c *CachedInspector) Indexes(ctx context.Context, schema, table string) ([]Index, error) {
	key := tableKey(schema, table)

	c.mu.RLock()
	if cached, ok := c.indexes[key]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	result, err := c.inner.Indexes(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.indexes[key] = result
	c.mu.Unlock()
	return result, nil
}

// Constraints returns cached constraint info, querying the inner Inspector on cache miss.
func (c *CachedInspector) Constraints(ctx context.Context, schema, table string) ([]Constraint, error) {
	key := tableKey(schema, table)

	c.mu.RLock()
	if cached, ok := c.constraints[key]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	result, err := c.inner.Constraints(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.constraints[key] = result
	c.mu.Unlock()
	return result, nil
}

// ForeignKeys returns cached FK info, querying the inner Inspector on cache miss.
func (c *CachedInspector) ForeignKeys(ctx context.Context, schema, table string) ([]ForeignKey, error) {
	key := tableKey(schema, table)

	c.mu.RLock()
	if cached, ok := c.foreignKeys[key]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	result, err := c.inner.ForeignKeys(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.foreignKeys[key] = result
	c.mu.Unlock()
	return result, nil
}

// Invalidate clears all cached data, forcing re-queries on next access.
func (c *CachedInspector) Invalidate() {
	c.mu.Lock()
	c.tables = make(map[string][]Table)
	c.columns = make(map[string][]ColumnInfo)
	c.indexes = make(map[string][]Index)
	c.constraints = make(map[string][]Constraint)
	c.foreignKeys = make(map[string][]ForeignKey)
	c.mu.Unlock()
}

// InvalidateTable clears cached data for a specific table.
func (c *CachedInspector) InvalidateTable(schema, table string) {
	key := tableKey(schema, table)
	c.mu.Lock()
	delete(c.columns, key)
	delete(c.indexes, key)
	delete(c.constraints, key)
	delete(c.foreignKeys, key)
	// also invalidate table list for this schema since row counts may change
	delete(c.tables, schema)
	c.mu.Unlock()
}
