package logtail

import (
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestPodLog(t *testing.T) {


	client := fake.NewSimpleClientset()

	namespace := "test"

	testLogTail := NewLogTail().PodLogs(client, namespace)




}