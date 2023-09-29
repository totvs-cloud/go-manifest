package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Filter func(u *unstructured.Unstructured) bool

func (l *list) Filter(funcs ...Filter) List {
	resources := make([]*unstructured.Unstructured, 0, l.Size())

	for _, v := range l.Resources() {
		resource := v.DeepCopy()
		for _, filter := range funcs {
			if filter(resource) {
				resources = append(resources, resource)
			}
		}
	}

	return &list{resources: resources, fieldManager: l.fieldManager, client: l.client, mapper: l.mapper}
}
