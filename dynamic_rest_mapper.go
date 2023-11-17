package manifest

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type dynamicRESTMapper struct {
	mapper      meta.RESTMapper
	client      *discovery.DiscoveryClient
	knownGroups map[string]*restmapper.APIGroupResources
	apiGroups   map[string]*metav1.APIGroup

	// thread-safe reloading
	mu sync.RWMutex
}

func newDynamicRESTMapper(cfg *rest.Config, httpClient *http.Client) (*dynamicRESTMapper, error) {
	client, err := discovery.NewDiscoveryClientForConfigAndClient(cfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discovery client: %w", err)
	}

	return &dynamicRESTMapper{
		mapper:      restmapper.NewDiscoveryRESTMapper([]*restmapper.APIGroupResources{}),
		client:      client,
		knownGroups: map[string]*restmapper.APIGroupResources{},
		apiGroups:   map[string]*metav1.APIGroup{},
	}, nil
}

func (m *dynamicRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	res, err := m.getMapper().KindFor(resource)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(resource.Group, resource.Version); err != nil {
			return schema.GroupVersionKind{}, fmt.Errorf("failed to retrieve kind for GroupVersionResource %q: %w", resource, err)
		}

		if res, err = m.getMapper().KindFor(resource); err != nil {
			return schema.GroupVersionKind{}, fmt.Errorf("failed to retrieve kind for GroupVersionResource %q: %w", resource, err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	res, err := m.getMapper().KindsFor(resource)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(resource.Group, resource.Version); err != nil {
			return nil, fmt.Errorf("failed to retrieve kinds for GroupVersionResource %q: %w", resource, err)
		}

		if res, err = m.getMapper().KindsFor(resource); err != nil {
			return nil, fmt.Errorf("failed to retrieve kinds for GroupVersionResource %q: %w", resource, err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	res, err := m.getMapper().ResourceFor(input)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(input.Group, input.Version); err != nil {
			return schema.GroupVersionResource{}, err
		}

		if res, err = m.getMapper().ResourceFor(input); err != nil {
			return schema.GroupVersionResource{}, fmt.Errorf("failed to retrieve resource for GroupVersionResource %q: %w", input, err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	res, err := m.getMapper().ResourcesFor(input)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(input.Group, input.Version); err != nil {
			return nil, fmt.Errorf("failed to retrieve resources for GroupVersionResource %q: %w", input, err)
		}

		if res, err = m.getMapper().ResourcesFor(input); err != nil {
			return nil, fmt.Errorf("failed to retrieve resources for GroupVersionResource %q: %w", input, err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	res, err := m.getMapper().RESTMapping(gk, versions...)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(gk.Group, versions...); err != nil {
			return nil, fmt.Errorf("failed to retrieve REST mapping for GroupKind %q and versions %s: %w", gk, strings.Join(versions, ","), err)
		}

		res, err = m.getMapper().RESTMapping(gk, versions...)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve REST mapping for GroupKind %q and versions %s: %w", gk, strings.Join(versions, ","), err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	res, err := m.getMapper().RESTMappings(gk, versions...)
	if meta.IsNoMatchError(err) {
		if err = m.addKnownGroupAndReload(gk.Group, versions...); err != nil {
			return nil, fmt.Errorf("failed to retrieve REST mappings for GroupKind %q and versions %s: %w", gk, strings.Join(versions, ","), err)
		}

		if res, err = m.getMapper().RESTMappings(gk, versions...); err != nil {
			return nil, fmt.Errorf("failed to retrieve REST mappings for GroupKind %q and versions %s: %w", gk, strings.Join(versions, ","), err)
		}
	}

	return res, nil
}

func (m *dynamicRESTMapper) ResourceSingularizer(resource string) (string, error) {
	singular, err := m.getMapper().ResourceSingularizer(resource)
	if err != nil {
		return "", fmt.Errorf("failed to singularize resource: %w", err)
	}

	return singular, nil
}

func (m *dynamicRESTMapper) getMapper() meta.RESTMapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mapper
}

func (m *dynamicRESTMapper) addKnownGroupAndReload(groupName string, versions ...string) error {
	if len(versions) == 1 && versions[0] == "" {
		versions = nil
	}

	if len(versions) == 0 {
		apiGroup, err := m.findAPIGroupByName(groupName)
		if err != nil {
			return err
		}

		for _, version := range apiGroup.Versions {
			versions = append(versions, version.Version)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	groupResources := &restmapper.APIGroupResources{
		Group:              metav1.APIGroup{Name: groupName},
		VersionedResources: make(map[string][]metav1.APIResource),
	}
	if _, ok := m.knownGroups[groupName]; ok {
		groupResources = m.knownGroups[groupName]
	}

	groupVersionResources, err := m.fetchGroupVersionResources(groupName, versions...)
	if err != nil {
		return err
	}

	for version, resources := range groupVersionResources {
		groupResources.VersionedResources[version.Version] = resources.APIResources
	}

	for _, version := range versions {
		found := false

		for _, v := range groupResources.Group.Versions {
			if v.Version == version {
				found = true
				break
			}
		}

		if !found {
			groupResources.Group.Versions = append(groupResources.Group.Versions, metav1.GroupVersionForDiscovery{
				GroupVersion: metav1.GroupVersion{Group: groupName, Version: version}.String(),
				Version:      version,
			})
		}
	}

	m.knownGroups[groupName] = groupResources
	updatedGroupResources := make([]*restmapper.APIGroupResources, 0, len(m.knownGroups))

	for _, agr := range m.knownGroups {
		updatedGroupResources = append(updatedGroupResources, agr)
	}

	m.mapper = restmapper.NewDiscoveryRESTMapper(updatedGroupResources)

	return nil
}

func (m *dynamicRESTMapper) findAPIGroupByName(groupName string) (*metav1.APIGroup, error) {
	{
		m.mu.RLock()
		group, ok := m.apiGroups[groupName]
		m.mu.RUnlock()

		if ok {
			return group, nil
		}
	}

	apiGroups, err := m.client.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to find API group with name %q: %w", groupName, err)
	}

	if len(apiGroups.Groups) == 0 {
		return nil, fmt.Errorf("unable to find API group with name %q", groupName)
	}

	m.mu.Lock()
	for i := range apiGroups.Groups {
		group := &apiGroups.Groups[i]
		m.apiGroups[group.Name] = group
	}
	m.mu.Unlock()
	{
		m.mu.RLock()
		group, ok := m.apiGroups[groupName]
		m.mu.RUnlock()

		if ok {
			return group, nil
		}
	}

	return nil, nil
}

func (m *dynamicRESTMapper) fetchGroupVersionResources(groupName string, versions ...string) (map[schema.GroupVersion]*metav1.APIResourceList, error) {
	groupVersionResources := make(map[schema.GroupVersion]*metav1.APIResourceList)
	failedGroups := make(map[schema.GroupVersion]error)

	for _, version := range versions {
		groupVersion := schema.GroupVersion{Group: groupName, Version: version}

		apiResourceList, err := m.client.ServerResourcesForGroupVersion(groupVersion.String())
		if err != nil {
			failedGroups[groupVersion] = err
		}

		if apiResourceList != nil {
			groupVersionResources[groupVersion] = apiResourceList
		}
	}

	if len(failedGroups) > 0 {
		subErrors := make([]string, 0, len(failedGroups))
		for k, v := range failedGroups {
			subErrors = append(subErrors, fmt.Sprintf("%s: %v", k, v))
		}

		sort.Strings(subErrors)

		return nil, fmt.Errorf("unable to retrieve the complete list of server APIs: %s", strings.Join(subErrors, ", "))
	}

	return groupVersionResources, nil
}
