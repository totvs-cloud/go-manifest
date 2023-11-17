package manifest

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type List interface {
	Delete(ctx context.Context) error
	Apply(ctx context.Context) error
	Filter(funcs ...Filter) List
	Transform(funcs ...Transformer) (List, error)
	Resources() []*unstructured.Unstructured
	Size() int
	Append(mfs ...List) List
}

func EmptyList() List {
	return new(empty)
}

type empty struct{}

func (e *empty) Delete(ctx context.Context) error {
	return nil
}

func (e *empty) Apply(ctx context.Context) error {
	return nil
}

func (e *empty) Filter(funcs ...Filter) List {
	return e
}

func (e *empty) Transform(funcs ...Transformer) (List, error) {
	return e, nil
}

func (e *empty) Resources() []*unstructured.Unstructured {
	return nil
}

func (e *empty) Size() int {
	return 0
}

func (e *empty) Append(mfs ...List) List {
	resources := make([]*unstructured.Unstructured, 0)

	for _, mf := range mfs {
		for _, v := range mf.Resources() {
			resource := v.DeepCopy()
			resources = append(resources, resource)
		}
	}

	var (
		fieldManager string
		client       dynamic.Interface
		mapper       meta.RESTMapper
	)

	for _, mf := range mfs {
		if l, ok := mf.(*list); ok {
			fieldManager = l.fieldManager
			client = l.client
			mapper = l.mapper

			break
		}
	}

	return &list{resources: resources, fieldManager: fieldManager, client: client, mapper: mapper}
}

type list struct {
	resources    []*unstructured.Unstructured
	fieldManager string
	client       dynamic.Interface
	mapper       meta.RESTMapper
}

func (l *list) Resources() []*unstructured.Unstructured {
	return l.resources
}

func (l *list) Size() int {
	return len(l.resources)
}

func (l *list) Append(mfs ...List) List {
	resources := make([]*unstructured.Unstructured, 0, l.Size())

	for _, v := range l.Resources() {
		resource := v.DeepCopy()
		resources = append(resources, resource)
	}

	for _, mf := range mfs {
		for _, v := range mf.Resources() {
			resource := v.DeepCopy()
			resources = append(resources, resource)
		}
	}

	return &list{resources: resources, fieldManager: l.fieldManager, client: l.client, mapper: l.mapper}
}
