/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"text/template"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

func parseCIDR(s string) (*net.IPNet, error) {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	ipnet.IP = ip
	return ipnet, nil
}

func addIp(ip net.IP, add uint64) net.IP {
	if len(ip) < 8 {
		return ip
	}

	out := make(net.IP, len(ip))
	copy(out, ip)

	i := binary.BigEndian.Uint64(out[len(out)-8:])
	i += add

	binary.BigEndian.PutUint64(out[len(out)-8:], i)
	return out
}

type ipPool struct {
	mut    sync.Mutex
	used   map[string]struct{}
	usable map[string]struct{}
	cidr   *net.IPNet
	index  uint64
}

func newIPPool(cidr *net.IPNet) *ipPool {
	return &ipPool{
		used:   make(map[string]struct{}),
		usable: make(map[string]struct{}),
		cidr:   cidr,
	}
}

func (i *ipPool) new() string {
	for {
		ip := addIp(i.cidr.IP, i.index).String()
		i.index++

		if _, ok := i.used[ip]; ok {
			continue
		}

		i.used[ip] = struct{}{}
		i.usable[ip] = struct{}{}
		return ip
	}
}

func (i *ipPool) Get() string {
	i.mut.Lock()
	defer i.mut.Unlock()
	ip := ""
	if len(i.usable) != 0 {
		for s := range i.usable {
			ip = s
		}
	}
	if ip == "" {
		ip = i.new()
	}
	delete(i.usable, ip)
	i.used[ip] = struct{}{}
	return ip
}

func (i *ipPool) Put(ip string) {
	i.mut.Lock()
	defer i.mut.Unlock()
	if !i.cidr.Contains(net.ParseIP(ip)) {
		return
	}
	delete(i.used, ip)
	i.usable[ip] = struct{}{}
}

func (i *ipPool) Use(ip string) {
	i.mut.Lock()
	defer i.mut.Unlock()
	if !i.cidr.Contains(net.ParseIP(ip)) {
		return
	}
	i.used[ip] = struct{}{}
}

func toTemplateJson(text string, original interface{}, funcMap template.FuncMap) ([]byte, error) {
	text = strings.TrimSpace(text)
	temp, err := template.New("_").Funcs(funcMap).Parse(text)
	if err != nil {
		return nil, err
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)

	buf.Reset()
	err = json.NewEncoder(buf).Encode(original)
	if err != nil {
		return nil, err
	}

	var data interface{}
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	buf.Reset()
	err = temp.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	out, err := yaml.YAMLToJSON(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, buf.String())
	}
	return out, nil
}

var (
	templateCache = sync.Map{}
	bufferPool    = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

type parallelTasks struct {
	wg     sync.WaitGroup
	bucket chan struct{}
	tasks  chan func()
}

func newParallelTasks(n int) *parallelTasks {
	return &parallelTasks{
		bucket: make(chan struct{}, n),
		tasks:  make(chan func()),
	}
}

func (p *parallelTasks) Add(fun func()) {
	p.wg.Add(1)
	select {
	case p.tasks <- fun: // there are idle threads
	case p.bucket <- struct{}{}: // there are free threads
		go p.fork()
		p.tasks <- fun
	}
}

func (p *parallelTasks) fork() {
	defer func() {
		<-p.bucket
	}()
	timer := time.NewTimer(time.Second / 2)
	for {
		select {
		case <-timer.C: // idle threads
			return
		case fun := <-p.tasks:
			fun()
			p.wg.Done()
			timer.Reset(time.Second / 2)
		}
	}
}

func (p *parallelTasks) Wait() {
	p.wg.Wait()
}

type stringSets struct {
	mut  sync.RWMutex
	sets map[string]struct{}
}

func newStringSets() *stringSets {
	return &stringSets{
		sets: make(map[string]struct{}),
	}
}

func (s *stringSets) Size() int {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return len(s.sets)
}

func (s *stringSets) Put(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.sets[key] = struct{}{}
}

func (s *stringSets) Delete(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.sets, key)
}

func (s *stringSets) Has(key string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	_, ok := s.sets[key]
	return ok
}

func (s *stringSets) Foreach(f func(string)) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	for k := range s.sets {
		f(k)
	}
}

func labelsParse(selector string) (labels.Selector, error) {
	if selector == "" {
		return nil, nil
	}
	return labels.Parse(selector)
}

type SafeCache struct {
	mu    sync.RWMutex
	cache map[string]any
}

func InitSafeCache() *SafeCache {
	return &SafeCache{
		cache: make(map[string]any),
		mu:    sync.RWMutex{},
	}
}

func (s *SafeCache) SetCache(key string, value any) {
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
}

func (s *SafeCache) UnsetCache(key string) {
	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()
}

func (s *SafeCache) GetCache(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if value, ok := s.cache[key]; ok {
		return value
	}
	return nil
}
