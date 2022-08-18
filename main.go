package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var kubeconfig string
var ctx string
var debug bool

func main() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&ctx, "context", "", "The context to use from the kubeconfig file")
	flag.BoolVar(&debug, "debug", false, "Enable some debug logging")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // Push signals into channel
	go func() {
		<-sigs
		log.Println("Caught signal...")
		close(stop)
	}()

	client, err := NewClient(kubeconfig, ctx)
	if err != nil {
		log.Fatal(err)
	}

	client.CoreV1().Endpoints("").List(context.TODO(), api_v1.ListOptions{})

	factory := informers.NewSharedInformerFactory(client, 30*time.Minute)

	endpointsInformer := factory.Core().V1().Endpoints().Informer()

	endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: endpointUpdate,
	})

	factory.Start(stop)
	factory.WaitForCacheSync(stop)
	log.Println("Caches synced. Ready to work")
	<-stop
	log.Println("Shutting down")

}

type endpoint struct {
	protocol v1.Protocol
	ip       string
	port     int32
}

func endpointUpdate(old, current interface{}) {
	oldEndpoints := old.(*v1.Endpoints)
	currentEndpoints := current.(*v1.Endpoints)

	oldActiveUDPEndpoints := getActiveUDPEndpoints(oldEndpoints)
	currentActiveUDPEndpoints := getActiveUDPEndpoints(currentEndpoints)
	if debug {
		log.Println("-----------------")
		log.Println("old udp endpoints: ", oldActiveUDPEndpoints)
		log.Println("new udp endpoints: ", currentActiveUDPEndpoints)
	}
	//Try to find all old endpoints in the new set
OUTER:
	for _, oldEndpoint := range oldActiveUDPEndpoints {
		for _, e := range currentActiveUDPEndpoints {
			if e == oldEndpoint {
				continue OUTER //endpoint found, proceed with next endpoint
			}
		}
		cleanupEndpoint(currentEndpoints.Namespace, currentEndpoints.Name, oldEndpoint)
	}

}

func getActiveUDPEndpoints(ep *v1.Endpoints) []endpoint {
	var activeEndpoints []endpoint
	for _, subset := range ep.Subsets {
		for _, port := range subset.Ports {
			if port.Protocol == v1.ProtocolUDP {
				for _, address := range subset.Addresses {
					activeEndpoints = append(activeEndpoints, endpoint{port.Protocol, address.IP, port.Port})
				}
			}
		}
	}
	return activeEndpoints
}

// NoConnectionToDelete is the error string returned by conntrack when no matching connections are found:w
const NoConnectionToDelete = "0 flow entries have been deleted"

var flowEntriesRE = regexp.MustCompile(`(\d+) flow entries have been deleted`)

func cleanupEndpoint(namespace string, name string, ep endpoint) {
	log.Printf("Purging conntrack table for endpoint %v (%s/%s)", ep, namespace, name)

	cmd := exec.Command(
		"/usr/sbin/conntrack",
		"-D",
		"-p", strings.ToLower(string(ep.protocol)),
		"--reply-src", ep.ip,
		"--reply-port-src", strconv.Itoa(int(ep.port)),
	)
	stderr := new(strings.Builder)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil && !strings.Contains(stderr.String(), NoConnectionToDelete) {
		log.Printf("Failed to run %s: %s stderr: %s", strings.Join(cmd.Args, " "), err, stderr.String())
		return
	}

	if matches := flowEntriesRE.FindStringSubmatch(stderr.String()); matches != nil {
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Println("Failed to extract number of deleted entries: ", err)
			return
		}
		log.Printf("Deleted %d entries", n)
	}

}
