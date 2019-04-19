package logtail

import (
	"bufio"
	"fmt"
	"errors"
	"io"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strings"
	"sync"
	"time"
)

type LogTimestamp string

type LogLine struct {
	Time LogTimestamp
	PodId string
	Content string
}


type LogLines []*LogLine

func (self LogLines) Len() int {
	return len(self)
}

func (self LogLines) Less(i, j int) bool {
	return self[i].Time < self[j].Time
}

func (self LogLines) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func ToLogLine(podId string, logStrs []string, startsWithDate bool) LogLines {
	logLines := LogLines{}
	for _, line := range logStrs {
		if line != ""{
			startsWithDate = startsWithDate &&  ('0' <= line[0] && line[0] <= '9')

			idx := strings.Index(line, " ")
			if idx > 0 && startsWithDate {
				timestamp := LogTimestamp(line[0:idx])
				content := line[idx+1:]
				logLines = append(logLines, &LogLine{Time: timestamp, PodId:podId, Content: content})
			} else {
				logLines = append(logLines, &LogLine{Time: LogTimestamp("0"), Content: line})
			}

		}
	}
	return logLines
}



type PodContainerList struct {
	Containers []string `json:"containers"`
}

func GetPodContainers(client kubernetes.Interface, namespace , podId string) (*PodContainerList, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(podId, v12.GetOptions{})

	if err != nil {
		return nil, err
	}

	containers := &PodContainerList{Containers: make([]string, 0)}

	for _, container := range pod.Spec.Containers {
		containers.Containers = append(containers.Containers, container.Name)
	}

	return containers, nil
}

func LabelSelectorToString(selector labels.Selector) (string, error) {

	selectStr := []string{}

	requirements, selectable := selector.Requirements()
	if !selectable{
		return "", errors.New("Selector cannot selectable")
	}

	for _, r:= range requirements{
		selectStr = append(selectStr, r.String())
	}

	if len(selectStr) > 0 {
		return strings.Join(selectStr, ","),nil
	}
	return "",nil

}


func GetDeploymentLogs(client kubernetes.Interface, namespace , deploymentId string, opt LogTailOption) (depLogs LogLines, err error){

	dep, err := client.AppsV1().Deployments(namespace).Get(deploymentId, v12.GetOptions{})

	if err != nil {
		return nil, err
	}

	selector := labels.SelectorFromSet(dep.Spec.Selector.MatchLabels)
	labelSelector, err := LabelSelectorToString(selector)
	if err != nil{
		return nil,err
	}

	pods, err := client.CoreV1().Pods(namespace).List(v12.ListOptions{LabelSelector:labelSelector})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Get Pod %v \n", pods)

	wg := new(sync.WaitGroup)
	logChan := make(chan *LogLines)
	finishChan := make(chan int)


	for _,pod := range pods.Items{
		wg.Add(1)
		go CollectPodLogs(client, namespace, pod.Name, opt, wg, logChan)
	}
	go func() {
		wg.Wait()
		finishChan<-1
	}()

	for {
		select {
			case podLogs := <- logChan:
				depLogs = append(depLogs, *podLogs...)
			case <- finishChan:
				fmt.Println("Quit")
				sort.Sort(depLogs)
				return depLogs, nil
		}
	}

}


func CollectPodLogs(client kubernetes.Interface, namespace, podId string, opt LogTailOption, wg *sync.WaitGroup,logChan chan *LogLines) {

	defer wg.Done()

	podLogs, err := GetPodLogs(client, namespace, podId, opt)

	if err != nil{
		errLogs := LogLines{
			&LogLine{
				Time: LogTimestamp(time.Now().Format(time.RFC3339Nano)),
				Content: err.Error(),
			},
		}
		logChan <- &errLogs
		return
	}
	logChan <- &podLogs
}

func GetPodLogs(client kubernetes.Interface, namespace, podId string, opt LogTailOption) (logLines LogLines, err error) {

	containers, err := GetPodContainers(client, namespace, podId)
	if err != nil {
		return nil, err
	}

	container := containers.Containers[0]

	logOption := &v13.PodLogOptions{
		Container: container,
		Follow: false,
		Previous: false,
		Timestamps: true,
	}

	if opt.LimitBytes > 0{
		logOption.LimitBytes = &opt.LimitBytes
	}else{
		logOption.LimitBytes = &LimitBytes

	}


	if !opt.FromTime.IsZero() {
		since := v12.Time{
			Time: opt.FromTime.Add(1*time.Nanosecond),
		}
		logOption.SinceTime = &since
	}else{
		if opt.TailLines <= 0{
			opt.TailLines = 200
		}
		logOption.TailLines = &opt.TailLines
	}

	request := client.CoreV1().Pods(namespace).GetLogs(podId, logOption)

	readCloser, err := request.Stream()
	if err != nil {
		return
	}
	defer readCloser.Close()

	reader := bufio.NewReader(readCloser)

	logs := []string{}

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF{
			break
		}
		logs = append(logs, line)
	}

	logLines = ToLogLine(podId, logs, logOption.Timestamps)

	return logLines, nil

}

