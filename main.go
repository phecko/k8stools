package main

import (
	"fmt"
	"github.com/phecko/k8stools/logtail"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"

	//"log"
)

func main() {
	var kubeconfig string


	kubeconfig = "kubeconfig/local"

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}


	namespace := "default"

	podId := "kubernetes-dashboard-57df4db6b-nm9dh"

	deploymentId := "zhijian-server-ucenter"

	t := logtail.NewLogTail()

	pTail := t.PodLogs(clientset, namespace)

	logOption := logtail.LogTailOption{
		FromTime: time.Now().Add(-20*time.Second),
	}

	logs, err := pTail.DateLogs(podId, logOption)
	if err != nil{
		fmt.Printf("Get Pod Error: %s \n", err.Error())
	}else{
		fmt.Printf("Get Pod Logs: %d \n", len(logs))
	}

	dTail := t.DeploymentLogs(clientset, namespace)
	logs, err = dTail.DateLogs(deploymentId, logOption)
	if err != nil{
		fmt.Printf("Get Deployment Error: %s \n", err.Error())
	}else{
		fmt.Printf("Get Deployment Logs: %d \n", len(logs))
	}

	for _, depLog := range logs{
		fmt.Println(depLog.Time, depLog.Content)
	}
	//fmt.Println(len(deplogs))

}




