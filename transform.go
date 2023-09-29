package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Transformer func(u *unstructured.Unstructured) error

func (l *list) Transform(funcs ...Transformer) (List, error) {
	resources := make([]*unstructured.Unstructured, 0, l.Size())

	for _, v := range l.Resources() {
		resource := v.DeepCopy()
		for _, transform := range funcs {
			if err := transform(resource); err != nil {
				return &list{}, err
			}
		}

		resources = append(resources, resource)
	}

	return &list{resources: resources, fieldManager: l.fieldManager, client: l.client, mapper: l.mapper}, nil
}
