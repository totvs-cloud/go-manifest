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
