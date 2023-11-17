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

func All(filters ...Filter) Filter {
	return func(u *unstructured.Unstructured) bool {
		for _, filter := range filters {
			if !filter(u) {
				return false
			}
		}

		return true
	}
}

func Any(filters ...Filter) Filter {
	return func(u *unstructured.Unstructured) bool {
		for _, filter := range filters {
			if filter(u) {
				return true
			}
		}

		return false
	}
}

func Not(filter Filter) Filter {
	return func(u *unstructured.Unstructured) bool {
		return !filter(u)
	}
}

func ByAPIVersion(apiVersion string) Filter {
	return func(u *unstructured.Unstructured) bool {
		return u.GetAPIVersion() == apiVersion
	}
}

func ByKind(kind string) Filter {
	return func(u *unstructured.Unstructured) bool {
		return u.GetKind() == kind
	}
}

func ByAnnotation(annotation, value string) Filter {
	return func(u *unstructured.Unstructured) bool {
		v, ok := u.GetAnnotations()[annotation]
		if value == "" {
			return ok
		}

		return v == value
	}
}

func ByLabel(label, value string) Filter {
	return func(u *unstructured.Unstructured) bool {
		v, ok := u.GetLabels()[label]
		if value == "" {
			return ok
		}

		return v == value
	}
}

func ByLabels(labels map[string]string) Filter {
	return func(u *unstructured.Unstructured) bool {
		for key, value := range labels {
			if v := u.GetLabels()[key]; v == value {
				return true
			}
		}

		return false
	}
}
