package mutation

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

var (
	codecs = serializer.NewCodecFactory(runtime.NewScheme())
)

func MutatePod(w http.ResponseWriter, r *http.Request) {
	logrus.Print("Started mutation flow")
	deserializer := codecs.UniversalDeserializer()
	admissionReview, err := admissionReviewFromRequest(r, deserializer)
	msg := fmt.Sprintf("Request UID: %v", admissionReview.Request.UID)
	logrus.Print(msg)
	if err != nil {
		msg := fmt.Sprintf("Error getting admission review from request: %v", err)
		logrus.Print(msg)
		w.WriteHeader(400)
		w.Write([]byte(msg))
		return
	}
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if admissionReview.Request.Resource != podResource {
		msg := fmt.Sprintf("Did not receive pod, got %s", admissionReview.Request.Resource.Resource)
		logrus.Print(msg)
		w.WriteHeader(400)
		w.Write([]byte(msg))
		return
	}
	rawRequest := admissionReview.Request.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := deserializer.Decode(rawRequest, nil, &pod); err != nil {
		msg := fmt.Sprintf("Error decoding raw pod: %v", err)
		logrus.Print(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}
	admissionResponse := &admissionv1.AdmissionResponse{}
	var patch string
	patchType := v1.PatchTypeJSONPatch
	if _, ok := pod.Labels["hello"]; !ok {
		patch = `[{"op":"add","path":"/metadata/labels","value":{"hello":"fckingworld"}}]`
		//patch = `[{"op":"add","path":"/metadata/labels/hello","value":"fckingworld"}]`
	}
	admissionResponse.Allowed = true
	admissionResponse.UID = admissionReview.Request.UID
	if patch != "" {
		admissionResponse.PatchType = &patchType
		admissionResponse.Patch = []byte(patch)
	}
	admissionReview.Response = admissionResponse
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling response json: %v", err)
		logrus.Print(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		msg := fmt.Sprintf("Error writing response json: %v", err)
		logrus.Print(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}
	logrus.Print("Completed mutation flow")
}

func admissionReviewFromRequest(r *http.Request, deserializer runtime.Decoder) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Expected application/json content-type")
	}
	var body []byte
	if r.Body != nil {
		requestData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = requestData
	}
	admissionReview := &admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, admissionReview); err != nil {
		return nil, err
	}
	return admissionReview, nil
}
