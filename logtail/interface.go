package logtail

import (
	"k8s.io/client-go/kubernetes"

	"time"
)

var LimitBytes int64 = 5000000

const (
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z"
)

type LogTailOption struct {
	FromTime time.Time
	TailLines int64
	LimitBytes int64
	LimitLines int64
}


type LogTail interface {
	PodLogs(client kubernetes.Interface, namespace string) PodLogTail
	DeploymentLogs(client kubernetes.Interface, namespace string) DeploymentLogTail
}

type PodLogTail interface {
	DateLogs(podId string, opt LogTailOption) (LogLines, error)
}

type DeploymentLogTail interface {
	DateLogs(deploymentId string, opt LogTailOption) (LogLines, error)
}


type logTail struct {
}

func (self *logTail) PodLogs(client kubernetes.Interface, namespace string) PodLogTail {
	return &podLogs{
		client: client,
		namespace: namespace,
	}
}

func (self *logTail) DeploymentLogs(client kubernetes.Interface, namespace string) DeploymentLogTail {
	return &deploymentLogs{
		client: client,
		namespace: namespace,
	}
}

func NewLogTail() *logTail {
	return &logTail{}
}

type podLogs struct {
	client kubernetes.Interface
	namespace string
}

func (self *podLogs) DateLogs(podId string, opt LogTailOption) (LogLines, error) {
	logs, err := GetPodLogs(self.client, self.namespace, podId, opt)
	return logs, err
}

type deploymentLogs struct {
	client kubernetes.Interface
	namespace string
}

func (self *deploymentLogs) DateLogs(depId string, opt LogTailOption) (LogLines, error) {
	logs, err := GetDeploymentLogs(self.client, self.namespace, depId, opt)
	return logs, err
}






