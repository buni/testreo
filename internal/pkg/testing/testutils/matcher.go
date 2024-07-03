package testutils

import (
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

var _ gomock.Matcher = (*Matcher[any])(nil)

type Matcher[T any] struct {
	expected T
	opts     []cmp.Option
}

func (m *Matcher[T]) Matches(x any) bool {
	diff := cmp.Diff(m.expected, x, m.opts...)
	if diff != "" {
		log.Println(diff)
		return false
	}

	return true
}

func (m *Matcher[T]) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}

func NewMatcher[T any](expected T, opts ...cmp.Option) *Matcher[T] {
	return &Matcher[T]{expected: expected, opts: opts}
}
