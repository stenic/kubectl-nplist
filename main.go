package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-nplist POD",
	Short: "NPlist help you debug network policies",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, _ := cmd.Flags().GetString("namespace")
		kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
		return runPod(kubeconfig, namespace, args[0])
	},
}

func init() {
	rootCmd.Flags().StringP("namespace", "n", "", "Namespace to look for the pod")
	if home := homedir.HomeDir(); home != "" {
		rootCmd.PersistentFlags().String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPod(kubeconfig string, namespace string, podName string) error {
	clientset := getKubeClient(kubeconfig)

	if namespace == "" {
		clientcmd.NewDefaultClientConfigLoadingRules().Load()
		clientCfg, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		namespace = clientCfg.Contexts[clientCfg.CurrentContext].Namespace
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Inspecting %s/%s\n", pod.Namespace, pod.Name)

	result, err := clientset.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var ingressPolicies []networkingv1.NetworkPolicy
	var egressPolicies []networkingv1.NetworkPolicy
	for _, item := range result.Items {
		if isPodMatch(pod, item.Spec.PodSelector) {
			if hasPolicy(item.Spec.PolicyTypes, networkingv1.PolicyTypeIngress) {
				ingressPolicies = append(ingressPolicies, item)
			}
			if hasPolicy(item.Spec.PolicyTypes, networkingv1.PolicyTypeEgress) {
				egressPolicies = append(egressPolicies, item)
			}
		}
	}

	fmt.Printf("%s\n", render(ingressPolicies, egressPolicies))
	return nil
}

func render(ingressPolicies []networkingv1.NetworkPolicy, egressPolicies []networkingv1.NetworkPolicy) string {
	t := table.NewWriter()
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})

	t.AppendSeparator()
	t.AppendRow(table.Row{"Ingress", "NetPol", "Action", "Port", "Source"})
	t.AppendSeparator()
	for _, pol := range ingressPolicies {
		row := table.Row{"Ingress", pol.Name}
		if pol.Spec.Ingress == nil {
			t.AppendRow(append(row, "DENY", "*", "*"))
		} else {
			for _, ing := range pol.Spec.Ingress {
				t.AppendRow(append(row, "ALLOW", printPorts(ing.Ports), printPeer(ing.From)))
			}
		}
	}

	t.AppendSeparator()
	t.AppendRow(table.Row{"Egress", "NetPol", "Action", "Port", "Destination"})
	t.AppendSeparator()
	for _, pol := range egressPolicies {
		row := table.Row{"Egress", pol.Name}
		if pol.Spec.Egress == nil {
			t.AppendRow(append(row, "DENY", "*", "*"))
		} else {
			for _, ing := range pol.Spec.Egress {
				t.AppendRow(append(row, "DENY", printPorts(ing.Ports), printPeer(ing.To)))
			}
		}
	}

	return t.Render()
}

func printPorts(ports []networkingv1.NetworkPolicyPort) string {
	var out []string

	if len(ports) == 0 {
		return "*"
	}
	for _, port := range ports {
		var line string
		if port.Protocol != nil {
			line += string(*port.Protocol)
		}
		line += "("
		line += port.Port.String()
		if port.EndPort != nil {
			line += "-" + strconv.FormatInt(int64(*port.EndPort), 10)
		}
		line += ")"
		out = append(out, line)
	}

	return strings.Join(out, " ")
}

func printPeer(peers []networkingv1.NetworkPolicyPeer) string {
	var out []string
	for _, peer := range peers {
		if peer.PodSelector != nil {
			for k, v := range peer.PodSelector.MatchLabels {
				out = append(out, fmt.Sprintf("podLabel(%s=%s)", k, v))
			}
		}
		if peer.NamespaceSelector != nil {
			for k, v := range peer.NamespaceSelector.MatchLabels {
				out = append(out, fmt.Sprintf("nsLabel(%s=%s)", k, v))
			}
		}
		if peer.IPBlock != nil {
			out = append(out, fmt.Sprintf("cidr(%s -%s)", peer.IPBlock.CIDR, strings.Join(peer.IPBlock.Except, ",")))
		}
	}

	return strings.Join(out, " ")
}

func isPodMatch(pod *v1.Pod, labelSelector metav1.LabelSelector) bool {
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		panic(err.Error())
	}
	return selector.Empty() || selector.Matches(apilabels.Set(pod.Labels))
}

func hasPolicy(policies []networkingv1.PolicyType, item networkingv1.PolicyType) bool {
	for _, pol := range policies {
		if pol == item {
			return true
		}
	}

	return false
}

func getKubeClient(kubeconfig string) *kubernetes.Clientset {

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

	return clientset
}
