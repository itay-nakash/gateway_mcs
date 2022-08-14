package multicluster_gw

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

// TODO: how to declare the set? should it be public varible, or part if the Set struct?
// can also use map[types.NamespacedName]struct{}:
type Set struct {
	Elements map[string]struct{}
}

func NewSiSet() *Set {
	var set Set
	set.Elements = make(map[string]struct{})
	return &set
}

func (s *Set) Add(elem string) {
	s.Elements[elem] = struct{}{}
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
