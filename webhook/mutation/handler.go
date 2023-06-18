package mutation

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	admissionv1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"strings"
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
	admissionReview.Response = admissionResponse
	admissionResponse.UID = admissionReview.Request.UID
	admissionResponse.Allowed = true

	if value, ok := pod.GetAnnotations()["simple-webhook/injection-enabled"]; ok && value == "true" {
		logrus.Print("Injection is enabled for pod")

		configVolumeConfig := os.Getenv("CONFIG_VOLUME_CONFIG")
		dataVolumeConfig := os.Getenv("DATA_VOLUME_CONFIG")
		initContainerConfig := os.Getenv("INIT_CONTAINER_CONFIG")
		configVolumeMountConfig := os.Getenv("CONFIG_VOLUME_MOUNT_CONFIG")
		dataVolumeMountConfig := os.Getenv("DATA_VOLUME_MOUNT_CONFIG")

		var patches []map[string]interface{}

		logrus.Print("Applying volumes patch ...")
		var volumes []string
		volumes = append(volumes, configVolumeConfig)
		volumes = append(volumes, dataVolumeConfig)
		volumesPatch := appendToArray(len(pod.Spec.Volumes) == 0, "/spec/volumes", volumes...)
		patches = append(patches, volumesPatch...)

		logrus.Print("Applying init containers patch ...")
		var initContainers []string
		initContainers = append(initContainers, initContainerConfig)
		initContainersPatch := appendToArray(len(pod.Spec.InitContainers) == 0, "/spec/initContainers", initContainers...)
		patches = append(patches, initContainersPatch...)

		logrus.Print("Applying volume mounts patch ...")
		var containers []string
		for _, currContainer := range pod.Spec.Containers {
			var volumeMount corev1.VolumeMount
			err := json.Unmarshal([]byte(configVolumeMountConfig), &volumeMount)
			if err != nil {
				msg := fmt.Sprintf("Error unmarshalling config volume mount: %v", err)
				logError(w, msg, 500)
				return
			}
			currContainer.VolumeMounts = append(currContainer.VolumeMounts, volumeMount)
			err = json.Unmarshal([]byte(dataVolumeMountConfig), &volumeMount)
			if err != nil {
				msg := fmt.Sprintf("Error unmarshalling data volume mount: %v", err)
				logError(w, msg, 500)
				return
			}
			currContainer.VolumeMounts = append(currContainer.VolumeMounts, volumeMount)
			currContainerBytes, err := json.Marshal(currContainer)
			if err != nil {
				msg := fmt.Sprintf("Error marshalling container: %v", err)
				logError(w, msg, 500)
				return
			}
			containers = append(containers, string(currContainerBytes))
		}
		containersPatch := appendToArray(true, "/spec/containers", containers...)
		patches = append(patches, containersPatch...)

		patchesBytes, err := json.Marshal(patches)
		if err != nil {
			msg := fmt.Sprintf("Error marshalling patches json: %v", err)
			logError(w, msg, 500)
			return
		}

		msg = fmt.Sprintf("Patches are: %v", string(patchesBytes))
		logrus.Print(msg)

		logrus.Print("All patches applied")

		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Patch = patchesBytes
	} else {
		logrus.Print("Skipping injection because it is not enabled for pod")
	}
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

func appendToArray(isEmpty bool, path string, jsonValues ...string) []map[string]interface{} {
	lastIndexSuffix := "/-"
	var patches []map[string]interface{}
	var jsonResult string
	if isEmpty {
		jsonResult = "[" + strings.Join(jsonValues, ",") + "]"
		bytes := json.RawMessage(jsonResult)
		patch := createPatch(path, bytes)
		patches = append(patches, patch)
		return patches
	}
	for _, jsonValue := range jsonValues {
		bytes := json.RawMessage(jsonValue)
		patch := createPatch(path+lastIndexSuffix, bytes)
		patches = append(patches, patch)
	}
	return patches
}

func createPatch(path string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":    "add",
		"path":  path,
		"value": value,
	}
}

func logError(w http.ResponseWriter, msg string, code int) {
	logrus.Print(msg)
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
