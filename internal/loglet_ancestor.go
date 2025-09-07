package internal

// logletAncestorInfo holds information about a cached ancestor
type logletAncestorInfo struct {
	baseMap     map[string]any
	pathToRoot  []*Loglet
	totalFields int
}

// canDelegate returns whether we can delegate to the cached ancestor.
// Delegation is possible when:
// 1. A baseMap exists (cached ancestor found)
// 2. A valid pathToRoot exists (not empty)
// 3. AND either:
//   - pathToRoot has length 1 AND current loglet has no fields (self delegation)
//   - totalFields equals baseMap size (no additional fields to add)
func (info *logletAncestorInfo) canDelegate() bool {
	if info.baseMap == nil || len(info.pathToRoot) == 0 {
		return false
	}
	// Can delegate if this loglet itself has the cache AND no fields,
	// or if no fields need to be added beyond the cached ancestor
	if len(info.pathToRoot) == 1 {
		// Self delegation only if current loglet has no fields
		current := info.pathToRoot[0]
		return len(current.keys) == 0
	}
	return info.totalFields == len(info.baseMap)
}

// getBaseMap returns the base map, ensuring ancestors with fields get cached
func (info *logletAncestorInfo) getBaseMap() map[string]any {
	if info.baseMap != nil {
		return info.baseMap
	}

	// Try to cache ancestor for delegation
	if info.shouldTryAncestorCaching() {
		info.tryAncestorCaching()
	}
	return info.baseMap
}

// tryAncestorCaching attempts to cache ancestors with fields for delegation
func (info *logletAncestorInfo) tryAncestorCaching() {
	for _, ancestor := range info.pathToRoot[1:] {
		if fieldsMap, shouldUpdate := ancestor.cacheIfNeeded(); shouldUpdate {
			info.baseMap = fieldsMap
			info.totalFields = len(fieldsMap)
			return
		}
	}
}

// shouldTryAncestorCaching checks if we should attempt ancestor caching
func (info *logletAncestorInfo) shouldTryAncestorCaching() bool {
	return len(info.pathToRoot) > 1 && len(info.pathToRoot[0].keys) == 0
}

// findCachedAncestor walks up the tree to find a cached ancestor
func (ll *Loglet) findCachedAncestor() logletAncestorInfo {
	var pathToRoot []*Loglet
	current := ll
	totalFields := 0

	for current != nil {
		// Only lock ancestor nodes (current != ll) since we already hold our own lock
		needsLock := current != ll
		if baseMap, hasCached := current.peekFieldsMap(needsLock); hasCached {
			// Found cached ancestor - don't add it to pathToRoot
			adjustedFields := len(baseMap) + totalFields
			return logletAncestorInfo{
				baseMap:     baseMap,
				pathToRoot:  pathToRoot,
				totalFields: adjustedFields,
			}
		}

		// Only add uncached nodes to pathToRoot
		pathToRoot = append(pathToRoot, current)
		totalFields += len(current.keys)
		current = current.GetParent()
	}

	return logletAncestorInfo{
		baseMap:     nil,
		pathToRoot:  pathToRoot,
		totalFields: totalFields,
	}
}
