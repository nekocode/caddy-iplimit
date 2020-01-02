package iplimit

import "container/list"

func (c *Config) addIP(ip string) bool {
	ipPool := c.IPPool

	count, ok := ipPool.Get(ip)
	if ok {
		if count.(int) > 0 {
			ipPool.Set(ip, count.(int)+1)
			return true
		}
		ipPool.Remove(ip)
	}

	// Check size of pool
	if ipPool.Count() < c.Max {
		ipPool.Set(ip, 1)
		return true
	}
	c.cleanPool()
	// Check again
	if ipPool.Count() < c.Max {
		ipPool.Set(ip, 1)
		return true
	}

	return false
}

func (c *Config) removeIP(ip string) {
	ipPool := c.IPPool

	count, ok := ipPool.Get(ip)
	if ok {
		ipPool.Set(ip, count.(int)-1)
	}
}

func (c *Config) cleanPool() {
	ipPool := c.IPPool

	inactiveIPs := list.New()
	for item := range ipPool.Iter() {
		if item.Val.(int) <= 0 {
			inactiveIPs.PushBack(item.Key)
		}
	}
	for i := inactiveIPs.Front(); i != nil; i = i.Next() {
		ipPool.Remove(i.Value.(string))
	}
}
