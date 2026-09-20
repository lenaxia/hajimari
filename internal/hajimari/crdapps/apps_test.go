package crdapps

import (
	"testing"

	"github.com/toboshii/hajimari/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

var listKinds = map[schema.GroupVersionResource]string{
	{Group: "hajimari.io", Version: "v1alpha1", Resource: "applications"}: "ApplicationList",
}

func newFakeClient(objects ...*unstructured.Unstructured) *dynamicfake.FakeDynamicClient {
	runtimeObjects := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		runtimeObjects[i] = obj
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme.Scheme, listKinds, runtimeObjects...)
}

func TestPopulateWithApplications(t *testing.T) {
	app := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "hajimari.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "proxmox",
				"namespace": "networking",
			},
			"spec": map[string]interface{}{
				"name":  "Proxmox",
				"group": "Infrastructure",
				"url":   "https://proxmox.thekao.cloud",
			},
		},
	}

	dynClient := newFakeClient(app)
	appConfig := config.Config{}

	list := NewList(dynClient, appConfig)
	items, err := list.Populate("").Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected app groups, got none")
	}

	found := false
	for _, group := range items {
		if group.Group == "Infrastructure" {
			for _, a := range group.Apps {
				if a.Name == "Proxmox" && len(a.VisibleGroups) == 0 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected Proxmox app without visible groups, got %+v", items)
	}
}

func TestPopulateWithVisibleGroups(t *testing.T) {
	app := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "hajimari.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"name":          "Test",
				"group":         "Misc",
				"url":           "https://test.thekao.cloud",
				"visibleGroups": []interface{}{"admins", "family"},
			},
		},
	}

	dynClient := newFakeClient(app)
	list := NewList(dynClient, config.Config{})
	items, err := list.Populate("").Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || len(items[0].Apps) != 1 {
		t.Fatalf("unexpected items: %+v", items)
	}
	vg := items[0].Apps[0].VisibleGroups
	if len(vg) != 2 || vg[0] != "admins" || vg[1] != "family" {
		t.Fatalf("visible groups = %v, want [admins family]", vg)
	}
}

func TestPopulateEmpty(t *testing.T) {
	dynClient := newFakeClient()
	list := NewList(dynClient, config.Config{})
	items, err := list.Populate("").Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %+v", items)
	}
}
