package requestdecoder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/buni/wallet/internal/pkg/syncx"
	"github.com/ggicci/httpin"
	"github.com/ggicci/httpin/core"
)

var (
	ErrProvidedTypeShouldBeAStruct    = errors.New("provided type should be a struct")
	ErrFoundAFieldWithoutMandatoryTag = errors.New("found a field without mandatory tag")
	ErrEncoderEngineError             = errors.New("encoder engine error")
	ErrContextValueIsEmpty            = errors.New("context value is empty")
	ErrContextValueCannotSetTarget    = errors.New("context value cannot be set")
	ErrValueFromContextCannotBeNil    = errors.New("value from context cannot be nil")
	ErrRequestCannotBeNil             = errors.New("request cannot be nil")
)

type cacheValue struct {
	hasFieldTags bool
	err          error
}

var typeCache *syncx.SyncMap[reflect.Type, cacheValue] = &syncx.SyncMap[reflect.Type, cacheValue]{} //nolint

// Decode ...
// Passing a nil req *http.Request will result in a panic.
func Decode[T any](req *http.Request, s *T) (*T, error) {
	var err error
	if req == nil {
		return nil, ErrRequestCannotBeNil
	}

	if s == nil {
		s = new(T)
	}
	typeT := reflect.TypeFor[T]()

	typeValue, ok := typeCache.Load(typeT)
	if !ok {
		hasFieldTags, err := enforceStructTag(s, "json")
		typeValue = cacheValue{
			hasFieldTags: hasFieldTags,
			err:          err,
		}
		typeCache.Store(typeT, typeValue)
		if err != nil {
			return nil, fmt.Errorf("failed to enforce struct tag: %w", err)
		}
	}

	if typeValue.err != nil {
		return nil, fmt.Errorf("failed to enforce struct tag: %w", err)
	}

	engine, err := httpin.New(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create httpin engine: %w", err)
	}

	input, err := engine.Decode(req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}

	inputT, ok := input.(*T)
	if !ok {
		return nil, ErrEncoderEngineError
	}

	*s = *inputT

	if !typeValue.hasFieldTags {
		return s, nil
	}

	err = json.NewDecoder(req.Body).Decode(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}

	return s, nil
}

func enforceStructTag(s any, tag string) (hasFieldTags bool, err error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false, ErrProvidedTypeShouldBeAStruct
	}

	for i := range v.NumField() {
		tagValue := v.Type().Field(i).Tag.Get(tag)
		if tagValue == "" {
			return false, fmt.Errorf("missing tag %s on field %s: %w", tag, v.Type().Field(i).Name, ErrFoundAFieldWithoutMandatoryTag)
		}

		if tagValue != "-" && !hasFieldTags {
			hasFieldTags = true
		}
	}

	return hasFieldTags, nil
}

type DirectiveFromContext[T any] struct {
	FromContext func(context.Context) *T
}

// Decode decodes the value from the request context and sets it to the target value.
func (f *DirectiveFromContext[T]) Decode(rtm *core.DirectiveRuntime) error { //nolint:revive
	ctx := rtm.GetRequest().Context()
	val := f.FromContext(ctx)
	if val == nil {
		return ErrValueFromContextCannotBeNil
	}

	if rtm.Value.Elem().Kind() != reflect.TypeFor[T]().Kind() {
		return ErrContextValueCannotSetTarget
	}

	rtm.Value.Elem().Set(reflect.ValueOf(*val))

	if rtm.Value.Elem().IsZero() {
		return ErrContextValueIsEmpty
	}

	return nil
}

// Encode is noop.
func (f *DirectiveFromContext[T]) Encode(rtm *core.DirectiveRuntime) error { //nolint:revive
	return nil
}
