package wrappers

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetVisibleGroups(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		want       []string
	}{
		{"absent annotation", "", []string{}},
		{"single group", "admins", []string{"admins"}},
		{"comma separated with padding", " Family , friends ", []string{"family", "friends"}},
		{"empty value", " ", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ingress := &v1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"hajimari.io/visible-groups": tt.annotation,
					},
				},
			}
			got := NewIngressWrapper(ingress).GetVisibleGroups()
			if len(tt.want) == 0 && len(got) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVisibleGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
