package multicluster_gw

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

// TODO: how to declare the set? should it be public varible, or part if the Set struct?
// can also use map[types.NamespacedName]struct{}:

type void struct{}

var member void

type Set struct {
	Elements map[string]void
}

func NewSiSet() *Set {
	var set Set
	set.Elements = make(map[string]void)
	return &set
}

func (s *Set) Add(elem string) {
	s.Elements[elem] = member
}

func (s *Set) Delete(elem string) error {
	if _, exists := s.Elements[elem]; !exists {
		// #TODO check about the error:
		return errors.NewBadRequest("Service Import is not present in set")
	}
	delete(s.Elements, elem)
	return nil
}

func (s *Set) Contains(elem string) bool {
	_, exists := s.Elements[elem]
	return exists
}

func (s *Set) GetSize() int {
	return len(s.Elements)
}
