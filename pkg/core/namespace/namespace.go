package namespace

type Namespace struct {
	Version     string            `json:"version,omitempty"`
	Cluster     string            `json:"cluster,omitempty"`
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	Finalizers []FinalizerName `json:"finalizers,omitempty"`

	Phase NamespacePhase `json:"phase,omitempty"`
}

type NamespacePhase int

const (
	NamespaceActive      NamespacePhase = iota
	NamespaceTerminating
)

type FinalizerName int

const (
	FinalizerKubernetes FinalizerName = iota
)