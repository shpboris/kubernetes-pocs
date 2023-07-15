package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MapDataSpec struct {
	Mapname string `json:"mapname,omitempty"`
	Key1    string `json:"key1,omitempty"`
	Key2    string `json:"key2,omitempty"`
}

type MapDataStatus struct {
}

type MapData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MapDataSpec   `json:"spec,omitempty"`
	Status MapDataStatus `json:"status,omitempty"`
}

type MapDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MapData `json:"items"`
}
