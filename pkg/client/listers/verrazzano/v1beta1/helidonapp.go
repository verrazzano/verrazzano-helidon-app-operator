// Copyright (c) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/verrazzano/verrazzano-helidon-app-operator/pkg/apis/verrazzano/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// HelidonAppLister helps list HelidonApps.
type HelidonAppLister interface {
	// List lists all HelidonApps in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.HelidonApp, err error)
	// HelidonApps returns an object that can list and get HelidonApps.
	HelidonApps(namespace string) HelidonAppNamespaceLister
	HelidonAppListerExpansion
}

// helidonAppLister implements the HelidonAppLister interface.
type helidonAppLister struct {
	indexer cache.Indexer
}

// NewHelidonAppLister returns a new HelidonAppLister.
func NewHelidonAppLister(indexer cache.Indexer) HelidonAppLister {
	return &helidonAppLister{indexer: indexer}
}

// List lists all HelidonApps in the indexer.
func (s *helidonAppLister) List(selector labels.Selector) (ret []*v1beta1.HelidonApp, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.HelidonApp))
	})
	return ret, err
}

// HelidonApps returns an object that can list and get HelidonApps.
func (s *helidonAppLister) HelidonApps(namespace string) HelidonAppNamespaceLister {
	return helidonAppNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// HelidonAppNamespaceLister helps list and get HelidonApps.
type HelidonAppNamespaceLister interface {
	// List lists all HelidonApps in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta1.HelidonApp, err error)
	// Get retrieves the HelidonApp from the indexer for a given namespace and name.
	Get(name string) (*v1beta1.HelidonApp, error)
	HelidonAppNamespaceListerExpansion
}

// helidonAppNamespaceLister implements the HelidonAppNamespaceLister
// interface.
type helidonAppNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all HelidonApps in the indexer for a given namespace.
func (s helidonAppNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.HelidonApp, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.HelidonApp))
	})
	return ret, err
}

// Get retrieves the HelidonApp from the indexer for a given namespace and name.
func (s helidonAppNamespaceLister) Get(name string) (*v1beta1.HelidonApp, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("helidonapp"), name)
	}
	return obj.(*v1beta1.HelidonApp), nil
}
