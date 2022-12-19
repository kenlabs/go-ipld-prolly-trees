package tree

import (
	"github.com/ipld/go-ipld-prime"
	"go-ipld-prolly-trees/pkg/tree/schema"
	"io"
	"sort"
)

type op int

type key []byte

func (k *key) Equal(other *key) bool {
	return DefaultCompareFunc(*k, *other) == 0
}

const (
	unknown op = 0
	modify  op = 1
	add     op = 2
	remove  op = 3
)

type Mutation struct {
	key []byte
	val ipld.Node
	op  op
}

type Mutations struct {
	mutations   []*Mutation
	compareFunc schema.CompareFunc
}

func NewMutations() *Mutations {
	return &Mutations{
		mutations:   make([]*Mutation, 0),
		compareFunc: DefaultCompareFunc,
	}
}

func (m *Mutations) keyMutation(item []byte) (int, *Mutation) {
	// KeyIndex finds the index that the closest but not smaller than the item
	length := len(m.mutations)
	l, r := 0, length-1
	for l < r {
		mid := (l + r) / 2
		midKey := m.mutations[mid].key
		if m.compareFunc(midKey, item) == 0 {
			return mid, m.mutations[mid]
		} else if m.compareFunc(midKey, item) > 0 {
			r = mid
		} else {
			l = mid + 1
		}
	}
	return r, m.mutations[r]
}

func (m *Mutations) addMutation(mut *Mutation) error {
	//if already exist, replace it
	if len(m.mutations) != 0 {
		idx, oldMut := m.keyMutation(mut.key)
		if m.compareFunc(oldMut.key, mut.key) == 0 {
			m.mutations[idx] = mut
			return nil
		}
	}

	m.mutations = append(m.mutations, mut)
	sort.Slice(m.mutations, func(i, j int) bool {
		if DefaultCompareFunc(m.mutations[i].key, m.mutations[j].key) < 0 {
			return true
		}
		return false
	})
	return nil
}

func (m *Mutations) NextMutation() (*Mutation, error) {
	if len(m.mutations) == 0 {
		return nil, io.EOF
	}
	mut := m.mutations[0]
	m.mutations = m.mutations[1:]

	return mut, nil
}

func (m *Mutations) Get(item []byte) (ipld.Node, error) {
	_, mut := m.keyMutation(item)
	if m.compareFunc(mut.key, item) == 0 {
		if mut.op != remove {
			return mut.val, nil
		}
	}
	return nil, nil
}

func (m *Mutations) Has(item []byte) (bool, error) {
	_, mut := m.keyMutation(item)
	if m.compareFunc(mut.key, item) == 0 {
		return true, nil
	}
	return false, nil
}