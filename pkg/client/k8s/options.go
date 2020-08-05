package k8s

type K8sOptions struct {
	KubeConfig string  `json:"kubeconfig" yaml:"kubeconfig"`
	QPS        float32 `json:"qps,omitemtpy" yaml:"qps"`
}

func NewK8sOptions() *K8sOptions {
	return &K8sOptions{
		KubeConfig: "",
		QPS:        1e6,
	}
}
