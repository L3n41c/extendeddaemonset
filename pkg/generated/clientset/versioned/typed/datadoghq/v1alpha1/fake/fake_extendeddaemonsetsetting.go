// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/datadog/extendeddaemonset/pkg/apis/datadoghq/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeExtendedDaemonsetSettings implements ExtendedDaemonsetSettingInterface
type FakeExtendedDaemonsetSettings struct {
	Fake *FakeDatadoghqV1alpha1
	ns   string
}

var extendeddaemonsetsettingsResource = schema.GroupVersionResource{Group: "datadoghq.com", Version: "v1alpha1", Resource: "extendeddaemonsetsettings"}

var extendeddaemonsetsettingsKind = schema.GroupVersionKind{Group: "datadoghq.com", Version: "v1alpha1", Kind: "ExtendedDaemonsetSetting"}

// Get takes name of the extendedDaemonsetSetting, and returns the corresponding extendedDaemonsetSetting object, and an error if there is any.
func (c *FakeExtendedDaemonsetSettings) Get(name string, options v1.GetOptions) (result *v1alpha1.ExtendedDaemonsetSetting, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(extendeddaemonsetsettingsResource, c.ns, name), &v1alpha1.ExtendedDaemonsetSetting{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ExtendedDaemonsetSetting), err
}

// List takes label and field selectors, and returns the list of ExtendedDaemonsetSettings that match those selectors.
func (c *FakeExtendedDaemonsetSettings) List(opts v1.ListOptions) (result *v1alpha1.ExtendedDaemonsetSettingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(extendeddaemonsetsettingsResource, extendeddaemonsetsettingsKind, c.ns, opts), &v1alpha1.ExtendedDaemonsetSettingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ExtendedDaemonsetSettingList{ListMeta: obj.(*v1alpha1.ExtendedDaemonsetSettingList).ListMeta}
	for _, item := range obj.(*v1alpha1.ExtendedDaemonsetSettingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested extendedDaemonsetSettings.
func (c *FakeExtendedDaemonsetSettings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(extendeddaemonsetsettingsResource, c.ns, opts))

}

// Create takes the representation of a extendedDaemonsetSetting and creates it.  Returns the server's representation of the extendedDaemonsetSetting, and an error, if there is any.
func (c *FakeExtendedDaemonsetSettings) Create(extendedDaemonsetSetting *v1alpha1.ExtendedDaemonsetSetting) (result *v1alpha1.ExtendedDaemonsetSetting, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(extendeddaemonsetsettingsResource, c.ns, extendedDaemonsetSetting), &v1alpha1.ExtendedDaemonsetSetting{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ExtendedDaemonsetSetting), err
}

// Update takes the representation of a extendedDaemonsetSetting and updates it. Returns the server's representation of the extendedDaemonsetSetting, and an error, if there is any.
func (c *FakeExtendedDaemonsetSettings) Update(extendedDaemonsetSetting *v1alpha1.ExtendedDaemonsetSetting) (result *v1alpha1.ExtendedDaemonsetSetting, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(extendeddaemonsetsettingsResource, c.ns, extendedDaemonsetSetting), &v1alpha1.ExtendedDaemonsetSetting{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ExtendedDaemonsetSetting), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeExtendedDaemonsetSettings) UpdateStatus(extendedDaemonsetSetting *v1alpha1.ExtendedDaemonsetSetting) (*v1alpha1.ExtendedDaemonsetSetting, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(extendeddaemonsetsettingsResource, "status", c.ns, extendedDaemonsetSetting), &v1alpha1.ExtendedDaemonsetSetting{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ExtendedDaemonsetSetting), err
}

// Delete takes name of the extendedDaemonsetSetting and deletes it. Returns an error if one occurs.
func (c *FakeExtendedDaemonsetSettings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(extendeddaemonsetsettingsResource, c.ns, name), &v1alpha1.ExtendedDaemonsetSetting{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeExtendedDaemonsetSettings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(extendeddaemonsetsettingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ExtendedDaemonsetSettingList{})
	return err
}

// Patch applies the patch and returns the patched extendedDaemonsetSetting.
func (c *FakeExtendedDaemonsetSettings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExtendedDaemonsetSetting, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(extendeddaemonsetsettingsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ExtendedDaemonsetSetting{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ExtendedDaemonsetSetting), err
}
