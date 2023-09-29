package manifest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (l *list) Delete(ctx context.Context) error {
	for _, v := range l.Resources() {
		if err := l.delete(ctx, v); err != nil {
			return err
		}
	}

	return nil
}

func (l *list) delete(ctx context.Context, obj *unstructured.Unstructured) error {
	log := logr.FromContextOrDiscard(ctx)

	gvk := obj.GroupVersionKind()
	kind := fmt.Sprintf("%s.%s", strings.ToLower(gvk.Kind), gvk.Group)

	if len(gvk.Group) == 0 {
		kind = strings.ToLower(gvk.Kind)
	}

	mapper, err := l.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to retrieve REST mapping for %s: %w", kind, err)
	}

	resource := l.client.Resource(mapper.Resource).Namespace(obj.GetNamespace())
	if mapper.Scope.Name() == meta.RESTScopeNameRoot {
		resource = l.client.Resource(mapper.Resource)
	}

	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to delete %s %q: %w", kind, obj.GetName(), err)
	}

	log.Info(fmt.Sprintf("%s %q deleted", kind, obj.GetName()))

	return nil
}
