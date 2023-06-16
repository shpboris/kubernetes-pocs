package mutation

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func MutatePod(w http.ResponseWriter, r *http.Request) {
	logrus.Print("Started mutation flow!")
	admissionReview, err := admissionReviewFromRequest(r)
	if err != nil {
		msg := fmt.Sprintf("Error getting admission review from request: %v", err)
		logError(w, msg, 400)
		return
	}
	msg := fmt.Sprintf("Request UID: %v", admissionReview.Request.UID)
	logrus.Print(msg)
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if admissionReview.Request.Resource != podResource {
		msg := fmt.Sprintf("Did not receive pod, got %s", admissionReview.Request.Resource.Resource)
		logError(w, msg, 400)
		return
	}
	rawRequest := admissionReview.Request.Object.Raw
	pod := corev1.Pod{}
	if err := json.Unmarshal(rawRequest, &pod); err != nil {
		msg := fmt.Sprintf("Error decoding raw pod: %v", err)
		logError(w, msg, 500)
		return
	}
	admissionResponse := &admissionv1.AdmissionResponse{}
	var patch string
	patchType := v1.PatchTypeJSONPatch
	if _, ok := pod.Labels["hello"]; !ok {
		patch = `[{"op":"add","path":"/metadata/labels/hello","value":"world1"}]`
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
		logError(w, msg, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		msg := fmt.Sprintf("Error writing response json: %v", err)
		logError(w, msg, 500)
		return
	}
	logrus.Print("Completed mutation flow!")
}

func admissionReviewFromRequest(r *http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("expected application/json content-type")
	}
	var body []byte
	if r.Body != nil {
		requestData, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = requestData
	}
	admissionReview := &admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed on %w", err)
	}
	return admissionReview, nil
}

func logError(w http.ResponseWriter, msg string, code int) {
	logrus.Print(msg)
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
